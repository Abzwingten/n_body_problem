package utils

import (
	"github.com/ungerik/go3d/vec2"
	"github.com/veandco/go-sdl2/sdl"
)

func ColorKeyToSDL(_color uint32) sdl.Color {
	var sdlcolor sdl.Color
	sdlcolor.R = (uint8)(_color >> 24)
	sdlcolor.G = (uint8)(_color >> 16)
	sdlcolor.B = (uint8)(_color >> 8)
	sdlcolor.A = (uint8)(_color)
	return sdlcolor
}

func DistanceTo(v1, v2 *vec2.T) float32 {
	a := v1.Sub((*vec2.T)(v2))
	return a.Length()
}