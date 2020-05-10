package ui

import (
	"image"
	"strings"

	. "github.com/gizak/termui/v3"
)

type Table struct {
	*Block

	Header       []string
	ColumnWidths []int
	Rows         [][]string
	Cursor       bool
	CursorColor  Color
	topRow       int

	SelectedRow int
}

func NewTable() *Table {
	return &Table{
		Block:       NewBlock(),
		Cursor:      true,
		topRow:      0,
		SelectedRow: 0,
	}
}

func (self *Table) Reset(title string, header []string, width []int) {
	self.Title = title
	self.Header = header
	self.ColumnWidths = width
	self.Rows = [][]string{}
	self.topRow = 0
	self.SelectedRow = 0
}

func (self *Table) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	if self.Inner.Dy() > 2 {
		// store positions for each column
		columnPositions := []int{}
		var cur int
		for _, w := range self.ColumnWidths {
			columnPositions = append(columnPositions, cur)
			cur += w
		}

		// describe a header
		for i, h := range self.Header {
			buf.SetString(
				h,
				NewStyle(Theme.Default.Fg, ColorClear, ModifierBold),
				image.Pt(self.Inner.Min.X+columnPositions[i], self.Inner.Min.Y),
			)
		}

		if self.SelectedRow < self.topRow {
			self.topRow = self.SelectedRow
		} else if self.SelectedRow > self.cursorBottom() {
			self.topRow = self.cursorBottom()
		}

		// describe rows
		for idx := self.topRow; idx >= 0 && idx < len(self.Rows) && idx < self.bottom(); idx++ {
			row := self.Rows[idx]
			// move y+1 for a header
			y := self.Inner.Min.Y + 1 + idx - self.topRow
			style := NewStyle(Theme.Default.Fg)
			if self.Cursor {
				if idx == self.SelectedRow {
					style.Fg = self.CursorColor
					style.Modifier = ModifierReverse
					buf.SetString(
						strings.Repeat(" ", self.Inner.Dx()),
						style,
						image.Pt(self.Inner.Min.X, y),
					)
					self.SelectedRow = idx
				}
			}
			for i, width := range self.ColumnWidths {
				r := TrimString(row[i], width)
				buf.SetString(
					r,
					style,
					image.Pt(self.Inner.Min.X+columnPositions[i], y),
				)
			}
		}
	}
}

func (self *Table) cursorBottom() int {
	return self.topRow + self.Inner.Dy() - 2
}

func (self *Table) bottom() int {
	return self.topRow + self.Inner.Dy() - 1
}

func (self *Table) scroll(i int) {
	self.SelectedRow += i
	maxRow := len(self.Rows) - 1
	if self.SelectedRow < 0 {
		self.SelectedRow = 0
	} else if self.SelectedRow > maxRow {
		self.SelectedRow = maxRow
	}
}

func (self *Table) ScrollUp() {
	self.scroll(-1)
}

func (self *Table) ScrollDown() {
	self.scroll(1)
}
