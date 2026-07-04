package render

import (
	_ "embed"
	"fmt"
	"image"
	"image/draw"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/JetBrainsMono-Regular.ttf
var fontData []byte

const (
	dpi           = 72
	margin        = 8
	referenceSize = 256
	maxImageWidth = 10000
)

func Text(content string, glyphHeight int, imageHeight int) (*image.Gray, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("text must not be empty")
	}
	if glyphHeight <= 0 {
		return nil, fmt.Errorf("glyph height must be positive")
	}
	if imageHeight <= 0 {
		return nil, fmt.Errorf("image height must be positive")
	}
	if glyphHeight > imageHeight {
		return nil, fmt.Errorf("glyph height %d exceeds image height %d", glyphHeight, imageHeight)
	}

	typeface, err := opentype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}

	newFace := func(size float64) (font.Face, error) {
		return opentype.NewFace(typeface, &opentype.FaceOptions{
			Size:    size,
			DPI:     dpi,
			Hinting: font.HintingFull,
		})
	}

	referenceFace, err := newFace(referenceSize)
	if err != nil {
		return nil, fmt.Errorf("create reference face: %w", err)
	}
	defer func() { _ = referenceFace.Close() }()

	referenceBounds, _ := font.BoundString(referenceFace, content)
	referenceInk := float64(referenceBounds.Max.Y-referenceBounds.Min.Y) / 64
	if referenceInk <= 0 {
		return nil, fmt.Errorf("text has no printable glyphs")
	}

	size := referenceSize * float64(glyphHeight) / referenceInk
	face, err := newFace(size)
	if err != nil {
		return nil, fmt.Errorf("create face: %w", err)
	}
	defer func() { _ = face.Close() }()

	bounds, _ := font.BoundString(face, content)
	inkHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

	if inkHeight > imageHeight {
		_ = face.Close()
		size *= float64(imageHeight) / float64(inkHeight)
		face, err = newFace(size)
		if err != nil {
			return nil, fmt.Errorf("create face: %w", err)
		}
		bounds, _ = font.BoundString(face, content)
		inkHeight = (bounds.Max.Y - bounds.Min.Y).Ceil()
	}

	inkWidth := (bounds.Max.X - bounds.Min.X).Ceil()

	imageWidth := inkWidth + margin*2
	if imageWidth > maxImageWidth {
		return nil, fmt.Errorf("text is too long: %d dots exceeds %d", imageWidth, maxImageWidth)
	}

	img := image.NewGray(image.Rect(0, 0, imageWidth, imageHeight))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	inkTop := (imageHeight - inkHeight) / 2
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(margin) - bounds.Min.X,
			Y: fixed.I(inkTop) - bounds.Min.Y,
		},
	}
	drawer.DrawString(content)

	return img, nil
}
