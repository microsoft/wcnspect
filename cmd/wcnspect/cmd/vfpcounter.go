// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"sync"

	"github.com/microsoft/wcnspect/common"
	"github.com/microsoft/wcnspect/pkg/client"
	"github.com/microsoft/wcnspect/pkg/k8sapi"
	pb "github.com/microsoft/wcnspect/rpc"
	v1 "k8s.io/api/core/v1"

	"github.com/spf13/cobra"
)

type vfpCounterCmd struct {
	pod       string
	verbose   bool
	namespace string

	*baseBuilderCmd
}

func (b *commandsBuilder) newVfpCounterCmd() *vfpCounterCmd {
	cc := &vfpCounterCmd{}

	cmd := &cobra.Command{
		Use:   "vfp-counter",
		Short: "The 'vfp-counter' command will retrieve packet counter tables from a specified windows pod's port VFP.",
		Long: `The 'vfp-counter' command will retrieve packet counter tables from a specified windows pod's port VFP. 
	For example:
	'wcnspect vfp-counter --pod {pod}'`,
		Run: func(cmd *cobra.Command, args []string) {
			cc.printVFPCounters()
		},
	}

	cmd.PersistentFlags().StringVarP(&cc.pod, "pod", "p", "", "Specify which pod wcnspect should send requests to using pod name. This flag is required.")
	cmd.PersistentFlags().BoolVarP(&cc.verbose, "detailed", "d", false, "Option to output Host vNic and External Adapter Port counters.")
	cmd.PersistentFlags().StringVar(&cc.namespace, "namespace", common.DefaultNamespace, "Optionally specify Kubernetes namespace to filter pods on.")
	cmd.MarkPersistentFlagRequired("pod")

	cc.baseBuilderCmd = b.newBuilderCmd(cmd)

	return cc
}

func (cc *vfpCounterCmd) printVFPCounters() {
	// Read in pods
	pods := []string{cc.pod}
	// Namespace
	ns := k8sclient.GetNamespace(cc.namespace)
	// Loop over Pod, Node
	var p *v1.Pod
	var nodeName string
	var wg sync.WaitGroup

	for _, podName := range pods {
		p = k8sclient.GetPod(podName, ns.GetName())
		nodeName = p.Spec.NodeName
		podIP := p.Status.PodIP
		if nodeName != "" {
			wg.Add(1)
			nodeIP := k8sapi.RetrieveInternalIP(cc.getNode(nodeName))
			c, closeClient := client.CreateConnection(nodeIP)
			defer closeClient()

			ctx := &client.ReqContext{
				Server: client.Node{
					Name: nodeName,
					Ip:   nodeIP,
				},
				Wg: &wg,
			}

			req := &pb.VFPCountersRequest{
				Pod:     podIP,
				Verbose: cc.verbose,
			}

			go client.PrintVFPCounters(c, req, ctx)
		}
	}
	wg.Wait()
}
