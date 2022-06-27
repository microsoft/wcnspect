package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
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

var (
	kubeconfig  string
	targetNodes []string
	nodeSet     []v1.Node
	podSet      []v1.Pod
)

var rootCmd = &cobra.Command{
	Use:   "winspect",
	Short: "Winspect is an advanced distributed packet capture and HNS log collection tool.",
	Long:  `An advanced distributed packet capture and HNS log collection tool made with Go (^‿^)`,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if home := homedir.HomeDir(); home != "" {
		rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "Optionally specify absolute path to the kubeconfig file.")
	} else {
		rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Specify absolute path to the kubeconfig file.")
		rootCmd.MarkPersistentFlagRequired("kubeconfig")
	}

	// Use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// create the clientset
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

	// Create node and pod sets
	nodeSet = nodes.Items
	podSet = pods.Items

	// Validation
	targetNodes = client.ParseValidateNodes(targetNodes, nodeSet)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
