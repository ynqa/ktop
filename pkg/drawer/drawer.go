package drawer

import (
	"fmt"
	"image"

	"github.com/gizak/termui/v3"
	"github.com/ynqa/widgets/pkg/table"
	"github.com/ynqa/widgets/pkg/table/node"
	corev1 "k8s.io/api/core/v1"

	"github.com/ynqa/ktop/pkg/formats"
	"github.com/ynqa/ktop/pkg/resources"
	"github.com/ynqa/ktop/pkg/ui"
)

type TableDrawer interface {
	Draw(*table.Table, resources.Resources)
}

type NopTableDrawer struct{}

func (d *NopTableDrawer) Draw(table *table.Table, r resources.Resources) {
	table.SetHeaders([]string{"message"})
	table.SetWidthFn(func(_ []string, rect image.Rectangle) []int {
		return []int{rect.Dx() - 1}
	})
	table.SetNode(node.New("", []string{"not found: nodes, pods, and containers"}))
}

type KubeTableDrawer struct{}

func (d *KubeTableDrawer) Draw(table *table.Table, r resources.Resources) {
	table.SetHeaders([]string{"name", "namespace", "usage.cpu", "usage.memory"})
	table.SetWidthFn(func(headers []string, rect image.Rectangle) []int {
		widths := []int{rect.Dx() / 2}
		denom := 2 * (len(headers) - 1)
		for i := 1; i < len(headers); i++ {
			widths = append(widths, rect.Dx()/denom)
		}
		return widths
	},
	)
	table.SetNode(node.ApplyChildVisible(table.GetNode(), r.GetTree()))
}

func DrawGraph(g *ui.Graph, r resources.Resources, typ corev1.ResourceName, nodes []*node.Node) {
	if len(nodes) == 1 {
		node, ok := r.GetNodeResource(nodes[0].Name())
		// fmt.Sprintf("usage (%v) / allocatable (%v) = %v",
		// 	formats.FormatResourceString(typ, node.Usage),
		// 	formats.FormatResourceString(typ, node.Allocatable),
		// 	formats.FormatResourcePercentage(typ, node.Usage, node.Allocatable),
		// )
		if ok {
			g.Update([]ui.Data{
				{
					Type:  ui.BaseLine,
					Value: float64(formats.FormatResource(typ, node.Allocatable)),
					Label: fmt.Sprintf("allocatable: %v", formats.FormatResourceString(typ, node.Allocatable)),
					Style: termui.NewStyle(termui.ColorRed, termui.ColorClear),
				},
				{
					Type:  ui.TimeSeries,
					Value: float64(formats.FormatResource(typ, node.Usage)),
					Label: fmt.Sprintf("usage: %v", formats.FormatResourceString(typ, node.Usage)),
					Style: termui.NewStyle(termui.ColorClear, termui.ColorClear),
				},
			})
		}
	}
	// else if len(nodes) == 2 {
	// 	node, ok := r[nodes[1].Name()]
	// 	if ok {
	// 		pod, ok := node.Pods[nodes[0].Name()]
	// 		if ok {
	// 			g.UpperLimit = float64(formats.FormatResource(typ, node.Allocatable))
	// 			g.Data = append(g.Data, float64(formats.FormatResource(typ, pod.Usage)))
	// 			g.LabelData = fmt.Sprintf(labelTmpl,
	// 				formats.FormatResourceString(typ, pod.Usage),
	// 				formats.FormatResourceString(typ, node.Allocatable),
	// 				formats.FormatResourcePercentage(typ, pod.Usage, node.Allocatable),
	// 			)
	// 		}
	// 	}
	// } else if len(nodes) == 3 {
	// 	node, ok := r[nodes[2].Name()]
	// 	if ok {
	// 		pod, ok := node.Pods[nodes[1].Name()]
	// 		if ok {
	// 			container, ok := pod.Containers[nodes[0].Name()]
	// 			g.Data = append(g.Data, float64(formats.FormatResource(typ, container.Usage)))
	// 			if ok {
	// 				g.UpperLimit = float64(formats.FormatResource(typ, node.Allocatable))
	// 				g.LabelData = fmt.Sprintf(labelTmpl,
	// 					formats.FormatResourceString(typ, container.Usage),
	// 					formats.FormatResourceString(typ, node.Allocatable),
	// 					formats.FormatResourcePercentage(typ, container.Usage, node.Allocatable),
	// 				)
	// 			}
	// 		}
	// 	}
	// }
}
