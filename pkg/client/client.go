package client

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/microsoft/winspect/common"
	"github.com/microsoft/winspect/pkg/k8spi"
	pb "github.com/microsoft/winspect/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
)

type client struct {
	pb.CaptureServiceClient
	pb.HCNServiceClient
}

type Node struct {
	Name string
	Ip   string
}

type ReqContext struct {
	Server Node
	Wg     *sync.WaitGroup
}

func (rq *ReqContext) Done() {
	if rq.Wg != nil {
		rq.Wg.Done()
	}
}

func CreateConnection(ip string) (*client, func() error) {
	//FIXME: hardcoded port addition
	cc, err := grpc.Dial(ip+":"+common.DefaultServerPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	c1 := pb.NewCaptureServiceClient(cc)
	c2 := pb.NewHCNServiceClient(cc)
	c := &client{c1, c2}

	return c, cc.Close
}

func RunCaptureStream(c pb.CaptureServiceClient, req *pb.CaptureRequest, reqCtx *ReqContext) {
	name, ip := reqCtx.Server.Name, reqCtx.Server.Ip
	fmt.Printf("Starting to do a Server Streaming RPC from %s (IP: %s)...\n", name, ip)

	// Create request object
	req.Timestamp = timestamppb.Now()

	// Send request
	resStream, err := c.StartCapture(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling StartCapture RPC from %s (from IP: %s): %v", name, ip, err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// we've reached the end of the stream
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream from %s (from IP: %s): %v", name, ip, err)
		}

		fmt.Printf("Response from %s StartCapture (%s) sent at %s: \n%v\n", name, ip, msg.GetTimestamp().AsTime(), msg.GetResult())
	}

	fmt.Printf("Finished receiving stream from %s (IP: %s).\n", name, ip)

	reqCtx.Done()
}

func RunStopCapture(c pb.CaptureServiceClient, reqCtx *ReqContext) {
	name, ip := reqCtx.Server.Name, reqCtx.Server.Ip
	res, err := c.StopCapture(context.Background(), &pb.Empty{})
	if err != nil {
		log.Fatalf("error while calling StopCapture RPC from %s (IP: %s): %v", name, ip, err)
	}

	msg, timestamp := res.GetResult(), res.GetTimestamp().AsTime()
	if len(msg) != 0 {
		fmt.Printf("StopCapture successfully ran on node: %s (IP: %s) at time: %s with output: \n%s\n", name, ip, timestamp, msg)
	} else {
		fmt.Printf("StopCapture successfully ran on node: %s (IP: %s) at time: %s.\n", name, ip, timestamp)
	}

	fmt.Printf("Packet capture ended on node: %s (IP: %s).\n", name, ip)

	reqCtx.Done()
}

func PrintCounters(c pb.CaptureServiceClient, req *pb.CountersRequest, reqCtx *ReqContext) {
	name, ip := reqCtx.Server.Name, reqCtx.Server.Ip
	fmt.Printf("Requesting packet counters table from %s (IP: %s)...\n", name, ip)

	// Send request
	res, err := c.GetCounters(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GetCounters RPC from %s (IP: %s): %v", name, ip, err)
	}

	msg, timestamp := res.GetResult(), res.GetTimestamp().AsTime()
	fmt.Printf("Received GetCounters RPC response from %s (IP: %s) at time: %s -\n%s\n", name, ip, timestamp, msg)

	reqCtx.Done()
}

func PrintVFPCounters(c pb.CaptureServiceClient, req *pb.VFPCountersRequest, reqCtx *ReqContext) {
	name, ip := reqCtx.Server.Name, reqCtx.Server.Ip
	fmt.Printf("Requesting VFP packet counters table from %s (IP: %s)...\n", name, ip)

	// Send request
	res, err := c.GetVFPCounters(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GetVFPCounters RPC from %s (IP: %s): %v", name, ip, err)
	}

	msg, timestamp := res.GetResult(), res.GetTimestamp().AsTime()
	fmt.Printf("Received GetVFPCounters RPC response from %s (IP: %s) at time: %s -\n%s\n", name, ip, timestamp, msg)

	reqCtx.Done()
}

func PrintHCNLogs(c pb.HCNServiceClient, req *pb.HCNRequest, reqCtx *ReqContext) {
	hcntype, name, ip := pb.HCNType_name[int32(req.GetHcntype())], reqCtx.Server.Name, reqCtx.Server.Ip
	fmt.Printf("Requesting HCN logs from %s (IP: %s)...\n", name, ip)

	// Send request
	res, err := c.GetHCNLogs(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GetHCNLogs RPC from %s (IP: %s): %v", name, ip, err)
	}

	fmt.Printf("Received logs for %s from %s (IP: %s):\n\n%s\n", hcntype, name, ip, string(res.GetHcnResult()))

	reqCtx.Done()
}

func Cleanup(nodes []v1.Node) {
	var wg sync.WaitGroup
	for _, node := range nodes {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Get target name and ip
		name, ip := node.GetName(), k8spi.RetrieveInternalIP(node)

		// Create connections
		c, closeClient := CreateConnection(ip)
		defer closeClient()

		// Create request context
		ctx := &ReqContext{
			Server: Node{
				Name: name,
				Ip:   ip,
			},
			Wg: &wg,
		}

		// Launch a goroutine to run the request
		go RunStopCapture(c, ctx)
	}

	// Wait for all captures to complete
	wg.Wait()
}
