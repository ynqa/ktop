package resources

import (
	"regexp"

	"github.com/ynqa/widgets/pkg/table/node"
	corev1 "k8s.io/api/core/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type Resources interface {
	GetNodeResource(string) (*NodeResource, bool)
	GetTree() *node.Node
	Len() int
}

// - https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/#node-allocatable
// - https://github.com/kubernetes/kubectl/blob/v0.18.8/pkg/describe/describe.go#L3521-L3602
type NodeResource struct {
	Pods        map[string]*PodResource
	Capacity    corev1.ResourceList
	Allocatable corev1.ResourceList
	Usage       corev1.ResourceList
}

type PodResource struct {
	Namespace  string
	Containers map[string]*ContainerResource
	Usage      corev1.ResourceList
}

type ContainerResource struct {
	Usage corev1.ResourceList
}

func matchNodes(query *regexp.Regexp, nodes []metricsv1beta1.NodeMetrics) []metricsv1beta1.NodeMetrics {
	var res []metricsv1beta1.NodeMetrics
	for _, node := range nodes {
		if query.MatchString(node.Name) {
			res = append(res, node)
		}
	}
	return res
}

func getNodeStatus(target string, nodes []corev1.Node) corev1.NodeStatus {
	for _, node := range nodes {
		if target == node.Name {
			return node.Status
		}
	}
	return corev1.NodeStatus{}
}

func matchPods(query *regexp.Regexp, pods []metricsv1beta1.PodMetrics) []metricsv1beta1.PodMetrics {
	var res []metricsv1beta1.PodMetrics
	for _, pod := range pods {
		if query.MatchString(pod.Name) {
			res = append(res, pod)
		}
	}
	return res
}

func getAssignedNode(target string, pods []corev1.Pod) string {
	for _, pod := range pods {
		if target == pod.Name {
			return pod.Spec.NodeName
		}
	}
	return ""
}

func matchContainers(query *regexp.Regexp, containers []metricsv1beta1.ContainerMetrics) []metricsv1beta1.ContainerMetrics {
	var res []metricsv1beta1.ContainerMetrics
	for _, container := range containers {
		if query.MatchString(container.Name) {
			res = append(res, container)
		}
	}
	return res
}
