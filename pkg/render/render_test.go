package render_test

import (
	"image"
	"testing"

	"crdx.org/label/pkg/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func inkBounds(img *image.Gray) (minY int, maxY int, inked bool) {
	bounds := img.Bounds()
	minY, maxY = bounds.Max.Y, bounds.Min.Y
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if img.GrayAt(x, y).Y < 128 {
				inked = true
				minY = min(minY, y)
				maxY = max(maxY, y)
			}
		}
	}
	return minY, maxY, inked
}

func TestTextFillsBand(t *testing.T) {
	const glyphHeight = 70
	const imageHeight = 128

	img, err := render.Text("Hgjpqy", glyphHeight, imageHeight)
	require.NoError(t, err)

	assert.Equal(t, imageHeight, img.Bounds().Dy())
	assert.Greater(t, img.Bounds().Dx(), 16)

	minY, maxY, inked := inkBounds(img)
	require.True(t, inked, "expected inked pixels")

	const tolerance = 3
	top := (imageHeight - glyphHeight) / 2
	assert.GreaterOrEqual(t, minY, top-tolerance, "ink strays above the centred band")
	assert.LessOrEqual(t, maxY, top+glyphHeight+tolerance, "ink strays below the centred band")
	assert.GreaterOrEqual(t, maxY-minY+1, glyphHeight-2*tolerance, "ink does not fill the band")
}

func TestTextErrors(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		content     string
		glyphHeight int
		imageHeight int
	}{
		{"empty", "", 70, 128},
		{"whitespace", "   ", 70, 128},
		{"glyph taller than image", "x", 200, 128},
		{"non-positive glyph height", "x", 0, 128},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := render.Text(testCase.content, testCase.glyphHeight, testCase.imageHeight)
			assert.Error(t, err)
		})
	}
}
