package kitty

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/samber/lo"
)

func PrintImage(img image.Image) error {
	var buffa bytes.Buffer
	if err := png.Encode(&buffa, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(buffa.Bytes())

	for offset := 0; offset < len(encoded); offset += 4096 {
		end := min(offset+4096, len(encoded))
		hasMore := end < len(encoded)

		control := fmt.Sprintf("m=%d", lo.Ternary(hasMore, 1, 0))
		if offset == 0 {
			control = "a=T,f=100," + control
		}

		if _, err := fmt.Fprintf(os.Stdout, "\x1b_G%s;%s\x1b\\", control, encoded[offset:end]); err != nil {
			return err
		}
	}

	_, err := fmt.Fprintln(os.Stdout)
	return err
}
