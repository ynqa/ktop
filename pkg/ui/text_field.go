package ui

import (
	"image"

	. "github.com/gizak/termui"
)

type TextField struct {
	*Block

	Text      string
	TextStyle Style
}

func NewTextField() *TextField {
	return &TextField{
		Block: NewBlock(),
	}
}

func (self *TextField) Draw(buf *Buffer) {
	cells := ParseStyles(self.Text, self.TextStyle)
	rows := SplitCells(cells, '\n')

	for y, row := range rows {
		if y+self.Inner.Min.Y >= self.Inner.Max.Y {
			break
		}
		row = TrimCells(row, self.Inner.Dx())
		for _, cx := range BuildCellWithXArray(row) {
			x, cell := cx.X, cx.Cell
			buf.SetCell(cell, image.Pt(x, y).Add(self.Inner.Min))
		}
	}
}
