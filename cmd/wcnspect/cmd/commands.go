// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmd

import (
	"fmt"
	"log"

	"github.com/microsoft/wcnspect/pkg/comprise"
	"github.com/microsoft/wcnspect/pkg/k8sapi"
	"github.com/spf13/cobra"

	v1 "k8s.io/api/core/v1"
)

type commandsBuilder struct {
	wcnspectBuilderCommon

	commands []cmder
}

type cmder interface {
	getCommand() *cobra.Command
}

func newCommandsBuilder() *commandsBuilder {
	return &commandsBuilder{}
}

func (b *commandsBuilder) addCommands(commands ...cmder) *commandsBuilder {
	b.commands = append(b.commands, commands...)
	return b
}

func (b *commandsBuilder) addAll() *commandsBuilder {
	b.addCommands(
		b.newCaptureCmd(),
		b.newCounterCmd(),
		b.newHnsCmd(),
		b.newVfpCounterCmd(),
	)

	return b
}

func (b *commandsBuilder) build() *wcnspectCmd {
	h := b.newwcnspectCmd()
	addCommands(h.getCommand(), b.commands...)
	return h
}

func addCommands(root *cobra.Command, commands ...cmder) {
	for _, command := range commands {
		cmd := command.getCommand()
		root.AddCommand(cmd)
	}
}

type baseCmd struct {
	cmd *cobra.Command
}

type baseBuilderCmd struct {
	*baseCmd
	*commandsBuilder
}

func (c *baseCmd) getCommand() *cobra.Command {
	return c.cmd
}

func (b *commandsBuilder) newBuilderCmd(cmd *cobra.Command) *baseBuilderCmd {
	bcmd := &baseBuilderCmd{commandsBuilder: b, baseCmd: &baseCmd{cmd: cmd}}
	return bcmd
}

type wcnspectCmd struct {
	*baseBuilderCmd
}

func (b *commandsBuilder) newwcnspectCmd() *wcnspectCmd {
	cc := &wcnspectCmd{}

	cc.baseBuilderCmd = b.newBuilderCmd(&cobra.Command{
		Use:   "wcnspect",
		Short: "wcnspect is an advanced distributed packet capture and HNS log collection tool.",
		Long:  `An advanced distributed packet capture and HNS log collection tool made with Go (^_^)`,
	})
	cc.cmd.PersistentFlags().StringVar(&cc.kubeconfig, "kubeconfig", "", "Specify absolute path to the kubeconfig file.")
	cc.cmd.CompletionOptions.DisableDefaultCmd = true
	cc.initializeAKSClusterValues()

	return cc
}

type wcnspectBuilderCommon struct {
	kubeconfig string

	winNodeNames map[string]v1.Node // node name -> v1.Node
}

func (cc *wcnspectBuilderCommon) initializeAKSClusterValues() {
	k8sclient = k8sapi.New(cc.kubeconfig)

	// Pull windows nodes
	nodes := k8sclient.GetAllNodesWindows()

	// Pull pods while creating mapping of node : pods
	cc.winNodeNames = make(map[string]v1.Node)

	for _, node := range nodes.Items {
		nodeName := node.GetName()

		// Set node names to node items
		cc.winNodeNames[nodeName] = node
	}

	// Validate there is at least one windows node
	if len(cc.winNodeNames) == 0 {
		log.Fatal("no Windows nodes exist")
	}
}

func (cc *wcnspectBuilderCommon) getNodes(nodeNames []string) (ret []v1.Node) {
	for _, name := range nodeNames {
		ret = append(ret, cc.winNodeNames[name])
	}
	return
}

func (cc *wcnspectBuilderCommon) getNode(nodeName string) (node v1.Node) {
	if n, ok := cc.winNodeNames[nodeName]; ok {
		node = n
	} else {
		log.Fatalf(fmt.Sprintf("Windows node %s not found", nodeName))
	}
	return
}

func (cc *wcnspectBuilderCommon) getWinNodes() []v1.Node {
	//FIXME: the below line should be commented out if not testing on local
	// return []v1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "localhost"}, Status: v1.NodeStatus{Addresses: []v1.NodeAddress{{Type: "InternalIP", Address: "0.0.0.0"}}}}}
	return comprise.Values(cc.winNodeNames)
}

func (cc *wcnspectBuilderCommon) getWinNodeNames() []string {
	return comprise.Keys(cc.winNodeNames)
}
