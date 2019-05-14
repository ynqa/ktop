package resource

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics"

	. "github.com/ynqa/ktop/pkg/util"
)

type Resource struct {
	nodeName      string
	podName       string
	containerName string
	usage         corev1.ResourceList
	limits        corev1.ResourceList
	requests      corev1.ResourceList
}

func NewResource(p corev1.Pod, c corev1.Container, cm metrics.ContainerMetrics) *Resource {
	return &Resource{
		nodeName:      p.Spec.NodeName,
		podName:       p.Name,
		containerName: c.Name,
		usage:         cm.Usage,
		limits:        c.Resources.Limits,
		requests:      c.Resources.Requests,
	}
}

func (r *Resource) GetNodeName() string {
	return r.nodeName
}

func (r *Resource) GetContainerName() string {
	return r.containerName
}

func (r *Resource) GetCpuLimits() (float64, string, bool) {
	_, ok := r.limits[corev1.ResourceCPU]
	str := GetResourceValueString(r.limits, corev1.ResourceCPU)
	return GetResourceValue(r.limits, corev1.ResourceCPU), str, ok
}

func (r *Resource) GetCpuUsage() (float64, string) {
	return GetResourceValue(r.usage, corev1.ResourceCPU),
		GetResourceValueString(r.usage, corev1.ResourceCPU)
}

func (r *Resource) GetMemoryLimits() (float64, string, bool) {
	_, ok := r.limits[corev1.ResourceMemory]
	str := GetResourceValueString(r.limits, corev1.ResourceMemory)
	return GetResourceValue(r.limits, corev1.ResourceMemory), str, ok
}

func (r *Resource) GetMemoryUsage() (float64, string) {
	return GetResourceValue(r.usage, corev1.ResourceMemory),
		GetResourceValueString(r.usage, corev1.ResourceMemory)
}

// header: "POD", "CONTAINER", "CPU(U)", "CPU(L)", "CPU(R)", "Mem(U)", "Mem(L)", "Mem(R)"
func (r *Resource) toRow() []string {
	return []string{
		r.podName,
		r.containerName,
		GetResourceValueString(r.usage, corev1.ResourceCPU),
		GetResourceValueString(r.limits, corev1.ResourceCPU),
		GetResourceValueString(r.requests, corev1.ResourceCPU),
		GetResourceValueString(r.usage, corev1.ResourceMemory),
		GetResourceValueString(r.limits, corev1.ResourceMemory),
		GetResourceValueString(r.requests, corev1.ResourceMemory),
	}
}
