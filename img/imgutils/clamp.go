package imgutils

import (
	"image"
)

// BWClamp converts a greyscale image to black and white based on a specific cutoff
func BWClamp(img *image.Gray, cutoff uint8) {
	for i, c := range img.Pix {
		if c >= cutoff {
			img.Pix[i] = 255
		} else {
			img.Pix[i] = 0
		}
	}
}
