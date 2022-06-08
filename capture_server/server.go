package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"winspect/capturespb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Microsoft/hcsshim/hcn"
)

var pktParams = map[string]string{
	"protocols": "-t",
	"ips":       "-i",
	"ports":     "-p",
	"macs":      "-m",
}

type server struct {
	capturespb.UnimplementedCaptureServiceServer
}

func pktmonReset(clear_filters bool) error {
	// Stop packet monitor
	if err := exec.Command("cmd", "/c", "pktmon stop").Run(); err != nil {
		log.Fatalf("Failed to stop pktmon when reseting: %v", err)
	}

	// Optionally clear filters
	if clear_filters {
		if err := exec.Command("cmd", "/c", "pktmon filter remove").Run(); err != nil {
			log.Fatalf("Failed to remove old filters: %v", err)
		}
	}

	return nil
}

func (*server) StartCapture(req *capturespb.CaptureRequest, stream capturespb.CaptureService_StartCaptureServer) error {
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

	if len(args["protocols"]) == 0 {
		args["protocols"] = append(args["protocols"], "")
	}

	// Ensure filters are reset and add new ones
	pktmonReset(true)
	for _, protocol := range args["protocols"] {
		name := " winspect" + protocol + " "
		filters := []string{}

		for arg, addrs := range args {
			// Short circuiting conditional for adding protocol(s) if in filter request
			if len(addrs) > 0 && len(addrs[0]) > 0 {
				filters = append(filters, pktParams[arg]+" "+strings.Join(addrs, " "))
			}
		}

		if err := exec.Command("cmd", "/c", "pktmon filter add"+name+strings.Join(filters, " ")).Run(); err != nil {
			log.Fatalf("Failed to add%sfilter: %v", name, err)
		}
	}

	// Execute pktmon command and check for errors
	exec.Command("cmd", "/c", "pktmon stop").Run()
	cmd := exec.Command("cmd", "/c", "pktmon start -c -m real-time")
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Scanning loop with timeout constraint
	var i int
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)
	for start := time.Now(); scanner.Scan(); {
		// Only check timeout every 10 iterations
		if i%10 == 0 {
			// if dur < 0, run until client sends StopCapture
			if dur > 0 && time.Since(start) > time.Second*time.Duration(dur) {
				break
			}
		}

		// Loop main body
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
	pktmonReset(true)

	return nil
}

func (*server) StopCapture(ctx context.Context, req *capturespb.Empty) (*capturespb.Empty, error) {
	if err := exec.Command("cmd", "/c", "pktmon stop").Run(); err != nil {
		log.Printf("Failed to stop pktmon: %v", err)
	}

	return req, nil
}

func (*server) GetHCNLogs(ctx context.Context, req *capturespb.HCNRequest) (*capturespb.HCNResponse, error) {
	// id := request.Id
	// fmt.Printf("User ID is: %d", id)

	// response := &UserResponse{
	// 	Name: "John Doe",
	// }
	// return response, nil
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
	// User variables declaration
	var port string

	// Flags declaration
	flag.StringVar(&port, "p", "50051", "Specify port for server to listen on.")
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
