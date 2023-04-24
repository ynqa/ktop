package ktop

import (
	"container/ring"
	"fmt"
	"regexp"
	"sync"

	"github.com/gizak/termui/v3"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	kr "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/ynqa/ktop/pkg/kube"
	"github.com/ynqa/ktop/pkg/resource"
	"github.com/ynqa/ktop/pkg/ui"
	. "github.com/ynqa/ktop/pkg/util"
)

var (
	// style
	titleStyle = termui.NewStyle(termui.ColorWhite, termui.ColorClear, termui.ModifierBold)
)

const (
	// label names
	containerLimitLabel  = "ContainerLimits"
	nodeAllocatableLabel = "NodeAllocatable"

	// colors
	borderColor         = termui.ColorBlue
	selectedTableColor  = termui.ColorYellow
	graphLabelNameColor = termui.ColorWhite
	graphLimitColor     = termui.ColorWhite
	graphDataColor      = termui.ColorGreen
)

type Monitor struct {
	*kube.KubeClients

	logs            *ui.Paragraph
	table           *ui.Table
	tableTypeCircle *ring.Ring

	cpuGraph *ui.Graph
	memGraph *ui.Graph

	podQuery       *regexp.Regexp
	containerQuery *regexp.Regexp
	nodeQuery      *regexp.Regexp
}

func NewMonitor(kubeclients *kube.KubeClients, podQuery, containerQuery, nodeQuery *regexp.Regexp) *Monitor {
	monitor := &Monitor{
		KubeClients:     kubeclients,
		tableTypeCircle: resource.TableTypeCircle(),
		podQuery:        podQuery,
		containerQuery:  containerQuery,
		nodeQuery:       nodeQuery,
	}

	// table for resources
	table := ui.NewTable()
	table.TitleStyle = titleStyle
	table.Cursor = true
	table.BorderStyle = termui.NewStyle(borderColor)
	table.CursorColor = selectedTableColor

	// logs of pod
	logs := ui.NewParagraph()
	logs.Title = "⎈ Logs ⎈"
	logs.TitleStyle = titleStyle
	logs.Text = `Loading...`
	logs.BorderStyle = termui.NewStyle(borderColor)
	logs.TextStyle = termui.NewStyle(termui.Color(244), termui.ColorClear)

	// graph for cpu
	cpu := ui.NewGraph()
	cpu.Title = "⎈ CPU Usage ⎈"
	cpu.TitleStyle = titleStyle
	cpu.BorderStyle = termui.NewStyle(borderColor)
	cpu.LabelNameColor = graphLabelNameColor
	cpu.DataColor = graphDataColor
	cpu.LimitColor = graphLimitColor

	// graph for memory
	mem := ui.NewGraph()
	mem.Title = "⎈ Memory Usage ⎈"
	mem.TitleStyle = titleStyle
	mem.BorderStyle = termui.NewStyle(borderColor)
	mem.LabelNameColor = graphLabelNameColor
	mem.DataColor = graphDataColor
	mem.LimitColor = graphLimitColor

	monitor.table = table
	monitor.logs = logs
	monitor.cpuGraph = cpu
	monitor.memGraph = mem
	return monitor
}

func (m *Monitor) resetGraph() {
	m.cpuGraph.Reset()
	m.memGraph.Reset()
}

func (m *Monitor) resetTable() {
	m.table.Reset(resource.ResetTableShapeFrom(
		m.tableTypeCircle.Value.(string),
		m.table.Inner,
	))
}

func (m *Monitor) ScrollDown() {
	m.scrollDown()
	m.resetGraph()
}

func (m *Monitor) scrollDown() {
	m.table.ScrollDown()
}

func (m *Monitor) ScrollUp() {
	m.scrollUp()
	m.resetGraph()
}

func (m *Monitor) scrollUp() {
	m.table.ScrollUp()
}

func (m *Monitor) Rotate() {
	m.rotate(1)
	m.resetGraph()
	m.resetTable()
}

func (m *Monitor) ReverseRotate() {
	m.rotate(-1)
	m.resetGraph()
	m.resetTable()
}

func (m *Monitor) rotate(i int) {
	m.tableTypeCircle = m.tableTypeCircle.Move(i)
}

func (m *Monitor) GetCPUGraph() *ui.Graph {
	return m.cpuGraph
}

func (m *Monitor) GetMemGraph() *ui.Graph {
	return m.memGraph
}

func (m *Monitor) GetPodTable() *ui.Table {
	return m.table
}

func (m *Monitor) GetLogs() *ui.Paragraph {
	return m.logs
}

func (m *Monitor) Update() error {
	nodeList, err := m.GetNodeList(labels.Everything())
	if err != nil {
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	resourcesCh := make(chan []*resource.Resource, 1)
	summarizedResourcesCh := make(chan []*resource.SummarizedResource, 1)
	nodeResourcesCh := make(chan []*resource.NodeResource, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		resources, summarizedResources, err := m.fetchPodResources()
		if err != nil {
			errCh <- err
			return
		}
		resourcesCh <- resources
		summarizedResourcesCh <- summarizedResources
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		nodeResources, err := m.fetchNodeResources(nodeList)
		if err != nil {
			errCh <- err
			return
		}
		nodeResourcesCh <- nodeResources
	}()

	go func() {
		wg.Wait()
		close(errCh)
		close(resourcesCh)
		close(summarizedResourcesCh)
		close(nodeResourcesCh)
	}()

	var mergedError error
	for err := range errCh {
		if mergedError == nil {
			mergedError = errors.New(err.Error())
		}
		mergedError = errors.Wrap(mergedError, err.Error())
	}
	if mergedError != nil {
		return mergedError
	}

	resources, ok := <-resourcesCh
	if !ok {
		return errors.New("Failed to get resources")
	}

	summarizedResources, ok := <-summarizedResourcesCh
	if !ok {
		return errors.New("Failed to get summarized resources")
	}

	nodeResources, ok := <-nodeResourcesCh
	if !ok {
		return errors.New("Failed to get node resources")
	}

	// temporary
	defer func() {
		if p := recover(); p != nil {
			m.table.SelectedRow = 0
		}
	}()

	switch m.tableTypeCircle.Value.(string) {
	case resource.SummarizedType:
		summarizedViewer := resource.AsSummarizedTableViewer(summarizedResources, resource.ByName)
		summarizedViewer.SortRows()
		m.updatePodTable(summarizedViewer)
		if len(summarizedResources) > 0 {
			current := summarizedResources[m.table.SelectedRow]
			if err := m.updateSummarizedGraph(nodeList, current); err != nil {
				return err
			}
		}
	case resource.AllType:
		viewer := resource.AsAllTableViewer(resources, resource.ByName)
		viewer.SortRows()
		m.updatePodTable(viewer)
		if len(resources) > 0 {
			current := resources[m.table.SelectedRow]
			if err := m.updateAllGraph(nodeList, current); err != nil {
				return err
			}
		}
	case resource.NodeType:
		nodeViewer := resource.AsNodeTableViewer(nodeResources, resource.ByName)
		nodeViewer.SortRows()
		m.updatePodTable(nodeViewer)
		if len(nodeResources) > 0 {
			current := nodeResources[m.table.SelectedRow]
			if err := m.updateNodeGraph(current); err != nil {
				return err
			}
		}
	default:
	}

	return nil
}

func (m *Monitor) fetchPodResources() ([]*resource.Resource, []*resource.SummarizedResource, error) {
	podMetricsList, err := m.GetPodMetricsList(*m.Flags.Namespace, labels.Everything())
	if err != nil {
		return nil, nil, err
	}
	podList, err := m.GetPodList(*m.Flags.Namespace, labels.Everything())
	if err != nil {
		return nil, nil, err
	}

	// collect resource list
	resources := make([]*resource.Resource, 0)
	summarizedResources := make([]*resource.SummarizedResource, 0)
	// filtered
	for _, podMetrics := range FilterPodMetrics(m.podQuery, podMetricsList.Items) {
		podName := podMetrics.Name
		podLogs, err := m.GetPodLogs(*m.Flags.Namespace, podName)
		if err != nil {
			fmt.Print(err)
		}
		pod := FindPod(podName, podList.Items)
		if pod == nil {
			continue
		}
		var cpu, mem kr.Quantity
		// filtered
		for _, containerMetrics := range FilterContainerMetrics(m.containerQuery, podMetrics.Containers) {
			container := FindContainer(containerMetrics.Name, pod.Spec.Containers)
			if container == nil {
				continue
			}
			containerResource := resource.NewResource(*pod, *container, containerMetrics)
			resources = append(resources, containerResource)
			cpu.Add(*containerMetrics.Usage.Cpu())
			mem.Add(*containerMetrics.Usage.Memory())
		}
		summarizedResource := resource.NewSummarizedResource(*pod,
			corev1.ResourceList{
				corev1.ResourceCPU:    cpu,
				corev1.ResourceMemory: mem,
			}, podLogs)
		summarizedResources = append(summarizedResources, summarizedResource)
	}
	return resources, summarizedResources, nil
}

func (m *Monitor) fetchNodeResources(nodeList *corev1.NodeList) ([]*resource.NodeResource, error) {
	nodeMetricsList, err := m.GetNodeMetricsList(labels.Everything())
	if err != nil {
		return nil, err
	}
	resources := make([]*resource.NodeResource, 0)
	// filtered
	for _, nodeMetrics := range FilterNodeMetrics(m.nodeQuery, nodeMetricsList.Items) {
		node := FindNode(nodeMetrics.Name, nodeList.Items)
		resources = append(resources, resource.NewNodeResource(*node, nodeMetrics))
	}
	return resources, nil
}

func (m *Monitor) updatePodTable(resources resource.ResourceTableViewer) {
	m.table.Title, m.table.Header, m.table.ColumnWidths, m.table.Rows = resources.GetTableShape(m.table.Inner)
}

func (m *Monitor) updateSummarizedGraph(nodeList *corev1.NodeList, summarized *resource.SummarizedResource) error {
	cpuUsage, cpuUsageStr := summarized.GetCpuUsage()
	memUsage, memUsageStr := summarized.GetMemoryUsage()
	node := FindNode(summarized.GetNodeName(), nodeList.Items)
	limitCpu := GetResourceValue(node.Status.Allocatable, corev1.ResourceCPU)
	limitCpuStr := GetResourceValueString(node.Status.Allocatable, corev1.ResourceCPU)
	limitMemory := GetResourceValue(node.Status.Allocatable, corev1.ResourceMemory)
	limitMemoryStr := GetResourceValueString(node.Status.Allocatable, corev1.ResourceMemory)

	m.logs.Text = summarized.GetLogs()

	m.cpuGraph.LabelHeader = fmt.Sprintf("Name: %v", summarized.GetPodName())
	m.cpuGraph.Data = append(m.cpuGraph.Data, cpuUsage)
	m.cpuGraph.LabelData = fmt.Sprintf("Usage: %v", cpuUsageStr)
	m.cpuGraph.UpperLimit = limitCpu
	m.cpuGraph.DrawUpperLimit = false
	m.cpuGraph.LabelUpperLimit = fmt.Sprintf("%v: %v", nodeAllocatableLabel, limitCpuStr)

	m.memGraph.LabelHeader = fmt.Sprintf("Name: %v", summarized.GetPodName())
	m.memGraph.Data = append(m.memGraph.Data, memUsage)
	m.memGraph.LabelData = fmt.Sprintf("Usage: %v", memUsageStr)
	m.memGraph.UpperLimit = limitMemory
	m.memGraph.DrawUpperLimit = false
	m.memGraph.LabelUpperLimit = fmt.Sprintf("%v: %v", nodeAllocatableLabel, limitMemoryStr)
	return nil
}

func (m *Monitor) updateAllGraph(nodeList *corev1.NodeList, all *resource.Resource) error {
	cpuUsage, cpuUsageStr := all.GetCpuUsage()
	memUsage, memUsageStr := all.GetMemoryUsage()

	limitCpuLabel := containerLimitLabel
	limitCpu, limitCpuStr, cok := all.GetCpuLimits()
	limitMemoryLabel := containerLimitLabel
	limitMemory, limitMemoryStr, mok := all.GetMemoryLimits()

	var node *corev1.Node
	if !cok || !mok {
		node = FindNode(all.GetNodeName(), nodeList.Items)
	}
	if !cok {
		limitCpuLabel = nodeAllocatableLabel
		limitCpu = GetResourceValue(node.Status.Allocatable, corev1.ResourceCPU)
		limitCpuStr = GetResourceValueString(node.Status.Allocatable, corev1.ResourceCPU)
	}
	if !mok {
		limitMemoryLabel = nodeAllocatableLabel
		limitMemory = GetResourceValue(node.Status.Allocatable, corev1.ResourceMemory)
		limitMemoryStr = GetResourceValueString(node.Status.Allocatable, corev1.ResourceMemory)
	}

	m.cpuGraph.LabelHeader = fmt.Sprintf("Name: %v", all.GetContainerName())
	m.cpuGraph.Data = append(m.cpuGraph.Data, cpuUsage)
	m.cpuGraph.LabelData = fmt.Sprintf("Usage: %v", cpuUsageStr)
	m.cpuGraph.UpperLimit = limitCpu
	m.cpuGraph.DrawUpperLimit = false
	m.cpuGraph.LabelUpperLimit = fmt.Sprintf("%v: %v", limitCpuLabel, limitCpuStr)

	m.memGraph.LabelHeader = fmt.Sprintf("Name: %v", all.GetContainerName())
	m.memGraph.Data = append(m.memGraph.Data, memUsage)
	m.memGraph.LabelData = fmt.Sprintf("Usage: %v", memUsageStr)
	m.memGraph.UpperLimit = limitMemory
	m.memGraph.DrawUpperLimit = false
	m.memGraph.LabelUpperLimit = fmt.Sprintf("%v: %v", limitMemoryLabel, limitMemoryStr)
	return nil
}

func (m *Monitor) updateNodeGraph(node *resource.NodeResource) error {
	cpuUsage, cpuUsageStr := node.GetCpuUsagePercentage()
	memUsage, memUsageStr := node.GetMemoryUsagePercentage()

	m.cpuGraph.LabelHeader = fmt.Sprintf("Name: %v", node.GetNodeName())
	m.cpuGraph.Data = append(m.cpuGraph.Data, cpuUsage)
	m.cpuGraph.UpperLimit = 100.
	m.cpuGraph.DrawUpperLimit = false
	m.cpuGraph.LabelData = fmt.Sprintf("%%Usage: %v", cpuUsageStr)

	m.memGraph.LabelHeader = fmt.Sprintf("Name: %v", node.GetNodeName())
	m.memGraph.Data = append(m.memGraph.Data, memUsage)
	m.memGraph.UpperLimit = 100.
	m.memGraph.DrawUpperLimit = false
	m.memGraph.LabelData = fmt.Sprintf("%%Usage: %v", memUsageStr)
	return nil
}
