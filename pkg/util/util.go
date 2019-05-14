package util

import (
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/metrics/pkg/apis/metrics"
)

func FilterNodeMetrics(query *regexp.Regexp, nodes []metrics.NodeMetrics) []metrics.NodeMetrics {
	var filtered []metrics.NodeMetrics
	for _, node := range nodes {
		if query.MatchString(node.Name) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func FilterPodMetrics(query *regexp.Regexp, pods []metrics.PodMetrics) []metrics.PodMetrics {
	var filtered []metrics.PodMetrics
	for _, pod := range pods {
		if query.MatchString(pod.Name) {
			filtered = append(filtered, pod)
		}
	}
	return filtered
}

func FilterContainerMetrics(query *regexp.Regexp, containers []metrics.ContainerMetrics) []metrics.ContainerMetrics {
	var filtered []metrics.ContainerMetrics
	for _, container := range containers {
		if query.MatchString(container.Name) {
			filtered = append(filtered, container)
		}
	}
	return filtered
}

func FindNode(name string, nodes []corev1.Node) *corev1.Node {
	for _, node := range nodes {
		if name == node.Name {
			return &node
		}
	}
	return nil
}

func FindPod(name string, pods []corev1.Pod) *corev1.Pod {
	for _, pod := range pods {
		if name == pod.Name {
			return &pod
		}
	}
	return nil
}

func FindContainer(name string, containers []corev1.Container) *corev1.Container {
	for _, container := range containers {
		if name == container.Name {
			return &container
		}
	}
	return nil
}

func GetResourceValue(lst corev1.ResourceList, typ corev1.ResourceName) float64 {
	val, ok := lst[typ]
	switch {
	case typ == corev1.ResourceCPU && ok:
		return float64(val.MilliValue())
	case typ == corev1.ResourceMemory && ok:
		return float64(val.Value() / (1024 * 1024))
	}
	return 0
}

func GetResourceValueString(lst corev1.ResourceList, typ corev1.ResourceName) string {
	val, ok := lst[typ]
	switch {
	case typ == corev1.ResourceCPU && ok:
		return fmt.Sprintf("%vm", val.MilliValue())
	case typ == corev1.ResourceMemory && ok:
		return fmt.Sprintf("%vMi", val.Value()/(1024*1024))
	default:
		return "-"
	}
}

func GetResourcePercentage(usage, available resource.Quantity) float64 {
	return float64(usage.MilliValue()) / float64(available.MilliValue()) * 100
}

func GetResourcePercentageString(usage, available resource.Quantity) string {
	return fmt.Sprintf("%v%%", int(float64(usage.MilliValue())/float64(available.MilliValue())*100))
}

func IntMax(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func IntMin(x, y int) int {
	if x < y {
		return x
	}
	return y
}
