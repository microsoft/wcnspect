package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"winspect/capturespb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	capturespb.UnimplementedCaptureServiceServer
}

func (*server) StartCapture(req *capturespb.CaptureRequest, stream capturespb.CaptureService_StartCaptureServer) error {
	fmt.Printf("StartCapture function was invoked with %v\n", req)
	dur := req.GetDuration() //TODO: default is already 0, make runtime infinite

	// execute command and check for errors
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
			if time.Since(start) > time.Second*time.Duration(dur) {
				break
			}
		}

		// loop main body
		m := scanner.Text()
		res := &capturespb.CaptureResponse{
			Result: m,
		}
		stream.Send(res)
		log.Printf("Sent: %v", res)
		i++
	}

	// Stop pktmon if still running
	exec.Command("cmd", "/c", "pktmon stop").Run()

	return nil
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
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
