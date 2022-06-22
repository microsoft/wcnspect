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

	"github.com/microsoft/winspect/common"
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

const namespace string = "default"
const winspectHelpString string = `winspect <command> [OPTIONS | help]
    Advanced distributed packet capture and HNS log collection.

Commands
    capture    Start packet capture on given nodes and stream to client.
    hns        Retrieve HNS logs from given nodes.

    help       Show help text for specific command.
               Example: winspect capture help

`
const captureHelpString string = `winspect capture <command>

Commands
    --help    Show help for available flags.

`
const hnsHelpString string = `winspect hns <command> [OPTIONS]

Commands
    all                 Retrieve all hnsdiag logs on each node.
    endpoints        	Retrieve logs for endpoints on each node.
    loadbalancers    	Retrieve logs for loadbalancers on each node.
    namespaces          Retrieve logs for namespaces on each node.
    networks         	Retrieve logs for networks on each node.

Flags
    -n, --nodes string	Specify which nodes winspect should send requests to using node names. Runs on all windows nodes by default.
    -d, --json          Detailed option for logs.

`

type params struct {
	cmd       string
	subcmd    string
	hosts     map[string][]string
	ips       []string
	protocols []string
	ports     []string
	macs      []string
	time      int32
	json      bool
}

type client struct {
	pb.CaptureServiceClient
	pb.HCNServiceClient
}

var (
	// Shared flags
	kubeconfig *string
	nodes      string

	// Capture flags
	pods, ips, protocols, ports, macs string
	time                              int32

	// Hns flags
	json bool

	// Commands
	captureCmd = flag.NewFlagSet("capture", flag.ExitOnError)
	hnsCmd     = flag.NewFlagSet("hns", flag.ExitOnError)
	helpCmd    = flag.NewFlagSet("help", flag.ExitOnError)
)

var subcommands = map[string]*flag.FlagSet{
	"capture": captureCmd,
	"hns":     hnsCmd,
	"help":    helpCmd,
}

func setupCommonFlags() {
	for name, fs := range subcommands {
		if name == "help" {
			continue
		}

		if home := homedir.HomeDir(); home != "" {
			kubeconfig = fs.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = fs.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}

		fs.StringVarP(&nodes, "nodes", "n", "", "Specify which nodes winspect should send requests to using node names. Runs on all windows nodes by default.")
	}
}

func setupCaptureFlags() {
	captureCmd.StringVarP(&pods, "pods", "p", "", "Specify which pods the capture should filter on. Supports up to two pod names. Automatically defines nodes to capture on.")
	captureCmd.StringVarP(&ips, "ips", "i", "", "Match source or destination IP address. CIDR supported.")
	captureCmd.StringVarP(&protocols, "protocols", "t", "", "Match by transport protocol (TCP, UDP, ICMP).")
	captureCmd.StringVarP(&ports, "ports", "r", "", "Match source or destination port number.")
	captureCmd.StringVarP(&macs, "macs", "m", "", "Match source or destination MAC address.")
	captureCmd.Int32VarP(&time, "time", "d", 0, "Time to run packet capture for (in seconds). Runs indefinitely given 0.")
}

func setupHnsFlags() {
	hnsCmd.BoolVarP(&json, "json", "d", false, "Detailed option for logs.")
}

func main() {
	setupCommonFlags()
	setupCaptureFlags()
	setupHnsFlags()

	if len(os.Args) < 2 {
		vlog.Fatalf(winspectHelpString)
	}

	// CLI structure
	cmd := os.Args[1]
	subcmd := ""
	switch cmd {
	case "capture":
		if len(os.Args) > 2 {
			subcommands[cmd].Parse(os.Args[2:])

			if os.Args[2] == "help" {
				vlog.Fatalf(captureHelpString)
			}
		}
	case "hns":
		if len(os.Args) < 3 || os.Args[2] == "help" {
			vlog.Fatalf(hnsHelpString)
		}

		subcmd = os.Args[2]
		subcommands[cmd].Parse(os.Args[2:])

		if !comprise.Contains(comprise.Keys(pb.HCNType_value), subcmd) {
			vlog.Fatalf("Unknown command '%s'. See winspect hns help.", subcmd)
		}
	case "help":
		vlog.Fatalf(winspectHelpString)
	default:
		vlog.Fatalf("Unknown subcommand '%s', see winspect help for more details.", cmd)
	}

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

	// Pull nodes
	nodeset, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Pull pods
	podset, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Create params struct
	hosts := parseValidatePods(pods, podset.Items)

	args := params{
		cmd:       cmd,
		subcmd:    subcmd,
		ips:       parseValidateIPAddrs(ips),
		protocols: parseValidateProts(protocols),
		ports:     parseValidatePorts(ports),
		macs:      parseValidateMACAddrs(macs),
		time:      validateTime(time),
		json:      json,
	}

	if len(hosts) == 0 {
		hosts = parseValidateNodes(nodes, nodeset.Items)
	}
	args.hosts = hosts
	nodeIPs := comprise.Keys(args.hosts)

	// Capture any sigint to send a StopCapture request
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(nodeIPs)
		os.Exit(1)
	}()

	// Create waitgroup to maintain each connection
	var wg sync.WaitGroup
	for _, ip := range nodeIPs {
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

	//FIXME: hardcoded port addition
	cc, err := grpc.Dial(ip+":"+common.DefaultServerPort, grpc.WithInsecure())
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
	fmt.Printf("Starting to do a Server Streaming RPC (from IP: %s)...\n", ip)
	pods := args.hosts[ip]

	// Create request object
	req := &pb.CaptureRequest{
		Duration:  args.time,
		Timestamp: timestamppb.Now(),
		Filter: &pb.Filters{
			Pods:      pods,
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
		Json:    args.json,
	}

	// Send request
	res, err := c.GetHCNLogs(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GetHCNLogs RPC (from IP: %s): %v", ip, err)
	}

	fmt.Printf("Received logs for %s (from IP: %s):\n\n%s\n", hcntype, ip, string(res.GetHcnResult()))
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
