package ui

import (
	"chip8/chip8"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const scale = 16

type Display struct {
	window   *sdl.Window
	surface  *sdl.Surface
	renderer *sdl.Renderer
}

func InitDisplay() *Display {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	// defer sdl.Quit()

	window, renderer, err := sdl.CreateWindowAndRenderer(chip8.VWidth*scale, chip8.VHeight*scale, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	display := &Display{window, surface, renderer}
	return display
}

func drawPixel(display *Display, r int32, c int32) {
	x := int32(c * scale)
	y := int32(r * scale)
	rect := sdl.Rect{X: x, Y: y, W: scale, H: scale}
	display.surface.FillRect(&rect, 0xFFFFFFFF)
	display.window.UpdateSurface()
}

func copyDisplay(video *[chip8.VHeight][chip8.VWidth]bool, display *Display) {
	for r := int32(0); r < chip8.VHeight; r++ {
		for c := int32(0); c < chip8.VWidth; c++ {
			if video[r][c] {
				drawPixel(display, r, c)
			}
		}
	}
	display.window.UpdateSurface()
}

func Loop(console *chip8.Console, display *Display) {
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
		console.EmulateCycle()
		// clearDisplay(display)
		copyDisplay(&console.Video, display)
		// fmt.Println(console.Video)
		time.Sleep(16 * time.Millisecond)
	}
}
