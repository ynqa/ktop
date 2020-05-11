package resource

import (
	"image"
	"sort"

	. "github.com/ynqa/ktop/pkg/util"
)

var (
	nodeTitle  = "⎈ Node ⎈"
	nodeHeader = []string{
		"NODE",
		"CPU(A)", "CPU(U)", "%CPU",
		"Memory(A)", "Memory(U)", "%Memory",
	}
	nodeWidthFn = func(rect image.Rectangle, maxLen int) []int {
		nameWidth := IntMax(50, IntMin(rect.Dx()-20, maxLen+indentSize))
		return []int{nameWidth, 10, 10, 10, 10, 10, 10}
	}
)

func AsNodeTableViewer(resources []*NodeResource, sortType SortType) ResourceTableViewer {
	switch sortType {
	case ByName:
		return sortByNameForNode(resources)
	default:
		return sortByNameForNode(resources)
	}
}

type sortByNameForNode []*NodeResource

func (s sortByNameForNode) GetTableShape(rect image.Rectangle) (string, []string, []int, [][]string) {
	rows := make([][]string, len(s))
	var maxLen int
	for i, v := range s {
		rows[i] = v.toRow()
		maxLen = IntMax(maxLen, len(rows[i][0]))
	}
	title, header, widths :=
		nodeTitle, nodeHeader, nodeWidthFn(rect, maxLen)

	if len(s) == 0 {
		header = emptyHeader
		widths = emptyWidthFn(rect)
		rows = emptyRows
	}
	return title, header, widths, rows
}

func (s sortByNameForNode) GetRows() [][]string {
	if len(s) == 0 {
		return emptyRows
	}
	rows := make([][]string, len(s))
	for i, v := range s {
		rows[i] = v.toRow()
	}
	return rows
}

func (s sortByNameForNode) SortRows() {
	sort.Slice(s, func(i, j int) bool {
		return s[i].nodeName < s[j].nodeName
	})
}
