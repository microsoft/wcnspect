package server

import (
	"context"
	"fmt"
	"log"
	"math"
	"os/exec"
	"time"

	"github.com/microsoft/winspect/pkg/netutil"
	"github.com/microsoft/winspect/pkg/pkt"
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

	// Retrieve and format request arguments
	dur := req.GetDuration()
	modifiers := req.GetModifier()
	filters := req.GetFilter()
	s.printCounters = modifiers.GetCountersOnly()

	// If duration is less than or equal to 0, we run for an "infinite" amount of time
	if dur <= 0 {
		dur = math.MaxInt32
	}

	// Ensure filters are reset and add new ones
	if err := pkt.ResetCaptureProgram(); err != nil {
		return err
	}

	if err := pkt.ResetFilters(); err != nil {
		return err
	}

	if err := pkt.AddFilters(filters); err != nil {
		return err
	}

	// Revise pktmonStartCommand based on Modifiers
	captureCmd, err := pkt.ModifyCaptureCmd(modifiers)
	if err != nil {
		return err
	}

	// Create a timeout context and set as server's pktmon canceller
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dur)*time.Second)
	s.pktContextCancel = cancel

	// Execute pktmon command and check for errors, if successful, set as server's currMonitor
	cmd, stdout, err := pkt.StartStream(ctx, captureCmd)
	if err != nil {
		return err
	}
	s.currMonitor = cmd

	// Create a channel to receive pktmon stream from
	c := pkt.CreateStreamChannel(stdout)

	// Goroutine with a timeout constraint and pulling on pktmon channel with scanning loop
loop:
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

			if err := stream.Send(res); err != nil {
				return err
			}

			log.Printf("Sent: \n%v", res)
		case <-ctx.Done():
			log.Printf("Packet monitoring stream finished.")
			break loop
		}
	}

	// If timeout reached and printCounters, then send counter table
	if s.printCounters {
		counters, err := pkt.PullCounters()

		if err != nil {
			return err
		}

		res := &pb.CaptureResponse{
			Result:    counters,
			Timestamp: timestamppb.Now(),
		}

		stream.Send(res)
		log.Printf("Sent: \n%v", res)
	}

	// Reset pktmon filters and CaptureServer's fields
	cancel()
	resetCaptureContext(s)

	if err := pkt.ResetFilters(); err != nil {
		return err
	}

	log.Printf("Packet monitor filters reset.")

	return nil
}

func (s *CaptureServer) StopCapture(ctx context.Context, req *pb.Empty) (*pb.StopCaptureResponse, error) {
	fmt.Println("StopCapture function was invoked.")
	var msg string
	var err error

	if s.currMonitor != nil {
		if s.printCounters {
			msg, err = pkt.PullCounters()
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

	log.Printf("Sending StopCapture execution timestamp: \n%v", res)

	return res, err
}

func (s *CaptureServer) GetCounters(ctx context.Context, req *pb.CountersRequest) (*pb.CountersResponse, error) {
	fmt.Println("GetCounters function was invoked.")
	includeHidden := req.GetIncludeHidden()

	counters, err := pkt.PullStreamCounters(includeHidden)
	res := &pb.CountersResponse{
		Result:    counters,
		Timestamp: timestamppb.Now(),
	}

	log.Printf("Sending: \n%v", res)

	return res, err
}

func (*CaptureServer) GetVFPCounters(ctx context.Context, req *pb.VFPCountersRequest) (*pb.VFPCountersResponse, error) {
	fmt.Println("GetVFPCounters function was invoked.")
	pod := req.GetPod()

	guid, err := netutil.GetPortGUID(pod)
	if err != nil {
		log.Print(err)
		return &pb.VFPCountersResponse{}, err
	}

	counters, err := pkt.PullVFPCounters(guid)
	res := &pb.VFPCountersResponse{
		Result:    counters,
		Timestamp: timestamppb.Now(),
	}

	log.Printf("Sending: \n%v", res)

	return res, err
}

func (*HcnServer) GetHCNLogs(ctx context.Context, req *pb.HCNRequest) (*pb.HCNResponse, error) {
	hcntype := pb.HCNType(req.GetHcntype())
	verbose := req.GetVerbose()

	fmt.Printf("GetHCNLogs function was invoked for %s.\n", hcntype)

	logs, err := netutil.GetLogs(hcntype.String(), verbose)
	res := &pb.HCNResponse{
		HcnResult: logs,
	}
	log.Printf("Sending: \n%v", res)

	return res, err
}

func resetCaptureContext(s *CaptureServer) {
	s.currMonitor = nil
	s.pktContextCancel = nil
	s.printCounters = false
}
