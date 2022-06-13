package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"winspect/capturespb"

	"github.com/Microsoft/hcsshim/hcn"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	flag "github.com/spf13/pflag"
)

var pktParams = map[string]string{
	"protocols": "-t",
	"ips":       "-i",
	"ports":     "-p",
	"macs":      "-m",
}

type server struct {
	capturespb.UnimplementedCaptureServiceServer
	currMonitor      *exec.Cmd          // Tracks the running pktmon stream
	pktContextCancel context.CancelFunc // Tracks the pktmon stream's context's cancel func
}

func resetPktmon(captures bool, filters bool) error {
	// Stop pktmon
	if captures {
		if err := exec.Command("cmd", "/c", "pktmon stop").Run(); err != nil {
			log.Fatalf("Failed to stop pktmon: %v", err)
		}
	}

	// Clear filters
	if filters {
		if err := exec.Command("cmd", "/c", "pktmon filter remove").Run(); err != nil {
			log.Fatalf("Failed to remove old filters: %v", err)
		}
	}

	return nil
}

func pktmonStream(stdout *io.ReadCloser) <-chan string {
	c := make(chan string)

	scanner := bufio.NewScanner(*stdout)
	scanner.Split(bufio.ScanLines)
	go func(s *bufio.Scanner) {
		s.Split(bufio.ScanLines)
		for s.Scan() {
			c <- s.Text()
		}
	}(scanner)

	return c
}

func (s *server) StartCapture(req *capturespb.CaptureRequest, stream capturespb.CaptureService_StartCaptureServer) error {
	fmt.Printf("StartCapture function was invoked with %v\n", req)

	// Retrieve and format request arguments
	dur := req.GetDuration()
	filter := req.GetFilter()
	args := map[string][]string{
		"protocols": filter.GetProtocols(),
		"ips":       filter.GetIps(),
		"ports":     filter.GetPorts(),
		"macs":      filter.GetMacs(),
	}

	// Add an empty protocol since our filtering mechanism depends on there being one
	if len(args["protocols"]) == 0 {
		args["protocols"] = append(args["protocols"], "")
	}

	// If duration is less than 0, we run for an "infinite" amount of time
	if dur <= 0 {
		dur = math.MaxInt32
	}

	// Ensure filters are reset and add new ones
	resetPktmon(true, true)
	for _, protocol := range args["protocols"] {
		name := " winspect" + protocol + " "
		filters := []string{}

		// Build filter slice
		for arg, addrs := range args {
			// Short circuiting conditional for adding protocol(s) if in filter request
			if len(addrs) > 0 && len(addrs[0]) > 0 {
				filters = append(filters, pktParams[arg]+" "+strings.Join(addrs, " "))
			}
		}

		// If there are no filters, break
		if len(filters) == 0 {
			break
		}

		// Execute filter command
		fmt.Println("Applying filters...")
		if err := exec.Command("cmd", "/c", "pktmon filter add"+name+strings.Join(filters, " ")).Run(); err != nil {
			log.Fatalf("Failed to add%sfilter: %v", name, err)
		}
	}

	// Create a timeout context and set as server's pktmon canceller
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dur)*time.Second)
	s.pktContextCancel = cancel
	defer cancel()

	// Execute pktmon command and check for errors, if successful, set as server's currMonitor
	exec.Command("cmd", "/c", "pktmon stop").Run()
	cmd := exec.CommandContext(ctx, "cmd", "/c", "pktmon start -c -m real-time")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	s.currMonitor = cmd

	// Create a channel to receive pktmon stream from
	c := pktmonStream(&stdout)

	// Goroutine with a timeout constraint and pulling on pktmon channel with scanning loop
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case out := <-c:
				res := &capturespb.CaptureResponse{
					Result:    out,
					Timestamp: timestamppb.Now(),
				}

				stream.Send(res)
				log.Printf("Sent: \n%v", res)
			case <-ctx.Done():
				log.Printf("Stream finished.")
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()

	// Reset pktmon filters, server's currMonitor, and server's pktContextCancel
	resetPktmon(false, true)
	s.currMonitor = nil
	s.pktContextCancel = nil
	log.Printf("Packet monitor filters reset.")

	return nil
}

func (s *server) StopCapture(ctx context.Context, req *capturespb.Empty) (*capturespb.Empty, error) {
	if s.currMonitor != nil {
		s.currMonitor.Process.Kill()
		s.pktContextCancel()
		log.Printf("Successfully killed packet capture stream.\n")
	}

	return req, nil
}

func (*server) GetHCNLogs(ctx context.Context, req *capturespb.HCNRequest) (*capturespb.HCNResponse, error) {
	hcntype := req.GetType()
	obj := []byte{}
	var lerr, jerr error

	switch hcntype {
	case "networks":
		var networks []hcn.HostComputeNetwork
		networks, lerr = hcn.ListNetworks()
		obj, jerr = json.Marshal(networks)
	case "endpoints":
		var endpoints []hcn.HostComputeEndpoint
		endpoints, lerr = hcn.ListEndpoints()
		obj, jerr = json.Marshal(endpoints)
	case "loadbalancers":
		var lbs []hcn.HostComputeLoadBalancer
		lbs, lerr = hcn.ListLoadBalancers()
		obj, jerr = json.Marshal(lbs)
	}

	if lerr != nil {
		log.Fatal(lerr)
	}

	if jerr != nil {
		log.Fatal(jerr)
	}

	res := &capturespb.HCNResponse{
		HcnResult: obj,
	}

	return res, nil
}

func main() {
	// User input variables
	var port string

	// Flags
	flag.StringVarP(&port, "port", "p", "50051", "Specify port for server to listen on.")
	flag.Parse()

	// Input validation
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Supplied value was not a valid port.")
	}

	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Print("Server started")
	s := grpc.NewServer()
	capturespb.RegisterCaptureServiceServer(s, &server{})

	// Register reflection service on gRPC server
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
