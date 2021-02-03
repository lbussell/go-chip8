package chip8

import (
	"fmt"
	"io/ioutil"
	"math/rand"
)

const (
	fontOffset  = 0x050
	romOffset   = 0x200
	stackOffset = 0x052
)

const (
	VHeight int32 = 32
	VWidth  int32 = 64
)

type Console struct {
	cpu      CPU
	ram      []byte
	keyboard [16]bool
	Video    [VHeight][VWidth]bool
}

type CPU struct {
	v     [16]byte // registers
	i     uint16
	pc    uint16 // program counter
	sp    uint8  // stack pointer
	stack [16]uint16
	dt    uint8  // delay timer
	st    uint8  // sound timer
	op    uint16 // opcode
}

func InitConsole() *Console {
	console := &Console{}
	console.cpu.pc = romOffset // starting point for the program
	console.cpu.sp = 0         // faking the stack
	console.ram = make([]byte, 4096)
	copy(console.ram[fontOffset:], font)
	return console
}

func LoadRom(console *Console, romFile string) {
	rom, error := ioutil.ReadFile(romFile)
	if error != nil {
		panic(error)
	}
	copy(console.ram[romOffset:], rom)
}

func noop(opcode uint16) {
	fmt.Printf("Instruction not implemented: %s\n", opToString(opcode))
}

func opToString(op uint16) string {
	return fmt.Sprintf("%04X", op)
}

func (console *Console) EmulateCycle() {
	// fetch opcode
	var opcode uint16 = uint16(console.ram[console.cpu.pc])<<8 | uint16(console.ram[console.cpu.pc+1])
	// fmt.Println("executing " + opToString(opcode))

	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	kk := byte(opcode & 0x00FF)
	nnn := uint16(opcode & 0x0FFF)

	// decode opcode - switch on the first byte
	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0: // clear screen
			for r := int32(0); r < VHeight; r++ {
				for c := int32(0); c < VWidth; c++ {
					console.Video[r][c] = false
				}
			}
			console.cpu.pc += 2
		case 0x00EE: // return from subroutine
			console.cpu.pc = console.cpu.stack[console.cpu.sp]
			console.cpu.sp--
			console.cpu.pc += 2
		default:
			noop(opcode)
		}
		break
	case 0x1000: // 1NNN - jump to address nnn
		console.cpu.pc = nnn
		break
	case 0x2000: // 2NNN - call address
		console.cpu.sp++
		console.cpu.stack[console.cpu.sp] = console.cpu.pc
		console.cpu.pc = nnn
		break
	case 0x3000: // 3XKK - skip next instruction if vx = kk
		if console.cpu.v[x] == kk {
			console.cpu.pc += 2
		}
		console.cpu.pc += 2
		break
	case 0x4000: // 4XKK - skip next instruction if vx != kk
		if console.cpu.v[x] != kk {
			console.cpu.pc += 2
		}
		console.cpu.pc += 2
		break
	case 0x5000: // 5XY0 - skip next instruction if vx = vy
		if console.cpu.v[x] == console.cpu.v[y] {
			console.cpu.pc += 2
		}
		console.cpu.pc += 2
		break
	case 0x6000: // 6XKK - set vx = kk
		console.cpu.v[x] = kk
		console.cpu.pc += 2
		break
	case 0x7000: // 7XKK - set vx = vx + kk
		console.cpu.v[x] = console.cpu.v[x] + kk
		console.cpu.pc += 2
		break
	case 0x8000: // lots of instructions
		switch opcode & 0x000F {
		case 0x0: // LD vx, vy
			console.cpu.v[x] = console.cpu.v[y]
			break
		case 0x1:
			console.cpu.v[x] = console.cpu.v[x] | console.cpu.v[y]
			break
		case 0x2:
			console.cpu.v[x] = console.cpu.v[x] & console.cpu.v[y]
			break
		case 0x3:
			console.cpu.v[x] = console.cpu.v[x] ^ console.cpu.v[y]
			break
		case 0x4:
			result := uint16(console.cpu.v[x]) + uint16(console.cpu.v[y])
			if result > 255 {
				console.cpu.v[0xF] = 1
			}
			console.cpu.v[x] = byte(result & 0xFF)
			break
		case 0x5:
			if console.cpu.v[x] > console.cpu.v[y] {
				console.cpu.v[0xF] = 1
			}
			console.cpu.v[x] = console.cpu.v[x] - console.cpu.v[y]
			break
		case 0x6:
			if console.cpu.v[x]&0x1 == 1 {
				console.cpu.v[0xF] = 1
			} else {
				console.cpu.v[0xF] = 0
			}
			console.cpu.v[x] = console.cpu.v[x] / 2
			break
		default:
			noop(opcode)
			break
		}
		console.cpu.pc += 2
		break
	case 0x9000: // 9XY0 - skip next instruction if vx != vy
		if console.cpu.v[x] != console.cpu.v[y] {
			console.cpu.pc += 2
		}
		console.cpu.pc += 2
		break
	case 0xA000: // ANNN - set i = nnn
		console.cpu.i = nnn
		console.cpu.pc += 2
		break
	case 0xB000: // BNNN - jump to location nnn + v0
		console.cpu.pc = nnn + uint16(console.cpu.v[0])
		console.cpu.pc += 2
		break
	case 0xC000: // CXKK - set vx = random byte & kk
		console.cpu.v[x] = byte(rand.Intn(255))
		console.cpu.pc += 2
		break
	case 0xD000: // DXYN - Display n-byte sprite starting at memory location i
		// at (vx, vy), set vf = collision
		sWidth := byte(8)                // all sprites are 8 pixels wide
		sHeight := byte(opcode & 0x000F) // n bytes means n pixels tall

		vx := console.cpu.v[x]
		vy := console.cpu.v[y]

		console.cpu.v[0xF] = 0

		for row := byte(0); row < sHeight; row++ {
			sprite := console.ram[console.cpu.i+uint16(row)]
			for col := byte(0); col < sWidth; col++ {
				if (sprite & 0x80) > 0 {
					if console.Video[vy+row][vx+col] {
						console.cpu.v[0xF] = 1
					}
					console.Video[vy+row][vx+col] = !console.Video[vy+row][vx+col]
				}
				sprite <<= 1
			}
		}

		console.cpu.pc += 2
		break
	case 0xE000: // read input
		noop(opcode)
		console.cpu.pc += 2
		break
	case 0xF000: // lots of instructions
		switch opcode & 0x00FF {
		case 0x07:
			console.cpu.v[x] = console.cpu.dt
			break
		case 0x15:
			console.cpu.dt = console.cpu.v[x]
			break
		case 0x18:
			console.cpu.st = console.cpu.v[x]
			break
		case 0x1E:
			console.cpu.i += uint16(console.cpu.v[x])
			break
		case 0x29:
		case 0x33: // binary coded decimal conversion
			vx := console.cpu.v[x]
			hundreds := vx / 100
			tens := (vx / 10) % 10
			ones := (vx % 100) % 10
			console.ram[console.cpu.i] = hundreds
			console.ram[console.cpu.i+1] = tens
			console.ram[console.cpu.i+2] = ones
			break
		case 0x55:
		case 0x65:
		case 0x0A: // wait for keypress
		default:
			noop(opcode)
		}
		console.cpu.pc += 2
		break
	default:
		noop(opcode)
		console.cpu.pc += 2
	}
	// execute opcode
	// update timers
}
