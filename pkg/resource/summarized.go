package resource

import (
	corev1 "k8s.io/api/core/v1"

	. "github.com/ynqa/ktop/pkg/util"
)

type SummarizedResource struct {
	podName  string
	nodeName string
	usage    corev1.ResourceList
}

func NewSummarizedResource(p corev1.Pod, sumUsage corev1.ResourceList) *SummarizedResource {
	return &SummarizedResource{
		podName:  p.Name,
		nodeName: p.Spec.NodeName,
		usage:    sumUsage,
	}
}

func (s *SummarizedResource) GetNodeName() string {
	return s.nodeName
}

func (s *SummarizedResource) GetPodName() string {
	return s.podName
}

func (s *SummarizedResource) GetCpuUsage() (float64, string) {
	return GetResourceValue(s.usage, corev1.ResourceCPU),
		GetResourceValueString(s.usage, corev1.ResourceCPU)
}

func (s *SummarizedResource) GetMemoryUsage() (float64, string) {
	return GetResourceValue(s.usage, corev1.ResourceMemory),
		GetResourceValueString(s.usage, corev1.ResourceMemory)
}

// header: "POD", "%CPU", "%MEM"
func (s *SummarizedResource) toRow() []string {
	return []string{
		s.podName,
		GetResourceValueString(s.usage, corev1.ResourceCPU),
		GetResourceValueString(s.usage, corev1.ResourceMemory),
	}
}
