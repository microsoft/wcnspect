// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/microsoft/wcnspect/common"
	"github.com/microsoft/wcnspect/pkg/server"
	pb "github.com/microsoft/wcnspect/rpc"

	flag "github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// User input variables
	var port string

	// Flags
	flag.StringVarP(&port, "port", "p", common.DefaultServerPort, "Specify port for server to listen on.")
	flag.Parse()

	// Input validation
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Supplied value was not a valid port.")
	}

	listener, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Printf("Server started on port %s\n", port)
	s := grpc.NewServer()
	pb.RegisterCaptureServiceServer(s, &server.CaptureServer{})
	pb.RegisterHCNServiceServer(s, &server.HcnServer{})

	// Register reflection service on gRPC server
	reflection.Register(s)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
