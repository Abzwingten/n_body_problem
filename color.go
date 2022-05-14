package color

import (
	"github.com/veandco/go-sdl2/sdl"
)

func ColorKeyToSDL(_color uint32) sdl.Color {
	var sdlcolor sdl.Color
	sdlcolor.R = (uint8)(_color >> 16)
	sdlcolor.G = (uint8)(_color >> 8)
	sdlcolor.B = (uint8)(_color)
	return sdlcolor
}