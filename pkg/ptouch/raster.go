package ptouch

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"

	"crdx.org/label/pkg/util"
)

const PrintHeadWidth = 128

func Rasterise(img image.Image) ([]byte, int, error) {
	bounds := img.Bounds()
	size := bounds.Size()

	transposed := size.Y == PrintHeadWidth
	width, height := size.X, size.Y
	if transposed {
		width, height = size.Y, size.X
	} else if size.X != PrintHeadWidth {
		return nil, 0, fmt.Errorf("image size must have %dpx width or height, got: %dx%d", PrintHeadWidth, size.X, size.Y)
	}

	bytesWidth := width / 8
	if width%8 != 0 {
		bytesWidth++
	}

	data := make([]byte, bytesWidth*height)

	for y := range height {
		for x := range width {
			sourceX, sourceY := width-1-x, y
			if transposed {
				sourceX, sourceY = y, x
			}

			r, g, b, _ := img.At(bounds.Min.X+sourceX, bounds.Min.Y+sourceY).RGBA()

			// Perceptual lightness in [0,1] from the 16-bit channels, using the Rec. 709 luma
			// weights (0.2126, 0.7152, 0.0722) scaled to sum 255.
			lightness := float64(55*r+182*g+18*b) / float64(0xffff*(55+182+18))
			if lightness <= 0.5 {
				data[y*bytesWidth+x/8] |= 0x80 >> uint(x%8)
			}
		}
	}

	return data, bytesWidth, nil
}

func Pack(data []byte, bytesWidth int) ([]byte, error) {
	if bytesWidth <= 0 {
		return nil, fmt.Errorf("bytesWidth must be positive, got: %d", bytesWidth)
	}

	var buffer bytes.Buffer
	max := len(data)

	for i := 0; i < max; i += bytesWidth {
		to := min(i+bytesWidth, max)
		line := data[i:to]

		compressedLine := util.PackBits(line)

		if len(compressedLine) > len(line) {
			raw := make([]byte, 0, len(line)+1)
			raw = append(raw, byte(len(line)-1)) //nolint:gosec // line length is at most bytesWidth, a single protocol byte
			raw = append(raw, line...)
			compressedLine = raw
		}

		var lengthBytes [2]byte
		binary.LittleEndian.PutUint16(lengthBytes[:], uint16(len(compressedLine))) //nolint:gosec // a compressed line is at most bytesWidth+overhead bytes

		buffer.Write(cmdRasterTransfer)
		buffer.Write(lengthBytes[:])
		buffer.Write(compressedLine)
	}

	return buffer.Bytes(), nil
}
