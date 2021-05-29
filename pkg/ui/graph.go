package ui

import (
	"image"

	. "github.com/gizak/termui/v3"
)

const (
	widthFromLeftBorder = 2
)

type Graph struct {
	graph *graph
	label *label
}

type Type string

const (
	TimeSeries Type = "timeseries"
	BaseLine   Type = "baseline"
)

type Data struct {
	Type  Type
	Value float64
	Label string
	Style Style
}

func NewGraph() *Graph {
	return &Graph{
		graph: newGraph(),
		label: newLabel(),
	}
}

func (self *Graph) Reset() {
	self.graph.reset()
	self.label.reset()
}

func (self *Graph) Update(list []Data) {
	var constants []struct {
		value float64
		style Style
	}
	for _, d := range list {
		switch d.Type {
		case TimeSeries:
			self.graph.timeseries.values = append(self.graph.timeseries.values, d.Value)
			self.graph.timeseries.style = d.Style
		case BaseLine:
			constants = append(constants, struct {
				value float64
				style Style
			}{
				value: d.Value,
				style: d.Style,
			})
		}
		self.graph.baselines = constants
		self.label.values = append(self.label.values, struct {
			value string
			style Style
		}{
			value: d.Label,
			style: d.Style,
		})
	}
}

func (self *Graph) SetTitle(title string, style Style) {
	self.graph.Title = title
	self.graph.TitleStyle = style
}

func (self *Graph) SetGraphBorderStyle(style Style) {
	self.graph.BorderStyle = style
}

func (self *Graph) SetLabelBorderStyle(style Style) {
	self.label.BorderStyle = style
}

func (self *Graph) Grid() *Grid {
	grid := NewGrid()
	grid.Set(NewCol(4./5, self.graph), NewCol(1./5, self.label))
	return grid
}

type graph struct {
	*Block

	baselines []struct {
		value float64
		style Style
	}
	timeseries struct {
		values []float64
		style  Style
	}
}

func newGraph() *graph {
	return &graph{
		Block: NewBlock(),

		timeseries: struct {
			values []float64
			style  Style
		}{
			values: make([]float64, 0),
		},
	}
}

func (self *graph) reset() {
	self.timeseries.values = make([]float64, 0)
	self.baselines = make([]struct {
		value float64
		style Style
	}, 0)
}

func (self *graph) getY(val float64) int {
	var max float64
	for _, c := range self.baselines {
		if max < c.value {
			max = c.value
		}
	}
	if max < val {
		max = val
	}
	dy := self.Inner.Max.Y - (self.Inner.Min.Y + 4)
	return int(float64(self.Inner.Max.Y) - float64(dy)*(val/max))
}

func (self *graph) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	canvas := NewCanvas()
	canvas.Rectangle = self.Inner

	for _, c := range self.baselines {
		canvas.SetLine(
			image.Pt(self.Inner.Min.X*2, self.getY(c.value)*4),
			image.Pt(self.Inner.Max.X*2, self.getY(c.value)*4),
			c.style.Fg,
		)
	}

	if len(self.timeseries.values) > 0 {
		dest := self.getY(self.timeseries.values[len(self.timeseries.values)-1])
		for di := len(self.timeseries.values) - 1; di >= 0; di-- {
			src := self.getY(self.timeseries.values[di])
			canvas.SetLine(
				image.Pt((self.Inner.Min.X+di)*2, dest*4),
				image.Pt((self.Inner.Min.X+di+1)*2, src*4),
				self.timeseries.style.Fg,
			)
			dest = src
		}
	}
	canvas.Draw(buf)
}

type label struct {
	*Block

	values []struct {
		value string
		style Style
	}
}

func newLabel() *label {
	return &label{
		Block: NewBlock(),
		values: make([]struct {
			value string
			style Style
		}, 0),
	}
}

func (self *label) reset() {
	self.values = make([]struct {
		value string
		style Style
	}, 0)
}

func (self *label) Draw(buf *Buffer) {
	self.Block.Draw(buf)

	startPosXForLabel := self.Inner.Min.X + widthFromLeftBorder
	trimWidth := self.Inner.Max.X - (widthFromLeftBorder + 1)

	for i, label := range self.values {
		buf.SetString(
			TrimString(label.value, trimWidth),
			label.style,
			image.Pt(startPosXForLabel, self.Inner.Min.Y+i),
		)
	}
}
