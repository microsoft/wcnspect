package server

import (
	"context"
	"fmt"
	"log"
	"math"
	"os/exec"
	"sync"
	"time"

	"github.com/microsoft/winspect/pkg/nets"
	pb "github.com/microsoft/winspect/rpc"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CaptureServer struct {
	pb.UnimplementedCaptureServiceServer
	currMonitor      *exec.Cmd          // Tracks the running pktmon stream
	pktContextCancel context.CancelFunc // Tracks the pktmon stream's context's cancel func
	printCounters    bool
}

type HcnServer struct {
	pb.UnimplementedHCNServiceServer
}

func (s *CaptureServer) StartCapture(req *pb.CaptureRequest, stream pb.CaptureService_StartCaptureServer) error {
	fmt.Printf("StartCapture function was invoked with %v\n", req)
	pktmonStartCommand := "pktmon start -c -m real-time"

	// Retrieve and format request arguments
	dur := req.GetDuration()
	modifiers := req.GetModifier()
	filters := req.GetFilter()
	s.printCounters = modifiers.GetCountersOnly()

	// If duration is less than 0, we run for an "infinite" amount of time
	if dur <= 0 {
		dur = math.MaxInt32
	}

	// Ensure filters are reset and add new ones
	resetPktmon(true, true)
	addPktmonFilters(filters)

	// Revise pktmonStartCommand based on Modifiers
	pktmonStartCommand = revisePktmonCommand(modifiers, pktmonStartCommand)

	// Create a timeout context and set as server's pktmon canceller
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dur)*time.Second)
	s.pktContextCancel = cancel

	// Execute pktmon command and check for errors, if successful, set as server's currMonitor
	exec.Command("cmd", "/c", "pktmon stop").Run()
	cmd := exec.CommandContext(ctx, "cmd", "/c", pktmonStartCommand)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Print(err)
	}
	if err := cmd.Start(); err != nil {
		log.Print(err)
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
				if s.printCounters {
					time.Sleep(time.Millisecond * 100)
					continue
				}

				res := &pb.CaptureResponse{
					Result:    out,
					Timestamp: timestamppb.Now(),
				}

				stream.Send(res)
				log.Printf("Sent: \n%v", res)
			case <-ctx.Done():
				log.Printf("Packet monitoring stream finished.")
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()

	// If timeout reached and printCounters, then send counter table
	if s.printCounters {
		res := &pb.CaptureResponse{
			Result:    pullCounters(),
			Timestamp: timestamppb.Now(),
		}

		stream.Send(res)
		log.Printf("Sent: \n%v", res)
	}

	// Reset pktmon filters and CaptureServer's fields
	cancel()
	resetPktmon(false, true)
	resetCaptureContext(s)
	log.Printf("Packet monitor filters reset.")

	return nil
}

func (s *CaptureServer) StopCapture(ctx context.Context, req *pb.Empty) (*pb.StopCaptureResponse, error) {
	fmt.Println("StopCapture function was invoked.")
	msg := ""

	if s.currMonitor != nil {
		if s.printCounters {
			msg = pullCounters()
			s.printCounters = false
		}

		s.currMonitor.Process.Kill()
		s.pktContextCancel()
		log.Printf("Successfully killed packet capture stream.\n")
	} else {
		log.Printf("Packet capture stream not found.\n")
	}

	res := &pb.StopCaptureResponse{
		Result:    msg,
		Timestamp: timestamppb.Now(),
	}
	log.Printf("Sending: \n%v", res)

	return res, nil
}

func (s *CaptureServer) GetCounters(ctx context.Context, req *pb.CountersRequest) (*pb.CountersResponse, error) {
	fmt.Println("GetCounters function was invoked.")
	pktmonCmd := "pktmon counter"

	if req.IncludeHidden {
		pktmonCmd += " --include-hidden"
	}

	cmd := exec.Command("cmd", "/c", pktmonCmd)

	out, err := cmd.Output()
	if err != nil {
		log.Print(err)
	}

	res := &pb.CountersResponse{
		Result: string(out),
	}
	log.Printf("Sending: \n%v", res)

	return res, nil
}

func (*HcnServer) GetHCNLogs(ctx context.Context, req *pb.HCNRequest) (*pb.HCNResponse, error) {
	hcntype := pb.HCNType(req.GetHcntype())
	verbose := req.GetVerbose()

	fmt.Printf("GetHCNLogs function was invoked for %s.\n", hcntype)

	logs := nets.GetLogs(hcntype.String(), verbose)
	res := &pb.HCNResponse{
		HcnResult: logs,
	}
	log.Printf("Sending: \n%v", res)

	return res, nil
}
