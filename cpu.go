package goboy

import "strconv"

// cpu emulates the DMG micro-controller
type cpu struct {
	mem *memory // A pointer to the memory system
	A   byte
	B   byte
	C   byte
	D   byte
	E   byte
	F   byte // flags
	H   byte
	L   byte
	SP  uint16 // stack pointer
	PC  uint16 // program counter
}

// Emulates a CPU clock cycle
func (c *cpu) tick() {
	// TODO when a clock will be implemented
}

// HL register
func (c *cpu) HL() uint16 {
	return (uint16(c.H) << 8) + uint16(c.L)
}

// AF register
func (c *cpu) AF() uint16 {
	return (uint16(c.A) << 8) + uint16(c.F)
}

// BC register
func (c *cpu) BC() uint16 {
	return (uint16(c.B) << 8) + uint16(c.C)
}

// DE register
func (c *cpu) DE() uint16 {
	return (uint16(c.D) << 8) + uint16(c.E)
}

// Z flag
func (c *cpu) Z() bool {
	return (c.F & 0x80) == 0x80
}

// N flag
func (c *cpu) N() bool {
	return (c.F & 0x40) == 0x40
}

// Hy flag
func (c *cpu) Hy() bool {
	return (c.F & 0x20) == 0x20
}

// Cy flag
func (c *cpu) Cy() bool {
	return (c.F & 0x10) == 0x10
}

func (c *cpu) setFlags(z bool, n bool, h bool, cy bool) {
	var value byte
	if z {
		value += 0x80
	}
	if n {
		value += 0x40
	}
	if h {
		value += 0x20
	}
	if cy {
		value += 0x10
	}
	c.F = value
}

func (c *cpu) load8(address uint16) byte {
	value := c.mem.Read(address)
	c.tick()
	return value
}

func (c *cpu) load8PC() byte {
	value := c.load8(c.PC)
	c.PC++
	return value
}

func (c *cpu) load16(address uint16) uint16 {
	high := c.load8(address)
	low := c.load8(address + 1)
	return uint16(high)*0x100 + uint16(low)
}

func (c *cpu) load16PC() uint16 {
	value := c.load16(c.PC)
	c.PC = c.PC + 2
	return value
}

func (c *cpu) write8(address uint16, value byte) {
	c.mem.Write(address, value)
	c.tick()
}

// Helper that store a value in a register designed by its rank.
// (from 0 to 7) B C D E H L HL A
// Most of the time, the value will be a byte, but we need int for the generic case
func (c *cpu) seti(register int, value int) {
	switch register {
	case 0:
		c.B = byte(value)
	case 1:
		c.C = byte(value)
	case 2:
		c.D = byte(value)
	case 3:
		c.E = byte(value)
	case 4:
		c.H = byte(value)
	case 5:
		c.L = byte(value)
	case 6:
		c.write8(c.HL(), byte(value))
	case 7:
		c.A = byte(value)
	default:
		panic("Wrong value")
	}
}

// Helper that load a common value specified by its rank.
// (from 0 to 7) B C D E H L (HL) A
func (c *cpu) geti(register int) int {
	switch register {
	case 0:
		return int(c.B)
	case 1:
		return int(c.C)
	case 2:
		return int(c.D)
	case 3:
		return int(c.E)
	case 4:
		return int(c.H)
	case 5:
		return int(c.L)
	case 6:
		return int(c.load8(c.HL()))
	case 7:
		return int(c.A)
	default:
		panic("Wrong value")
	}
}

func (c *cpu) processOpcode() {
	opcode := c.load8PC()

	// HALT (must be done before the LD group)
	if opcode == 0x76 {
		// TODO
		return
	}

	// Common LD operations
	if 0x40 <= opcode && opcode < 0x80 {
		src := int(opcode>>3) - 4
		dst := int(opcode & 7)
		c.seti(src, dst)
		return
	}

	// Other opcodes
	switch opcode {
	case 0: // NOP
		return
	case 0x06: // LD B, n
		c.B = c.load8PC()
	case 0x0e: // LD C, n
		c.C = c.load8PC()
	case 0x16: // LD D, n
		c.D = c.load8PC()
	case 0x1e: // LD E, n
		c.E = c.load8PC()
	case 0x26: // LD H, n
		c.H = c.load8PC()
	case 0x2e: // LD L, n
		c.L = c.load8PC()
	case 0x31: // LD SP, nn
		c.SP = c.load16PC()
	default:
		panic("Unknown opcode 0x" + strconv.FormatInt(int64(opcode), 16))
	}
}
