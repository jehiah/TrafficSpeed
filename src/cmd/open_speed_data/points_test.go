package main

import (
	"testing"
)

func TestRadians(t *testing.T) {
	type testCase struct {
		A       Point
		B       Point
		Radians float64
	}
	tests := []testCase{
		{Point{10, 10}, Point{20, 11}, 0.0996683256962656},
		{Point{10, 10}, Point{20, 9}, -0.0996683256962656},
	}

	for _, tc := range tests {
		got := Radians(tc.A, tc.B)
		if got != tc.Radians {
			t.Errorf("got %v expected %v for %v %v", got, tc.Radians, tc.A, tc.B)
		}
	}
}
