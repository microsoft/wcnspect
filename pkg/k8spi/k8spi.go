package k8spi

import (
	v1 "k8s.io/api/core/v1"
)

// Node Methods
func FilterNodes(nodes []v1.Node, test func(v1.Node) bool) (ret []v1.Node) {
	for _, node := range nodes {
		if test(node) {
			ret = append(ret, node)
		}
	}
	return
}

/* Retrieves node names and ips given a list of nodes
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

// Pod Methods
func GetPodMaps(pods []v1.Pod) (map[string]string, map[string]string) {
	ips := map[string]string{}
	nodes := map[string]string{}
	for _, pod := range pods {
		ips[pod.GetName()] = pod.Status.PodIP
		nodes[pod.GetName()] = pod.Status.HostIP
	}

	return ips, nodes
}

func MapPods(pods []v1.Pod, f func(v1.Pod) string) (ret []string) {
	for _, pod := range pods {
		ret = append(ret, f(pod))
	}
	return
}
