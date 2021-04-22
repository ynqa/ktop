package kube

import (
	"io"

	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/kubectl/metricsutil"
	"k8s.io/metrics/pkg/apis/metrics"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type KubeClients struct {
	Flags         *genericclioptions.ConfigFlags
	clientset     *kubernetes.Clientset
	metricsClient metricsClient
}

func NewKubeClients(flags *genericclioptions.ConfigFlags) (*KubeClients, error) {
	config, err := flags.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	var metricsClient metricsClient
	mergedErr := errors.New("Failed to create metrics client")
	metricsClient, err = newMetricsServerClient(config)
	if err != nil {
		mergedErr = errors.Wrap(mergedErr, err.Error())
		metricsClient, err = newHeapsterClient(clientset.CoreV1())
		if err != nil {
			mergedErr = errors.Wrap(mergedErr, err.Error())
			return nil, mergedErr
		}
	}
	return &KubeClients{
		Flags:         flags,
		clientset:     clientset,
		metricsClient: metricsClient,
	}, nil
}

func (k *KubeClients) GetPodList(namespace string, labelSelector labels.Selector) (*corev1.PodList, error) {
	return k.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector.String()})
}

func (k *KubeClients) GetPodLogs(namespace string, podName string) (string, error) {
	count := int64(100)
	follow := false
	message := ""
	podLogOptions := corev1.PodLogOptions{
		//Container: containerName,
		Follow:    follow,
		TailLines: &count,
	}
	podLogRequest := k.clientset.CoreV1().
		Pods(namespace).
		GetLogs(podName, &podLogOptions)
	stream, err := podLogRequest.Stream()
	if err != nil {
		return "failure1", err
	}
	defer stream.Close()
	for {
		buf := make([]byte, 2000)
		numBytes, err := stream.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "failure2", err
		}
		if numBytes == 0 {
			continue
		}
		message += string(buf[:numBytes])
	}
	return message, nil
}

func (k *KubeClients) GetPodMetricsList(namespace string, labelSelector labels.Selector) (*metrics.PodMetricsList, error) {
	return k.metricsClient.getPodMetricsList(namespace, labelSelector)
}

func (k *KubeClients) GetNodeList(labelSelector labels.Selector) (*corev1.NodeList, error) {
	return k.clientset.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: labelSelector.String()})
}

func (k *KubeClients) GetNodeMetricsList(labelSelector labels.Selector) (*metrics.NodeMetricsList, error) {
	return k.metricsClient.getNodeMetricsList(labelSelector)
}

type metricsClient interface {
	getPodMetricsList(namespace string, labelSelector labels.Selector) (*metrics.PodMetricsList, error)
	getNodeMetricsList(labelSelector labels.Selector) (*metrics.NodeMetricsList, error)
}

type metricsServerClient struct {
	*versioned.Clientset
}

func newMetricsServerClient(config *rest.Config) (*metricsServerClient, error) {
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &metricsServerClient{
		Clientset: clientset,
	}, nil
}

func (c *metricsServerClient) getPodMetricsList(namespace string, labelSelector labels.Selector) (*metrics.PodMetricsList, error) {
	list, err := c.MetricsV1beta1().PodMetricses(namespace).List(metav1.ListOptions{LabelSelector: labelSelector.String()})
	if err != nil {
		return nil, err
	}
	old := &metrics.PodMetricsList{}
	if err := metricsv1beta1.Convert_v1beta1_PodMetricsList_To_metrics_PodMetricsList(list, old, nil); err != nil {
		return nil, err
	}
	return old, nil
}

func (c *metricsServerClient) getNodeMetricsList(labelSelector labels.Selector) (*metrics.NodeMetricsList, error) {
	list, err := c.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{LabelSelector: labelSelector.String()})
	if err != nil {
		return nil, err
	}
	old := &metrics.NodeMetricsList{}
	if err := metricsv1beta1.Convert_v1beta1_NodeMetricsList_To_metrics_NodeMetricsList(list, old, nil); err != nil {
		return nil, err
	}
	return old, nil
}

type heapsterClient struct {
	*metricsutil.HeapsterMetricsClient
}

func newHeapsterClient(svcClient corev1client.ServicesGetter) (*heapsterClient, error) {
	heapster := metricsutil.NewHeapsterMetricsClient(
		svcClient,
		metricsutil.DefaultHeapsterNamespace,
		metricsutil.DefaultHeapsterScheme,
		metricsutil.DefaultHeapsterService,
		metricsutil.DefaultHeapsterPort,
	)
	return &heapsterClient{
		HeapsterMetricsClient: heapster,
	}, nil
}

func (c *heapsterClient) getPodMetricsList(namespace string, labelSelector labels.Selector) (*metrics.PodMetricsList, error) {
	return c.GetPodMetrics(namespace, "", false, labelSelector)
}

func (c *heapsterClient) getNodeMetricsList(labelSelector labels.Selector) (*metrics.NodeMetricsList, error) {
	return c.GetNodeMetrics("", labelSelector.String())
}
