package resource

import (
	"container/ring"
	"image"
	"sort"

	. "github.com/ynqa/ktop/pkg/util"
)

type SortType int

const (
	ByName SortType = iota
)

type ResourceTableViewer interface {
	GetTableShape(rect image.Rectangle) (string, []string, []int, [][]string)
	SortRows()
}

var (
	SummarizedType = "Summarized"
	AllType        = "All"
	NodeType       = "Node"

	allTitle  = "⎈  Pod/Container ⎈ "
	allHeader = []string{
		"POD", "CONTAINER",
		"CPU(U)", "CPU(L)", "CPU(R)",
		"Memory(U)", "Memory(L)", "Memory(R)",
	}
	indentSize = 4
	allWidthFn = func(rect image.Rectangle, maxLen0, maxLen1 int) []int {
		podWidth := IntMax(40, IntMin(rect.Dx()-60, maxLen0+indentSize))
		containerWidth := IntMax(30, IntMin(rect.Dx()-60, maxLen1+indentSize))
		return []int{podWidth, containerWidth, 10, 10, 10, 10, 10, 10}
	}

	emptyHeader = []string{
		"Message",
	}
	emptyWidthFn = func(rect image.Rectangle) []int {
		return []int{rect.Dx() - 1}
	}
	emptyRows = [][]string{[]string{"No data points"}}
)

func TableTypeCircle() *ring.Ring {
	types := []string{SummarizedType, AllType, NodeType}
	circle := ring.New(len(types))
	for _, typ := range types {
		circle.Value = typ
		circle = circle.Next()
	}
	return circle
}

func ResetTableShapeFrom(typ string, rect image.Rectangle) (string, []string, []int) {
	switch typ {
	case SummarizedType:
		return summarizedTitle, summarizedHeader, summarizedWidthFn(rect, 0)
	case AllType:
		return allTitle, allHeader, allWidthFn(rect, 0, 0)
	case NodeType:
		return nodeTitle, nodeHeader, nodeWidthFn(rect, 0)
	default:
		return summarizedTitle, summarizedHeader, summarizedWidthFn(rect, 0)
	}
}

func AsAllTableViewer(resources []*Resource, sortType SortType) ResourceTableViewer {
	switch sortType {
	case ByName:
		return sortByName(resources)
	default:
		return sortByName(resources)
	}
}

type sortByName []*Resource

func (s sortByName) GetTableShape(rect image.Rectangle) (string, []string, []int, [][]string) {
	rows := make([][]string, len(s))
	var maxLen0, maxLen1 int
	for i, v := range s {
		rows[i] = v.toRow()
		maxLen0 = IntMax(maxLen0, len(rows[i][0]))
		maxLen1 = IntMax(maxLen1, len(rows[i][1]))
	}
	title, header, widths :=
		allTitle, allHeader, allWidthFn(rect, maxLen0, maxLen1)

	if len(s) == 0 {
		header = emptyHeader
		widths = emptyWidthFn(rect)
		rows = emptyRows
	}
	return title, header, widths, rows
}

func (s sortByName) SortRows() {
	sort.Slice(s, func(i, j int) bool {
		if s[i].podName < s[j].podName {
			return true
		}
		if s[i].podName > s[j].podName {
			return false
		}
		return s[i].containerName < s[j].containerName
	})
}
