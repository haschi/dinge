package model

import (
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
)

func LoadImage(reader io.Reader) (image.Image, error) {
	im, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return im, nil
}

func EncodeImage(w io.Writer, image image.Image) error {
	return png.Encode(w, image)
}

const thumbnailWidth = 285

func Resize(src image.Image) image.Image {

	width := src.Bounds().Dx()
	height := src.Bounds().Dy()

	size := image.Rect(0, 0, thumbnailWidth, height*thumbnailWidth/width)
	dst := image.NewRGBA(size)

	for dstX := 0; dstX < size.Dx(); dstX++ {
		for dstY := 0; dstY < size.Dy(); dstY++ {
			srcX := dstX * width / thumbnailWidth
			srcY := dstY * height / size.Dy()
			orig := src.At(srcX, srcY)
			dst.Set(dstX, dstY, orig)
		}
	}

	return dst
}
