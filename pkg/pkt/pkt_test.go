// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pkt

import (
	"testing"

	pb "github.com/microsoft/wcnspect/rpc"
)

func TestAddFilters(t *testing.T) {
	cases := []struct {
		desc    string
		filters *pb.Filters
	}{
		{"TestNoFilters", &pb.Filters{}},
		{"TestNoProtocol", &pb.Filters{
			Ips:   []string{"126.167.32.175"},
			Ports: []string{"43"},
		}},
		{"TestOnlyProtocol", &pb.Filters{
			Protocols: []string{"UDP"},
		}},
		{"TestMultipleProtocols", &pb.Filters{
			Protocols: []string{"TCP", "ICMP"},
		}},
		{"TestTCPFlagProtocols", &pb.Filters{
			Protocols: []string{"TCP_SYN", "TCP_ACK", "TCP_ECE"},
		}},
		{"TestAllFiltersOneArgs", &pb.Filters{
			Ips:       []string{"251.151.118.164"},
			Protocols: []string{"UDP"},
			Ports:     []string{"6788"},
			Macs:      []string{"BA-3E-32-37-7F-1A"},
		}},
		{"TestAllFiltersMaxArgs", &pb.Filters{
			Ips:       []string{"251.151.118.164", "16.74.81.164"},
			Protocols: []string{"UDP", "ICMP", "TCP_ACK"},
			Ports:     []string{"6788", "10022"},
			Macs:      []string{"BA-3E-32-37-7F-1A", "F1-0E-44-23-E8-72"},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := AddFilters(tc.filters); err != nil {
				t.Log("error should be nil -", err)
				t.Fail()
			}

			if err := ResetFilters(); err != nil {
				t.Log("error resetting filters -", err)
			}
		})
	}
}

func TestModifyCaptureCmd(t *testing.T) {
	cases := []struct {
		desc     string
		mods     *pb.Modifiers
		expected string
	}{
		{"TestNoModifiers", &pb.Modifiers{}, "pktmon start -c -m real-time --type all"},
		{
			"TestOnlyDropPackets",
			&pb.Modifiers{
				PacketType: pb.PacketType(pb.PacketType_value["drop"]),
			},
			"pktmon start -c -m real-time --type drop",
		},
		{
			"TestAllModifiers",
			&pb.Modifiers{
				Pods:         []string{}, // Can't test pods here because getting the IDs pulls on hcn.ListEndpoints()
				PacketType:   pb.PacketType(pb.PacketType_value["flow"]),
				CountersOnly: true,
			},
			"pktmon start -c -m real-time --type flow --counters-only",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual, err := ModifyCaptureCmd(tc.mods)

			if err != nil {
				t.Log("error should be nil -", err)
				t.Fail()
			}

			if actual != tc.expected {
				t.Fatalf("expected: '%s' got: '%s' for mods: %v", tc.expected, actual, tc.mods)
			}
		})
	}
}
