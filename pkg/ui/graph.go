package ui

import (
	"image"

	. "github.com/gizak/termui/v3"
)

type Graph struct {
	*Block
	// plot data
	Data           []float64
	UpperLimit     float64
	DrawUpperLimit bool

	// label
	LabelHeader     string
	LabelData       string
	LabelUpperLimit string

	// color
	DataColor      Color
	LimitColor     Color
	LabelNameColor Color
}

func NewGraph() *Graph {
	return &Graph{
		Block: NewBlock(),
		Data:  make([]float64, 0),
	}
}

func (self *Graph) Reset() {
	self.Data = make([]float64, 0)
	self.UpperLimit = 0
	self.LabelHeader = ""
	self.LabelData = ""
	self.LabelUpperLimit = ""
}

func (self *Graph) calcHeight(val float64) int {
	return int((val / self.UpperLimit) * float64(self.Inner.Dy()-5))
}

func (self *Graph) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	// describe graph
	if len(self.Data) != 0 {
		canvas := NewCanvas()
		canvas.Rectangle = self.Inner
		// draw upper limit
		if self.DrawUpperLimit {
			limitHeight := self.calcHeight(self.UpperLimit)
			canvas.SetLine(
				image.Pt(
					self.Inner.Min.X*2,
					(self.Inner.Max.Y-limitHeight-1)*4,
				),
				image.Pt(
					self.Inner.Max.X*2,
					(self.Inner.Max.Y-limitHeight-1)*4,
				),
				self.LimitColor,
			)
		}

		// use latest data
		data := self.Data
		if len(self.Data) > self.Inner.Dx() {
			data = data[len(self.Data)-1-self.Inner.Dx() : len(self.Data)-1]
		}
		previousHeight := self.calcHeight(data[len(data)-1])
		for i := len(data) - 1; i >= 0; i-- {
			height := self.calcHeight(data[i])
			// draw data
			canvas.SetLine(
				image.Pt(
					(self.Inner.Min.X+i)*2,
					(self.Inner.Max.Y-previousHeight-1)*4,
				),
				image.Pt(
					(self.Inner.Min.X+i+1)*2,
					(self.Inner.Max.Y-height-1)*4,
				),
				self.DataColor,
			)
			previousHeight = height
		}
		canvas.Draw(buf)
	}

	// describe labels
	stage := 1
	if self.Inner.Dy() >= 3 {
		if self.LabelHeader != "" {
			buf.SetString(
				self.LabelHeader,
				NewStyle(self.LabelNameColor, ColorClear, ModifierBold),
				image.Pt(self.Inner.Min.X+1, self.Inner.Min.Y+stage),
			)
			stage++
		}
		if self.LabelUpperLimit != "" {
			buf.SetString(
				self.LabelUpperLimit,
				NewStyle(self.LimitColor),
				image.Pt(self.Inner.Min.X+2, self.Inner.Min.Y+stage),
			)
			stage++
		}
		if self.LabelData != "" {
			buf.SetString(
				self.LabelData,
				NewStyle(self.DataColor),
				image.Pt(self.Inner.Min.X+2, self.Inner.Min.Y+stage),
			)
		}
	}
}
