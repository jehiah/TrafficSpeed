package labelimg

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelimg(t *testing.T) {

	l := &image.Gray{
		Pix: []uint8{3, 3, 0, 4,
			3, 0, 4, 4,
		},
		Stride: 4,
		Rect:   image.Rect(1, 1, 5, 3), //  4 x 2
	}

	labeled := New(l, 1)

	assert.Equal(t, []uint8{1, 1, 0, 2,
		1, 0, 2, 2}, labeled.Pix)

}
