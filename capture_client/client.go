package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"winspect/capturespb"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	flag "github.com/spf13/pflag"
)

var validCommands = []string{"capture", "hns", "help"}
var validHCNLogs = []string{"loadbalancers", "endpoints", "networks"}

const winspectHelpString string = `winspect <command> [OPTIONS | help]
    Advanced distributed packet capture and HNS log collection.

Commands
    capture    Start packet capture on given nodes and stream to client.
    hns        Retrieve HNS logs from given nodes.

    help       Show help text for specific command.
               Example: winspect capture help

    --help     Show help for available flags.

`
const captureHelpString string = `winspect capture <command>

Commands
    --help    Show help for available flags.
`
const hnsHelpString string = `winspect hns <command> [OPTIONS]

Commands
    loadbalancers    Retrieve logs for loadbalancers on each node.
    endpoints        Retrieve logs for endpoints on each node.
    networks         Retrieve logs for networks on each node.

`

type params struct {
	cmd       string
	subcmd    string
	nodes     []string
	ips       []string
	protocols []string
	ports     []string
	macs      []string
	time      int32
}

func contains(s []string, el string) bool {
	for _, value := range s {
		if value == el {
			return true
		}
	}
	return false
}

func cleanup(nodes []string) {
	var wg sync.WaitGroup
	for _, ip := range nodes {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Launch a goroutine to run the capture
		go createConnectionAndRoute(ip, &params{cmd: "stop"}, &wg)
	}

	// Wait for all captures to complete
	wg.Wait()
}

func main() {
	// User input variables
	var nodes, ips, protocols, ports, macs string //TODO: nodes required --can just check value then log fatal if invalid
	var time int32

	// Flags
	flag.Int32VarP(&time, "time", "d", 0, "Time to run packet capture for (in seconds). Runs indefinitely given 0.")
	flag.StringVarP(&nodes, "nodes", "n", "", "Specify which nodes winspect should send requests to using node IPs. This field is required.")
	flag.StringVarP(&ips, "ips", "i", "", "Match source or destination IP address. CIDR supported.")
	flag.StringVarP(&protocols, "protocols", "t", "", "Match by transport protocol (TCP, UDP, ICMP).")
	flag.StringVarP(&ports, "ports", "p", "", "Match source or destination port number.")
	flag.StringVarP(&macs, "macs", "m", "", "Match source or destination MAC address.")
	flag.Parse()

	// Some error handling for input
	if len(os.Args) <= 1 {
		fmt.Println(winspectHelpString)
		return
	}

	cmd := os.Args[1]
	if !contains(validCommands, cmd) {
		fmt.Printf("Unknown command '%s'. See winspect help.\n", cmd)
		return
	}

	if cmd == "help" {
		fmt.Println(winspectHelpString)
		return
	}

	if cmd == "hns" && len(os.Args) <= 1 && !contains(validHCNLogs, os.Args[2]) {
		fmt.Printf("Unknown command '%s'. See winspect hcn help.\n", os.Args[2])
		return
	}

	if len(os.Args) >= 3 && os.Args[2] == "help" {
		switch cmd {
		case "capture":
			fmt.Println(captureHelpString)
		case "hns":
			fmt.Println(hnsHelpString)
		}
		return
	}

	if len(nodes) == 0 {
		fmt.Println("Must pass at least one IP to the --nodes flag.")
		return
	}

	// Create params struct
	args := params{
		cmd:       cmd,
		subcmd:    os.Args[2],
		nodes:     strings.Split(nodes, ","),
		ips:       strings.Split(ips, ","),
		protocols: strings.Split(protocols, ","),
		ports:     strings.Split(ports, ","),
		macs:      strings.Split(macs, ","),
		time:      time,
	}

	// Capture any sigint to send a StopCapture request
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(args.nodes)
		os.Exit(1)
	}()

	// Create waitgroup to maintain each connection
	var wg sync.WaitGroup
	for _, ip := range args.nodes {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Launch a goroutine to run the capture
		go createConnectionAndRoute(ip, &args, &wg)
	}

	// Wait for all captures to complete
	wg.Wait()
}

func createConnectionAndRoute(ip string, args *params, wg *sync.WaitGroup) {
	// Decrement relevant waitgroup counter when goroutine completes
	defer wg.Done()

	cc, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := capturespb.NewCaptureServiceClient(cc)

	switch args.cmd {
	case "capture":
		runCaptureStream(c, args, ip)
	case "hns":
		printHCNLogs(c, args, ip)
	case "stop":
		runStopCapture(c, ip)
	}
}

func runCaptureStream(c capturespb.CaptureServiceClient, args *params, ip string) {
	fmt.Printf("Starting to do a Server Streaming RPC (from IP: %s)...\n", ip)

	// Create request object
	req := &capturespb.CaptureRequest{
		Duration:  args.time,
		Timestamp: timestamppb.Now(),
		Filter: &capturespb.Filters{
			Ips:       args.ips,
			Protocols: args.protocols,
			Ports:     args.ports,
			Macs:      args.macs,
		},
	}

	// Send request
	resStream, err := c.StartCapture(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling StartCapture RPC (from IP: %s): %v", ip, err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// we've reached the end of the stream
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream (from IP: %s): %v", ip, err)
		}

		fmt.Printf("Response from StartCapture (%s) sent at %s: \n%v\n", ip, msg.GetTimestamp().AsTime(), msg.GetResult())
	}
}

func runStopCapture(c capturespb.CaptureServiceClient, ip string) {
	_, err := c.StopCapture(context.Background(), &capturespb.Empty{})
	if err != nil {
		log.Fatalf("error while calling StopCapture RPC (from IP: %s): %v", ip, err)
	}

	fmt.Printf("Ended packet capture on IP: %s.\n", ip)
}

func printHCNLogs(c capturespb.CaptureServiceClient, args *params, ip string) {
	fmt.Printf("Requesting HCN logs (from IP: %s)...\n", ip)
	hcntype := args.subcmd

	// Create request object
	req := &capturespb.HCNRequest{
		Type: hcntype,
	}

	// Send request
	res, err := c.GetHCNLogs(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GetHCNLogs RPC (from IP: %s): %v", ip, err)
	}

	fmt.Printf("Received logs for %s (from IP: %s): \n%s\n", hcntype, ip, string(res.GetHcnResult()))
}
