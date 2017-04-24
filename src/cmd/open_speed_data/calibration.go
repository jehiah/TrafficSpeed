package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Calibration struct {
	Seek   float64 `json:"seek"`
	A      Point   `json:"a"`
	B      Point   `json:"b"`
	Inches float64 `json:"inches"`
}

func (c *Calibration) String() string {
	return fmt.Sprintf("%0.4f %s %s %0.4f", c.Seek, c.A, c.B, c.Inches)
}
func (c *Calibration) Pretty() string {
	return fmt.Sprintf("Seek:%0.4fsec Points{%s %s} Inches:%0.4f", c.Seek, c.A, c.B, c.Inches)
}

func ParseCalibration(s string) (c *Calibration) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "x") || !strings.Contains(s, " ") {
		return nil
	}
	chunks := strings.SplitN(s, " ", 4)

	c = &Calibration{}
	var err error
	c.Seek, err = strconv.ParseFloat(chunks[0], 64)
	if err != nil {
		return nil
	}
	c.A = ParsePoint(chunks[1])
	c.B = ParsePoint(chunks[2])
	c.Inches, err = strconv.ParseFloat(chunks[3], 64)
	if err != nil {
		return nil
	}
	return
}
