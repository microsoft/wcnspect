package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/microsoft/winspect/pkg/comprise"
	pb "github.com/microsoft/winspect/rpc"

	flag "github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var validCommands = []string{"capture", "hns", "help"}

const namespace string = "default"
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

type client struct {
	pb.CaptureServiceClient
	pb.HCNServiceClient
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
	var kubeconfig *string
	var nodes, ips, protocols, ports, macs string
	var time int32

	// Flags
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Int32VarP(&time, "time", "d", 0, "Time to run packet capture for (in seconds). Runs indefinitely given 0.")
	flag.StringVarP(&nodes, "nodes", "n", "", "Specify which nodes winspect should send requests to using node names. Runs on all windows nodes by default.")
	flag.StringVarP(&ips, "ips", "i", "", "Match source or destination IP address. CIDR supported.")
	flag.StringVarP(&protocols, "protocols", "t", "", "Match by transport protocol (TCP, UDP, ICMP).")
	flag.StringVarP(&ports, "ports", "p", "", "Match source or destination port number.")
	flag.StringVarP(&macs, "macs", "m", "", "Match source or destination MAC address.")
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// FIXME: Some temporary hardcoded error handling for input
	if len(os.Args) <= 1 {
		fmt.Println(winspectHelpString)
		return
	}

	cmd := os.Args[1]
	subcmd := ""
	if len(os.Args) >= 3 {
		subcmd = os.Args[2]
	}

	if !comprise.Contains(validCommands, cmd) {
		fmt.Printf("Unknown command '%s'. See winspect help.\n", cmd)
		return
	}

	if cmd == "help" {
		fmt.Println(winspectHelpString)
		return
	}

	if cmd == "hns" && len(os.Args) <= 1 && !comprise.Contains(comprise.Keys(pb.HCNType_value), os.Args[2]) {
		fmt.Printf("Unknown command '%s'. See winspect hcn help.\n", os.Args[2])
		return
	}

	if subcmd == "help" {
		switch cmd {
		case "capture":
			fmt.Println(captureHelpString)
		case "hns":
			fmt.Println(hnsHelpString)
		}
		return
	}

	// Pull nodes and pods
	nodeset, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// NEED: list of ips from nodes that have a windows os; map to translate between node name -> node ip;
	//			map of pod name -> pod ip; map of pod name -> node ip

	// Create params struct
	args := params{
		cmd:       cmd,
		subcmd:    subcmd,
		nodes:     parseValidateNodes(nodes, nodeset.Items),
		ips:       parseValidateIPAddrs(ips),
		protocols: parseValidateProts(protocols),
		ports:     parseValidatePorts(ports),
		macs:      parseValidateMACAddrs(macs),
		time:      validateTime(time),
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

	c1 := pb.NewCaptureServiceClient(cc)
	c2 := pb.NewHCNServiceClient(cc)
	c := &client{c1, c2}

	switch args.cmd {
	case "capture":
		runCaptureStream(c, args, ip)
	case "hns":
		printHCNLogs(c, args, ip)
	case "stop":
		runStopCapture(c, ip)
	}
}

func runCaptureStream(c pb.CaptureServiceClient, args *params, ip string) {
	// Capture any sigint to send a StopCapture request

	fmt.Printf("Starting to do a Server Streaming RPC (from IP: %s)...\n", ip)

	// Create request object
	req := &pb.CaptureRequest{
		Duration:  args.time,
		Timestamp: timestamppb.Now(),
		Filter: &pb.Filters{
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

	fmt.Printf("Finished receiving stream from IP: %s.\n", ip)
}

func runStopCapture(c pb.CaptureServiceClient, ip string) {
	_, err := c.StopCapture(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("error while calling StopCapture RPC (from IP: %s): %v", ip, err)
	}

	fmt.Printf("Ended packet capture on IP: %s.\n", ip)
}

func printHCNLogs(c pb.HCNServiceClient, args *params, ip string) {
	fmt.Printf("Requesting HCN logs (from IP: %s)...\n", ip)
	hcntype := args.subcmd

	// Create request object
	req := &pb.HCNRequest{
		Hcntype: pb.HCNType(pb.HCNType_value[hcntype]),
	}

	// Send request
	res, err := c.GetHCNLogs(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GetHCNLogs RPC (from IP: %s): %v", ip, err)
	}

	fmt.Printf("Received logs for %s (from IP: %s): \n%s\n", hcntype, ip, string(res.GetHcnResult()))
}
