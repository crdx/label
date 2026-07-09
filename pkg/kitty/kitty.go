package kitty

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"

	"github.com/samber/lo"
)

const chunkSize = 4096

func PrintImage(fd io.Writer, img image.Image) error {
	var buffa bytes.Buffer
	if err := png.Encode(&buffa, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buffa.Bytes())

	for offset := 0; offset < len(encoded); offset += chunkSize {
		end := min(offset+chunkSize, len(encoded))
		hasMore := end < len(encoded)

		control := fmt.Sprintf("m=%d", lo.Ternary(hasMore, 1, 0))
		if offset == 0 {
			control = "a=T,f=100," + control
		}

		if _, err := fmt.Fprintf(fd, "\x1b_G%s;%s\x1b\\", control, encoded[offset:end]); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(fd)
	return err
}
