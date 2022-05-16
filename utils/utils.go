package utils

import (
	"image/color"

	"github.com/ungerik/go3d/float64/vec2"
)

func ColorKeyToColor(_color uint32) color.RGBA {
	var pixelcolor color.RGBA
	pixelcolor.R = (uint8)(_color >> 24)
	pixelcolor.G = (uint8)(_color >> 16)
	pixelcolor.B = (uint8)(_color >> 8)
	pixelcolor.A = (uint8)(_color)
	return pixelcolor
}

func DistanceTo(v1, v2 *vec2.T) float64 {
	a := v1.Sub(v2)
	return a.Length()
}