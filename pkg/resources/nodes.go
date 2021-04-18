package resources

import (
	"context"
	"regexp"
	"sort"

	"github.com/ynqa/widgets/pkg/table/node"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"

	"github.com/ynqa/ktop/pkg/formats"
)

type NodeResources map[string]*NodeResource

func (r NodeResources) GetNodeResource(node string) (*NodeResource, bool) {
	res, ok := r[node]
	return res, ok
}

func (r NodeResources) GetTree() *node.Node {
	tree := node.Root()
	for _, nd := range r.sortedNodes() {
		usage := r[nd].Usage
		cursorNode := node.New(nd, []string{
			nd,
			"",
			formats.FormatResourceString(corev1.ResourceCPU, usage),
			formats.FormatResourceString(corev1.ResourceMemory, usage),
		})
		tree.Append(cursorNode)

		for _, pd := range r.sortedPods(nd) {
			usage := r[nd].Pods[pd].Usage
			cursorPod := node.New(pd, []string{
				pd,
				r[nd].Pods[pd].Namespace,
				formats.FormatResourceString(corev1.ResourceCPU, usage),
				formats.FormatResourceString(corev1.ResourceMemory, usage),
			})
			cursorNode.Append(cursorPod)

			for _, ct := range r.sortedContainers(nd, pd) {
				usage = r[nd].Pods[pd].Containers[ct].Usage
				cursorPod.Append(node.New(ct, []string{
					ct,
					"",
					formats.FormatResourceString(corev1.ResourceCPU, usage),
					formats.FormatResourceString(corev1.ResourceMemory, usage),
				}))
			}
		}
	}
	return tree
}

func (r NodeResources) Len() int {
	return len(r)
}

func (r NodeResources) sortedNodes() []string {
	var res []string
	for k := range r {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func (r NodeResources) sortedPods(node string) []string {
	var res []string
	for k := range r[node].Pods {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func (r NodeResources) sortedContainers(node, pod string) []string {
	var res []string
	for k := range r[node].Pods[pod].Containers {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func FetchResources(
	namespace string,
	clientset *kubernetes.Clientset,
	metricsclientset *versioned.Clientset,
	nodeQuery, podQuery, containerQuery *regexp.Regexp,
) (Resources, error) {

	data := make(NodeResources)

	// get nodes and their metrics
	nodes, err := clientset.CoreV1().Nodes().List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: labels.Everything().String(),
		},
	)
	if err != nil {
		return nil, err
	}
	nodeMetrics, err := metricsclientset.MetricsV1beta1().
		NodeMetricses().List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: labels.Everything().String(),
		},
	)
	if err != nil {
		return nil, err
	}

	nodeMetrics.Items = matchNodes(nodeQuery, nodeMetrics.Items)
	for _, metric := range nodeMetrics.Items {
		nodeStatus := getNodeStatus(metric.Name, nodes.Items)
		data[metric.Name] = &NodeResource{
			Pods:        make(map[string]*PodResource),
			Capacity:    nodeStatus.Capacity,
			Allocatable: nodeStatus.Allocatable,
			Usage:       metric.Usage.DeepCopy(),
		}
	}

	// get pods and their metrics
	pods, err := clientset.CoreV1().Pods(namespace).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: labels.Everything().String(),
		},
	)
	if err != nil {
		return nil, err
	}
	podMetrics, err := metricsclientset.MetricsV1beta1().
		PodMetricses(namespace).List(
		context.Background(),
		metav1.ListOptions{
			LabelSelector: labels.Everything().String(),
		},
	)
	if err != nil {
		return nil, err
	}

	podMetrics.Items = matchPods(podQuery, podMetrics.Items)
	for _, podMetric := range podMetrics.Items {
		node := getAssignedNode(podMetric.Name, pods.Items)
		// it sometimes cannot get nodes because of the filters by `matchNodes`.
		if _, ok := data[node]; !ok {
			continue
		}

		// investigate all pods (without filtering) to aggregate the usages of their own containers.
		var cpuperpod, memoryperpod resource.Quantity
		for _, containerMetric := range podMetric.Containers {
			cpuperpod.Add(containerMetric.Usage.Cpu().DeepCopy())
			memoryperpod.Add(containerMetric.Usage.Memory().DeepCopy())
		}
		data[node].Pods[podMetric.Name] = &PodResource{
			Namespace:  podMetric.Namespace,
			Containers: make(map[string]*ContainerResource),
			Usage: corev1.ResourceList{
				corev1.ResourceCPU:    cpuperpod,
				corev1.ResourceMemory: memoryperpod,
			},
		}

		// but to search them by a given query is required on viewing.
		podMetric.Containers = matchContainers(containerQuery, podMetric.Containers)
		for _, containerMetric := range podMetric.Containers {
			data[node].Pods[podMetric.Name].Containers[containerMetric.Name] = &ContainerResource{
				Usage: containerMetric.Usage.DeepCopy(),
			}
		}
	}

	return data, nil
}
