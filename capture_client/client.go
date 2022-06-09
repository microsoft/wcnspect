package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"winspect/capturespb"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	flag "github.com/spf13/pflag"
)

var VALID_HCN_LOGS string = "loadbalancers endpoints networks"

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

func main() {
	// User input variables
	var nodes, ips, protocols, ports, macs string //TODO: nodes required --can just check value then log fatal if invalid
	var time int32

	// Flags
	flag.Int32VarP(&time, "time", "d", 0, "Time to run packet capture for in seconds. Runs indefinitely given 0.")
	flag.StringVarP(&nodes, "nodes", "n", "", "Specify which nodes winspect should send requests to using node IPs. This field is required.")
	flag.StringVarP(&ips, "ips", "i", "", "Match source or destination IP address. CIDR supported.")
	flag.StringVarP(&protocols, "protocols", "t", "", "Match by transport protocol (TCP, UDP, ICMP).")
	flag.StringVarP(&ports, "ports", "p", "", "Match source or destination port number.")
	flag.StringVarP(&macs, "macs", "m", "", "Match source or destination MAC address.")
	flag.Parse()

	// Some error handling for input
	cmd := os.Args[1]
	if (cmd != "capture" && cmd != "hns") || (cmd == "hns" && !strings.Contains(VALID_HCN_LOGS, os.Args[2])) {
		fmt.Println("Invalid command.")
		return
	}

	if len(nodes) == 0 {
		fmt.Println("Must pass at least one ip to the nodes flag.")
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

	fmt.Printf("Stopped capture on IP: %s.", ip)
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
