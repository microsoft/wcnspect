package cmd

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/microsoft/winspect/common"
	"github.com/microsoft/winspect/pkg/comprise"
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
		Long:  `An advanced distributed packet capture and HNS log collection tool made with Go (^_^)`,
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

	winNodeNames map[string]v1.Node // node name -> v1.Node
	podNames     map[string]v1.Pod  // pod name -> v1.Pod

	podsNode map[string]v1.Node // pod name -> v1.Node (pod's node)
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

	// Pull windows nodes
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: "kubernetes.io/os=windows"})
	if err != nil {
		log.Fatal(err)
	}

	// Pull pods while creating mapping of node : pods
	cc.winNodeNames = make(map[string]v1.Node)
	cc.podNames = make(map[string]v1.Pod)
	cc.podsNode = make(map[string]v1.Node)
	for _, node := range nodes.Items {
		nodeName := node.GetName()

		// Set node names to node items
		cc.winNodeNames[nodeName] = node

		// Pull pods
		filter := fmt.Sprintf("spec.nodeName=%s", nodeName)
		pods, err := clientset.CoreV1().Pods(common.Namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: filter})
		if err != nil {
			log.Fatal(err)
		}

		// Set pods' names to node
		for _, pod := range pods.Items {
			podName := pod.GetName()
			cc.podsNode[podName] = node
			cc.podNames[podName] = pod
		}
	}

	// Validate there is at least one windows node
	if len(cc.winNodeNames) == 0 {
		log.Fatal("no windows nodes exist")
	}
}

func (cc *winspectBuilderCommon) getNodes(nodeNames []string) (ret []v1.Node) {
	for _, name := range nodeNames {
		ret = append(ret, cc.winNodeNames[name])
	}
	return
}

/* Takes a list of pod names. Creates a map with these pods and returns it.
Returns a map of node names to a list of pod ips.
*/
func (cc *winspectBuilderCommon) getNodePodMap(podNames []string) map[string][]string {
	ret := make(map[string][]string)

	for _, podName := range podNames {
		node := cc.podsNode[podName]

		nodeName := node.GetName()
		if nodeName != "" {
			ret[nodeName] = append(ret[nodeName], cc.podNames[podName].Status.PodIP)
		}
	}

	return ret
}

/*
Returns a list of the nodes associated with the passed pod names.
*/
func (cc *winspectBuilderCommon) getPodsNodes(podNames []string) (ret []v1.Node) {
	for _, name := range podNames {
		ret = append(ret, cc.podsNode[name])
	}
	return
}

func (cc *winspectBuilderCommon) getPodNames() []string {
	return comprise.Keys(cc.podNames)
}

func (cc *winspectBuilderCommon) getWinNodes() []v1.Node {
	//FIXME: the below line should be commented out if not testing on local
	// return []v1.Node{{ObjectMeta: metav1.ObjectMeta{Name: "localhost"}, Status: v1.NodeStatus{Addresses: []v1.NodeAddress{{Type: "InternalIP", Address: "0.0.0.0"}}}}}
	return comprise.Values(cc.winNodeNames)
}

func (cc *winspectBuilderCommon) getWinNodeNames() []string {
	return comprise.Keys(cc.winNodeNames)
}
