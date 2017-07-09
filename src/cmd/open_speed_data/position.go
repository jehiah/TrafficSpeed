package main

import (
	"fmt"
	"image"
	"math"
	"time"

	"labelimg"
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

type FramePosition struct {
	Frame     int
	Time      time.Duration
	Positions []labelimg.Label
}

type VehiclePosition struct {
	Frame     int
	Time      time.Duration
	VehicleID int
	Position  labelimg.Label
}

// TrackVehicles tracks detected objects and correlates across frames
// based on logic from identifyvehicles from https://github.com/mbauman/TrafficSpeed/blob/master/TrafficSpeed.ipynb
func TrackVehicles(frames []FramePosition) []VehiclePosition {
	var vehicleCount int
	var vehicles []VehiclePosition
	var lastFrameVehicles []VehiclePosition

	for _, frame := range frames {
		var currentFrameVehicles []VehiclePosition
		for _, position := range frame.Positions {
			// is position overlap from lastFrame
			var vehicleID int
			// TODO: use median point not center
			if closest := ClosestPosition(position.Center, lastFrameVehicles); position.Center.In(closest.Position.Bounds) {
				vehicleID = closest.VehicleID
			} else {
				vehicleCount++
				vehicleID = vehicleCount
			}
			currentFrameVehicles = append(currentFrameVehicles, VehiclePosition{
				Frame:     frame.Frame,
				Time:      frame.Time,
				VehicleID: vehicleID,
				Position:  position,
			})
		}
		lastFrameVehicles = currentFrameVehicles
		vehicles = append(vehicles, currentFrameVehicles...)
	}
	return vehicles
}

func ClosestPosition(point image.Point, v []VehiclePosition) VehiclePosition {
	var closest VehiclePosition
	var min float64 = -1
	for _, p := range v {
		// TODO: use median point not center
		d := distance(point, p.Position.Center)
		if min == -1 || d < min {
			min = d
			closest = p
		}
	}
	return closest
}

func distance(a, b image.Point) float64 {
	x := math.Abs(float64(a.X) - float64(b.X))
	y := math.Abs(float64(a.Y) - float64(b.Y))
	return math.Sqrt((x * x) + (y * y))
}
