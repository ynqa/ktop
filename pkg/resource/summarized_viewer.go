package resource

import (
	"image"
	"sort"

	. "github.com/ynqa/ktop/pkg/util"
)

var (
	summarizedTitle  = "⎈  Pod ⎈ "
	summarizedHeader = []string{
		"POD", "CPU(U)", "Memory(U)",
	}
	summarizedWidthFn = func(rect image.Rectangle, maxLen int) []int {
		nameWidth := IntMax(50, IntMin(rect.Dx()-20, maxLen+indentSize))
		return []int{nameWidth, 10, 10}
	}
)

func AsSummarizedTableViewer(resources []*SummarizedResource, sortType SortType) ResourceTableViewer {
	switch sortType {
	case ByName:
		return sortByNameForSummarized(resources)
	default:
		return sortByNameForSummarized(resources)
	}
}

type sortByNameForSummarized []*SummarizedResource

func (s sortByNameForSummarized) GetTableShape(rect image.Rectangle) (string, []string, []int, [][]string) {
	rows := make([][]string, len(s))
	var maxLen int
	for i, v := range s {
		rows[i] = v.toRow()
		maxLen = IntMax(maxLen, len(rows[i][0]))
	}
	title, header, widths :=
		summarizedTitle, summarizedHeader, summarizedWidthFn(rect, maxLen)

	if len(s) == 0 {
		header = emptyHeader
		widths = emptyWidthFn(rect)
		rows = emptyRows
	}
	return title, header, widths, rows
}

func (s sortByNameForSummarized) SortRows() {
	sort.Slice(s, func(i, j int) bool {
		return s[i].podName < s[j].podName
	})
}
