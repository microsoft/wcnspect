package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"time"

	"winspect/capturespb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	capturespb.UnimplementedCaptureServiceServer
}

func (*server) StartCapture(req *capturespb.CaptureRequest, stream capturespb.CaptureService_StartCaptureServer) error {
	fmt.Printf("StartCapture function was invoked with %v\n", req)
	dur := req.GetDuration()

	// ensure pktmon isn't running, then execute command and check for errors
	exec.Command("cmd", "/c", "pktmon stop").Run()
	cmd := exec.Command("cmd", "/c", "pktmon start -c -m real-time")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// scanning loop with timeout constraint
	var i int
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for start := time.Now(); scanner.Scan(); {
		// only check timeout every 10 iterations
		if i%10 == 0 {
			// if dur < 0, run until client sends StopCapture
			if dur > 0 && time.Since(start) > time.Second*time.Duration(dur) {
				break
			}
		}

		// loop main body
		m := scanner.Text()
		res := &capturespb.CaptureResponse{
			Result:    m,
			Timestamp: timestamppb.Now(),
		}
		stream.Send(res)
		log.Printf("Sent: \n%v", res)
		i++
	}

	// Stop pktmon if still running
	if err := exec.Command("cmd", "/c", "pktmon stop").Run(); err != nil {
		log.Printf("failed to stop pktmon at end of stream: %v", err)
	}

	return nil
}

func (*server) StopCapture(ctx context.Context, req *capturespb.Empty) (*capturespb.Empty, error) {
	if err := exec.Command("cmd", "/c", "pktmon stop").Run(); err != nil {
		log.Printf("failed to stop pktmon: %v", err)
	}

	return req, nil
}

func main() {
	// User variables declaration
	var port string

	// Flags declaration
	flag.StringVar(&port, "p", "50051", "Specify port for server to listen on. Default is 50051.")
	flag.Parse()

	// Input validation
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Supplied value was not a valid port.")
	}

	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Print("Server started")
	s := grpc.NewServer()
	capturespb.RegisterCaptureServiceServer(s, &server{})

	// Register reflection service on gRPC server
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
