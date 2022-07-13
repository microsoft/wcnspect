package cmd

import (
	"context"
	"log"
	"path/filepath"

	"github.com/microsoft/winspect/common"
	"github.com/microsoft/winspect/pkg/client"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type commandsBuilder struct {
	winspectBuilderCommon

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

func (b *commandsBuilder) build() *winspectCmd {
	h := b.newWinspectCmd()
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

type winspectCmd struct {
	*baseBuilderCmd
}

func (b *commandsBuilder) newWinspectCmd() *winspectCmd {
	cc := &winspectCmd{}

	cc.baseBuilderCmd = b.newBuilderCmd(&cobra.Command{
		Use:   "winspect",
		Short: "Winspect is an advanced distributed packet capture and HNS log collection tool.",
		Long:  `An advanced distributed packet capture and HNS log collection tool made with Go (^‿^)`,
	})

	if home := homedir.HomeDir(); home != "" {
		cc.cmd.PersistentFlags().StringVar(&cc.kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Optionally specify absolute path to the kubeconfig file.")
	} else {
		cc.cmd.PersistentFlags().StringVar(&cc.kubeconfig, "kubeconfig", "", "Specify absolute path to the kubeconfig file.")
		cc.cmd.MarkPersistentFlagRequired("kubeconfig")
	}

	cc.cmd.CompletionOptions.DisableDefaultCmd = true

	cc.initializeAKSClusterValues()

	return cc
}

type winspectBuilderCommon struct {
	kubeconfig string

	targetNodes []string
	nodeSet     []v1.Node
	podSet      []v1.Pod
}

func (cc *winspectBuilderCommon) initializeAKSClusterValues() {
	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", cc.kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Pull nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Pull pods
	pods, err := clientset.CoreV1().Pods(common.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Set node and pod sets
	cc.nodeSet = nodes.Items
	cc.podSet = pods.Items

	// Validation
	cc.targetNodes = client.ParseValidateNodes(cc.targetNodes, cc.nodeSet)
	if len(cc.targetNodes) == 0 {
		log.Fatal("no windows nodes exist")
	}
}

func (cc *winspectBuilderCommon) getTargetNodes() []string {
	// return []string{"localhost"} //FIXME: this line should be commented out if not testing
	return append([]string{}, cc.targetNodes...)
}
