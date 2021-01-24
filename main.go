package main

import (
	"chip8/chip8"
	"chip8/ui"
)

func main() {
	display := ui.InitDisplay()
	console := chip8.InitConsole()
	chip8.LoadRom(console, "test_opcode.ch8")
	ui.Loop(console, display)
}
