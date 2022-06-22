package main

import (
	"bufio"
	"context"
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

	"github.com/microsoft/winspect/common"
	"github.com/microsoft/winspect/pkg/nets"
	pb "github.com/microsoft/winspect/rpc"

	flag "github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var pktParams = map[string]string{
	"protocols": "-t",
	"ips":       "-i",
	"ports":     "-p",
	"macs":      "-m",
}

type captureServer struct {
	pb.UnimplementedCaptureServiceServer
	currMonitor      *exec.Cmd          // Tracks the running pktmon stream
	pktContextCancel context.CancelFunc // Tracks the pktmon stream's context's cancel func
}

type hcnServer struct {
	pb.UnimplementedHCNServiceServer
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

func (s *captureServer) StartCapture(req *pb.CaptureRequest, stream pb.CaptureService_StartCaptureServer) error {
	fmt.Printf("StartCapture function was invoked with %v\n", req)
	pktmonStartCommand := "pktmon start -c -m real-time"

	// Retrieve and format request arguments
	dur := req.GetDuration()
	filter := req.GetFilter()
	pods := filter.GetPods()
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

	// If we have pod IPs, then change the pktmonStartCommand
	if len(pods) > 0 {
		podIDs := nets.GetPodIDs(pods)
		pktmonStartCommand += fmt.Sprintf(" --comp %s", strings.Join(podIDs, " "))
	}

	// Create a timeout context and set as server's pktmon canceller
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dur)*time.Second)
	s.pktContextCancel = cancel

	// Execute pktmon command and check for errors, if successful, set as server's currMonitor
	exec.Command("cmd", "/c", "pktmon stop").Run()
	cmd := exec.CommandContext(ctx, "cmd", "/c", pktmonStartCommand)
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
				res := &pb.CaptureResponse{
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
	cancel()
	resetPktmon(false, true)
	s.currMonitor = nil
	s.pktContextCancel = nil
	log.Printf("Packet monitor filters reset.")

	return nil
}

func (s *captureServer) StopCapture(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	fmt.Println("StopCapture function was invoked.")

	if s.currMonitor != nil {
		s.currMonitor.Process.Kill()
		s.pktContextCancel()
		log.Printf("Successfully killed packet capture stream.\n")
	} else {
		log.Printf("Packet capture stream not found.\n")
	}

	return req, nil
}

func (*hcnServer) GetHCNLogs(ctx context.Context, req *pb.HCNRequest) (*pb.HCNResponse, error) {
	hcntype := pb.HCNType(req.GetHcntype())
	viewJson := req.GetJson()

	fmt.Printf("GetHCNLogs function was invoked for %s.\n", hcntype)

	logs := nets.GetLogs(hcntype.String(), viewJson)
	res := &pb.HCNResponse{
		HcnResult: logs,
	}
	log.Printf("Sending: \n%v", res)

	return res, nil
}

func main() {
	// User input variables
	var port string

	// Flags
	flag.StringVarP(&port, "port", "p", common.DefaultServerPort, "Specify port for server to listen on.")
	flag.Parse()

	// Input validation
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Supplied value was not a valid port.")
	}

	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Printf("Server started on port %s\n", port)
	s := grpc.NewServer()
	pb.RegisterCaptureServiceServer(s, &captureServer{})
	pb.RegisterHCNServiceServer(s, &hcnServer{})

	// Register reflection service on gRPC server
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
