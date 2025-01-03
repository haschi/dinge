package model

import (
	"image"
	"image/draw"
)

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

func Crop(src image.Image) image.Image {

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width > height {
		dst := image.NewRGBA(image.Rect(0, 0, height, height))
		min := image.Pt((width/2)-(height/2), 0)
		max := image.Pt((width/2)+(height/2), height)
		crop := image.Rectangle{Min: min, Max: max}.Canon()
		draw.Draw(dst, crop.Sub(min), src, min, draw.Src)
		return dst
	}

	if width < height {
		min := image.Pt(0, (height/2)-(width/2))
		max := image.Pt(width, (height/2)+(width/2))
		crop := image.Rectangle{Min: min, Max: max}.Canon()
		dst := image.NewRGBA(image.Rect(0, 0, width, width))
		draw.Draw(dst, crop.Sub(min), src, min, draw.Src)
		return dst
	}

	return src
}
