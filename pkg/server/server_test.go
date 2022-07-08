package server

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/microsoft/winspect/pkg/netutil"
	pb "github.com/microsoft/winspect/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)

	s := grpc.NewServer()
	pb.RegisterCaptureServiceServer(s, &CaptureServer{})
	pb.RegisterHCNServiceServer(s, &HcnServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestStopCapture(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewCaptureServiceClient(conn)
	res, err := client.StopCapture(ctx, &pb.Empty{})

	if err != nil {
		t.Fatalf("StopCapture failed: %v", err)
	}

	log.Printf("Response: %+v", res)

	actual := res.GetResult()
	expected := ""
	if actual != expected {
		t.Fatalf("expected: '%s' got: '%s'", expected, actual)
	}
}

func TestGetHCNLogs(t *testing.T) {
	getLogString := func(option string, verbose bool) string {
		res, _ := netutil.GetLogs(option, verbose)
		return string(res)
	}

	cases := []struct {
		desc     string
		in       *pb.HCNRequest
		expected string
	}{
		{"TestAllLogs", &pb.HCNRequest{Hcntype: pb.HCNType(pb.HCNType_value["all"])}, getLogString("all", false)},
		{"TestAllLogsVerbose", &pb.HCNRequest{Hcntype: pb.HCNType(pb.HCNType_value["all"]), Verbose: true}, getLogString("all", true)},
		{"TestNetworkLogs", &pb.HCNRequest{Hcntype: pb.HCNType(pb.HCNType_value["networks"])}, getLogString("networks", false)},
		{"TestNetworkLogs", &pb.HCNRequest{Hcntype: pb.HCNType(pb.HCNType_value["networks"]), Verbose: true}, getLogString("networks", true)},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())

			if err != nil {
				t.Fatalf("Failed to dial bufnet: %v", err)
			}
			defer conn.Close()

			client := pb.NewHCNServiceClient(conn)
			res, err := client.GetHCNLogs(ctx, tc.in)

			if err != nil {
				t.Fatalf("GetHCNLogs failed: %v", err)
			}

			log.Printf("Response: %+v", res)

			actual := string(res.GetHcnResult())
			if actual != tc.expected {
				t.Fatalf("expected: \n%s\n got: \n%s\n", tc.expected, actual)
			}
		})
	}
}
