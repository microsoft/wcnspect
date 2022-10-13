// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package k8sapi

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/microsoft/wcnspect/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// K8sapi is used to request data from  the Kubernetes API-Server
type K8sapi struct {
	config *rest.Config
	conn   *kubernetes.Clientset
}

// Constructor for K8sapi
func New(config string) K8sapi {
	kubeconfig := loadConfiguration(config)

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Fatalln("Error reading kubeconfig: ", err)
	}
	return K8sapi{kubeconfig, clientset}
}

func loadConfiguration(config string) *rest.Config {
	kubeconfig, err := rest.InClusterConfig()
	if err == nil {
		log.Println("using in cluster config")
		return kubeconfig
	}

	// $HOME/.kube/config
	home := homedir.HomeDir()
	filename := filepath.Join(home, ".kube", "config")

	// Check KUBECONFIG environment var
	if conf, ok := os.LookupEnv(common.KubeConfigEnvVar); ok && len(conf) != 0 {
		config = filepath.Clean(conf)
		// Check $HOME/.kube/config
	} else if _, err := ioutil.ReadFile(filename); err == nil {
		config = filename
	} else {
		log.Fatal("Error reading $KUBECONFIG environment variable. Exiting...")
	}

	// Use the current context in kubeconfig
	kubeconfig, err = clientcmd.BuildConfigFromFlags("", config)
	if err != nil {
		log.Fatalln("Error reading kubeconfig: ", err)
	}
	return kubeconfig
}

func (k8sclient *K8sapi) GetPod(podName string, kubenamespace string) *v1.Pod {
	pod, err := k8sclient.conn.CoreV1().Pods(kubenamespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return pod
}

func (k8sclient *K8sapi) GetNode(nodeName string) *v1.Node {
	node, err := k8sclient.conn.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return node
}

func (k8sclient *K8sapi) GetAllNodesWindows() *v1.NodeList {
	nodes, err := k8sclient.conn.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: "kubernetes.io/os=windows"})
	if err != nil {
		log.Fatal(err)
	}
	return nodes
}

func (k8sclient *K8sapi) GetNamespace(namespace string) *v1.Namespace {
	ns, err := k8sclient.conn.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	return ns
}

// Operations on v1.Node objects
func FilterNodes(nodes []v1.Node, test func(v1.Node) bool) (ret []v1.Node) {
	for _, node := range nodes {
		if test(node) {
			ret = append(ret, node)
		}
	}
	return

}

/*
	Retrieves node names and ips given a list of nodes

return map of node name to node ip
*/
func GetNodesNameToIp(nodes []v1.Node) map[string]string {
	ret := map[string]string{}
	for _, node := range nodes {
		ret[node.GetName()] = RetrieveInternalIP(node)
	}

	return ret
}

func GetNodesIpToName(nodes []v1.Node) map[string]string {
	ret := map[string]string{}
	for _, node := range nodes {
		ret[RetrieveInternalIP(node)] = node.GetName()
	}

	return ret
}

func MapNodes(nodes []v1.Node, f func(v1.Node) string) (ret []string) {
	for _, node := range nodes {
		ret = append(ret, f(node))
	}
	return
}

func MapNodeNames(nodes []string, f func(string) v1.Node) (ret []v1.Node) {
	for _, node := range nodes {
		ret = append(ret, f(node))
	}
	return
}

func RetrieveInternalIP(node v1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == "InternalIP" {
			return addr.Address
		}
	}

	return ""
}

func WindowsOS(node v1.Node) bool {
	return node.GetLabels()["kubernetes.io/os"] == "windows"
}

/* Pod Methods */
