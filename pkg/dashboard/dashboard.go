package dashboard

import (
	"sync"

	"github.com/gizak/termui/v3"
	"github.com/ynqa/widgets/pkg/table"
	corev1 "k8s.io/api/core/v1"

	"github.com/ynqa/ktop/pkg/drawer"
	"github.com/ynqa/ktop/pkg/resources"
	"github.com/ynqa/ktop/pkg/ui"
)

type Dashboard struct {
	mu                    sync.RWMutex
	table                 *table.Table
	cpuGraph, memoryGraph *ui.Graph
}

func New() *Dashboard {
	return &Dashboard{
		table:       newTable("RESOURCES"),
		cpuGraph:    newGraph("CPU"),
		memoryGraph: newGraph("MEMORY"),
	}
}

func newTable(title string) *table.Table {
	block := termui.NewBlock()
	block.Title = title
	block.TitleStyle = termui.NewStyle(termui.ColorClear)
	block.BorderStyle = termui.NewStyle(termui.ColorBlue)
	table := table.New(table.Block(block))
	return table
}

func newGraph(title string) *ui.Graph {
	graph := ui.NewGraph()
	graph.SetTitle(title, termui.NewStyle(termui.ColorClear))
	graph.SetGraphBorderStyle(termui.NewStyle(termui.ColorBlue))
	graph.SetLabelBorderStyle(termui.NewStyle(termui.ColorBlue))
	return graph
}

func (d *Dashboard) Table() *table.Table {
	return d.table
}

func (d *Dashboard) CPUGraph() *ui.Graph {
	return d.cpuGraph
}

func (d *Dashboard) MemoryGraph() *ui.Graph {
	return d.memoryGraph
}

func (d *Dashboard) Toggle() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.table.GetNode().Toggle(d.table.GetSelectedRow())
}

func (d *Dashboard) ScrollUp() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.table.ScrollUp()
	d.cpuGraph.Reset()
	d.memoryGraph.Reset()
}

func (d *Dashboard) ScrollDown() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.table.ScrollDown()
	d.cpuGraph.Reset()
	d.memoryGraph.Reset()
}

func (d *Dashboard) DrawTable(drawer drawer.TableDrawer, r resources.Resources) {
	d.mu.Lock()
	defer d.mu.Unlock()
	drawer.Draw(d.table, r)
}

func (d *Dashboard) DrawCPUGraph(r resources.Resources) {
	d.mu.Lock()
	defer d.mu.Unlock()
	stack := d.table.GetNode().Flatten()
	if d.table.GetSelectedRow() < len(stack) {
		drawer.DrawGraph(d.cpuGraph, r, corev1.ResourceCPU, stack[d.table.GetSelectedRow()].Parents())
	}
}

func (d *Dashboard) DrawMemoryGraph(r resources.Resources) {
	d.mu.Lock()
	defer d.mu.Unlock()
	stack := d.table.GetNode().Flatten()
	if d.table.GetSelectedRow() < len(stack) {
		drawer.DrawGraph(d.memoryGraph, r, corev1.ResourceMemory, stack[d.table.GetSelectedRow()].Parents())
	}
}
