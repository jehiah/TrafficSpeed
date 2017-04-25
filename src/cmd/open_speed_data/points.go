package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

type BBox struct {
	A Point `json:"a"`
	B Point `json:"b"`
}

func ParseBBox(s string) (b *BBox) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "x") || !strings.Contains(s, " ") {
		return nil
	}
	b = &BBox{}
	c := strings.SplitN(s, " ", 2)
	p1 := ParsePoint(c[0])
	p2 := ParsePoint(c[1])
	// for a bounding box, always top left and bottom right
	b.A.X = math.Min(p1.X, p2.X)
	b.A.Y = math.Min(p1.Y, p2.Y)
	b.B.X = math.Max(p1.X, p2.X)
	b.B.Y = math.Max(p1.Y, p2.Y)
	return
}

func (b *BBox) IsZero() bool {
	if b.A.X == 0 && b.A.Y == 0 && b.B.X == 0 && b.B.Y == 0 {
		return true
	}
	return false
}
func (b *BBox) String() string {
	return fmt.Sprintf("%s %s", b.A, b.B)
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (p Point) String() string {
	return fmt.Sprintf("%dx%d", int64(p.X), int64(p.Y))
}

func ParsePoint(s string) (p Point) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "x") {
		return
	}
	c := strings.SplitN(s, "x", 2)
	x, _ := strconv.Atoi(c[0])
	y, _ := strconv.Atoi(c[1])
	return Point{float64(x), float64(y)}
}

const SkipRotate = 0.00001

func RadiansToDegrees(rad float64) (deg float64) {
	deg = rad * 180.0 / math.Pi
	return
}

const rightAngelRadians = 1.570796 // 90 degrees

func Radians(a, b Point) float64 {
	if a.Y == b.Y {
		return SkipRotate
	}

	adjacent := math.Max(a.X, b.X) - math.Min(a.X, b.X)
	opposite := math.Max(a.Y, b.Y) - math.Min(a.Y, b.Y)
	radians := math.Atan(adjacent / opposite)
	log.Printf("adjacent: %v opposite %v radians %v", adjacent, opposite, radians)
	if a.Y > b.Y {
		adjusted := (-1 * radians) + rightAngelRadians
		log.Printf("adjusting to %v because %v < %v", adjusted, a, b)
		return adjusted
	}
	adjusted := radians - rightAngelRadians
	log.Printf("adjusting to %v because %v > %v", adjusted, a, b)
	return adjusted
}
