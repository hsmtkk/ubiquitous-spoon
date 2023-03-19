package thumbnail

import (
	"fmt"
	"image"
	"io"

	"golang.org/x/image/draw"

	"image/gif"
	"image/jpeg"
	"image/png"
)

type Maker interface {
	Make(input io.Reader, output io.Writer, size int) error
}

func NewMaker() Maker {
	return &makerImpl{}
}

type makerImpl struct {
}

func (t *makerImpl) Make(input io.Reader, output io.Writer, size int) error {
	decoded, imageType, err := image.Decode(input)
	if err != nil {
		return fmt.Errorf("failed to decode image; %w", err)
	}
	origRect := decoded.Bounds()
	newX := origRect.Dx()
	newY := origRect.Dy()
	if origRect.Dx() > origRect.Dy() {
		newX = size
		newY = origRect.Dy() * size / origRect.Dx()
	} else {
		newX = origRect.Dx() * size / origRect.Dy()
		newY = size
	}
	newRect := image.NewRGBA(image.Rect(0, 0, newX, newY))
	draw.CatmullRom.Scale(newRect, newRect.Bounds(), decoded, origRect, draw.Over, nil)
	switch imageType {
	case "gif":
		if err := gif.Encode(output, newRect, nil); err != nil {
			return fmt.Errorf("failed to encode gif; %w", err)
		}
	case "jpeg":
		if err := jpeg.Encode(output, newRect, nil); err != nil {
			return fmt.Errorf("failed to encode jpeg; %w", err)
		}
	case "png":
		if err := png.Encode(output, newRect); err != nil {
			return fmt.Errorf("failed to encode png; %w", err)
		}
	}
	return fmt.Errorf("unsupported image type %s; %w", imageType, err)
}
