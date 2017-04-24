package main

import (
	"fmt"
)

// Position matches Position in position.jl
type Position struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Mass  int     `json:"mass"`
	XSpan []int   `json:"xspan"` // [1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16]
	YSpan []int   `json:"yspan"`
}

func (p Position) Span() string {
	mm := func(d []int) (min int, max int) {
		for i, n := range d {
			if n < min || i == 0 {
				min = n
			}
			if n > max || i == 0 {
				max = n
			}
		}
		return
	}
	xmin, xmax := mm(p.XSpan)
	ymin, ymax := mm(p.YSpan)
	return fmt.Sprintf("x:%d-%d y:%d-%d", xmin, xmax, ymin, ymax)
}
func (p Position) Size() string {
	return fmt.Sprintf("%dx%d", len(p.XSpan), len(p.YSpan))
}
