package labelimg

import (
	"image"
	"image/color"
	"sort"
)

type Label struct {
	Center image.Point
	Mean   image.Point
	Bounds image.Rectangle
	Color  color.Color
	Pixels int
	// Mask image.Imate
}

func Labels(i *image.Paletted) []Label {
	var labels []Label
	for ci, c := range i.Palette {
		if ci == 0 {
			// the background
			continue
		}
		l := Label{Color: c}
		var xaxis, yaxis []int = make([]int, i.Rect.Max.X), make([]int, i.Rect.Max.Y)
		for x := i.Rect.Min.X; x < i.Rect.Max.X; x++ {
			for y := i.Rect.Min.Y; y < i.Rect.Max.Y; y++ {
				if i.ColorIndexAt(x, y) == uint8(ci) {
					xaxis[x]++
					yaxis[y]++
					l.Pixels += 1
				}
			}
		}

		xmin, xmax := extent(xaxis)
		ymin, ymax := extent(yaxis)
		l.Bounds = image.Rect(xmin, ymin, xmax, ymax)
		l.Center = image.Pt((xmin+xmax)/2, (ymin+ymax)/2)
		if l.Pixels > 0 {
			labels = append(labels, l)
		}
	}
	sort.Slice(labels, func(i, j int) bool { return labels[i].Pixels > labels[j].Pixels })
	return labels
}

// extent returns the first and last non-zero index
func extent(v []int) (min int, max int) {
	max = len(v) - 1
	if max == -1 {
		return 0, 0
	}
	for ; min <= max; min++ {
		if v[min] != 0 {
			break
		}
	}
	for ; max > 0; max-- {
		if v[max] != 0 {
			break
		}
	}
	if min > max {
		return 0, 0
	}
	return
}
