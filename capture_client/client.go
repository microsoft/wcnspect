package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"winspect/capturespb"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	var wg sync.WaitGroup
	var ips = []string{
		"localhost:50051",
		//"10.216.116.143:50051", // corresponding to t-nikoVM1
		//"10.216.119.237:50051", // corresponding to t-nikoVM2
		//"10.224.0.35:50051", // corresponding to internal aks windows node
		//"10.224.0.67:50051", // corresponding to internal aks subnet myVM2
	}

	for _, ip := range ips {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Launch a goroutine to run the capture
		go createConnection(ip, &wg)
	}

	// Wait for all captures to complete
	wg.Wait()
}

func createConnection(ip string, wg *sync.WaitGroup) {
	// Decrement relevant waitgroup counter when goroutine completes
	defer wg.Done()

	cc, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := capturespb.NewCaptureServiceClient(cc)

	doServerStreaming(c, ip)
}

func doServerStreaming(c capturespb.CaptureServiceClient, ip string) {
	fmt.Printf("Starting to do a Server Streaming RPC (from IP: %s)...\n", ip)

	// Create request object
	req := &capturespb.CaptureRequest{
		Duration:  2,
		Timestamp: timestamppb.Now(),
		Filter: &capturespb.Filters{
			Ips:       []string{"157.58.214.96", "10.16.84.36"},
			Protocols: []string{"TCP"},
			Ports:     []string{"49422"},
			Macs:      []string{},
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
