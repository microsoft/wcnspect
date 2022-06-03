package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"winspect/capturespb"

	"google.golang.org/grpc"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := capturespb.NewCaptureServiceClient(cc)

	doServerStreaming(c)
}

func doServerStreaming(c capturespb.CaptureServiceClient) {
	fmt.Println("Starting to do a Server Streaming RPC...")

	req := &capturespb.CaptureRequest{Duration: 10}

	resStream, err := c.StartCapture(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling StartCapture RPC: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// we've reached the end of the stream
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}

		fmt.Printf("Response from StartCapture: %v\n", msg.GetResult())
	}
}
