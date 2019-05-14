package resource

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics"

	. "github.com/ynqa/ktop/pkg/util"
)

type NodeResource struct {
	nodeName    string
	capacity    corev1.ResourceList
	allocatable corev1.ResourceList
	usage       corev1.ResourceList
}

func NewNodeResource(n corev1.Node, nm metrics.NodeMetrics) *NodeResource {
	return &NodeResource{
		nodeName:    nm.Name,
		capacity:    n.Status.Capacity,
		allocatable: n.Status.Allocatable,
		usage:       nm.Usage,
	}
}

func (r *NodeResource) GetNodeName() string {
	return r.nodeName
}

func (r *NodeResource) GetCpuUsagePercentage() (float64, string) {
	return GetResourcePercentage(*r.usage.Cpu(), *r.allocatable.Cpu()),
		GetResourcePercentageString(*r.usage.Cpu(), *r.allocatable.Cpu())
}

func (r *NodeResource) GetMemoryUsagePercentage() (float64, string) {
	return GetResourcePercentage(*r.usage.Memory(), *r.allocatable.Memory()),
		GetResourcePercentageString(*r.usage.Memory(), *r.allocatable.Memory())
}

// header: "NODE", "CPU(C)", "CPU(A)", "CPU(U)", "%CPU", "Memory(C)", "Memory(A)", "Memory(U)", "%Memory",
func (r *NodeResource) toRow() []string {
	return []string{
		r.nodeName,
		GetResourceValueString(r.allocatable, corev1.ResourceCPU),
		GetResourceValueString(r.usage, corev1.ResourceCPU),
		GetResourcePercentageString(*r.usage.Cpu(), *r.allocatable.Cpu()),
		GetResourceValueString(r.allocatable, corev1.ResourceMemory),
		GetResourceValueString(r.usage, corev1.ResourceMemory),
		GetResourcePercentageString(*r.usage.Memory(), *r.allocatable.Memory()),
	}
}
