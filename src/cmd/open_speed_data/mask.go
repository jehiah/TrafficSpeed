package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Mask struct {
	Start    int64 `json:"start,omitempty"`
	End      int64 `json:"end,omitempty"`
	BBox     *BBox `json:"bbox,omitempty"`
	NullMask bool  `json:"null_mask,omitempty"`
}

func (m Mask) String() string {
	if m.NullMask {
		return "-"
	}
	if m.Start != 0 && m.End != 0 {
		return fmt.Sprintf("%d:%d", m.Start, m.End)
	}
	return m.BBox.String()
}

func ParseMask(s string) (m Mask, ok bool) {
	s = strings.TrimSpace(s)
	switch {
	case s == "-":
		m.NullMask = true
		ok = true
	case strings.Count(s, ":") == 1:
		c := strings.SplitN(s, ":", 2)
		x, _ := strconv.Atoi(c[0])
		y, _ := strconv.Atoi(c[1])
		m.Start = int64(math.Min(float64(x), float64(y)))
		m.End = int64(math.Max(float64(x), float64(y)))
		ok = true
	case strings.Count(s, "x") == 2 && strings.Count(s, " ") == 1:
		m.BBox = ParseBBox(s)
		ok = !m.BBox.IsZero()
	}
	return
}
