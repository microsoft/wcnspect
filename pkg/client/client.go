package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/microsoft/winspect/common"
	pb "github.com/microsoft/winspect/rpc"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CaptureParams struct {
	Node      string
	Pods      []string
	Ips       []string
	Protocols []string
	Ports     []string
	Macs      []string
	Time      int32
}

type HCNParams struct {
	Cmd     string
	Node    string
	Verbose bool
}

type client struct {
	pb.CaptureServiceClient
	pb.HCNServiceClient
}

func CreateConnection(ip string) (*client, func() error) {
	//FIXME: hardcoded port addition
	cc, err := grpc.Dial(ip+":"+common.DefaultServerPort, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	c1 := pb.NewCaptureServiceClient(cc)
	c2 := pb.NewHCNServiceClient(cc)
	c := &client{c1, c2}

	return c, cc.Close
}

func RunCaptureStream(c pb.CaptureServiceClient, args *CaptureParams, wg *sync.WaitGroup) {
	ip := args.Node
	fmt.Printf("Starting to do a Server Streaming RPC (from IP: %s)...\n", ip)

	// Create request object
	req := &pb.CaptureRequest{
		Duration:  args.Time,
		Timestamp: timestamppb.Now(),
		Filter: &pb.Filters{
			Pods:      args.Pods,
			Ips:       args.Ips,
			Protocols: args.Protocols,
			Ports:     args.Ports,
			Macs:      args.Macs,
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

	if wg != nil {
		wg.Done()
	}
}

func RunStopCapture(c pb.CaptureServiceClient, ip string, wg *sync.WaitGroup) {
	_, err := c.StopCapture(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("error while calling StopCapture RPC (from IP: %s): %v", ip, err)
	}

	fmt.Printf("Ended packet capture on IP: %s.\n", ip)

	if wg != nil {
		wg.Done()
	}
}

func PrintHCNLogs(c pb.HCNServiceClient, args *HCNParams, wg *sync.WaitGroup) {
	hcntype, ip := args.Cmd, args.Node
	fmt.Printf("Requesting HCN logs (from IP: %s)...\n", ip)

	// Create request object
	req := &pb.HCNRequest{
		Hcntype: pb.HCNType(pb.HCNType_value[hcntype]),
		Verbose: args.Verbose,
	}

	// Send request
	res, err := c.GetHCNLogs(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GetHCNLogs RPC (from IP: %s): %v", ip, err)
	}

	fmt.Printf("Received logs for %s (from IP: %s):\n\n%s\n", hcntype, ip, string(res.GetHcnResult()))

	if wg != nil {
		wg.Done()
	}
}

func Cleanup(nodes []string) {
	var wg sync.WaitGroup
	for _, ip := range nodes {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Create connections
		c, closeClient := CreateConnection(ip)
		defer closeClient()

		// Launch a goroutine to run the request
		go RunStopCapture(c, ip, &wg)
	}

	// Wait for all captures to complete
	wg.Wait()
}