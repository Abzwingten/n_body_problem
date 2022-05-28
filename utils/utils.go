package utils

import (
	"image/color"
)

func ColorKeyToColor(_color uint32) color.RGBA {
	var pixelcolor color.RGBA
	pixelcolor.R = (uint8)(_color >> 24)
	pixelcolor.G = (uint8)(_color >> 16)
	pixelcolor.B = (uint8)(_color >> 8)
	pixelcolor.A = (uint8)(_color)
	return pixelcolor
}
