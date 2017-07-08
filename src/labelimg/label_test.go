package labelimg

import (
	"fmt"
	"testing"
)

func TestExtent(t *testing.T) {
	type testCase struct {
		min  int
		max  int
		have []int
	}

	tests := []testCase{
		{0, 0, []int{}},
		{1, 2, []int{0, 1, 1, 0}},
		{1, 3, []int{0, 1, 1, 3}},
		{0, 1, []int{2, 1, 0, 0}},
		{0, 0, []int{0, 0, 0, 0}},
	}

	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			min, max := extent(tc.have)
			t.Logf("%#v", tc.have)
			if min != tc.min {
				t.Errorf("got min %d expected %d", min, tc.min)
			}
			if max != tc.max {
				t.Errorf("got max %d expected %d", max, tc.max)
			}

		})
	}
}
