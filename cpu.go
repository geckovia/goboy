package goboy

import (
	"strconv"
)

// cpu emulates the DMG micro-controller
type cpu struct {
	mem   *memory // A pointer to the memory system
	A     byte
	B     byte
	C     byte
	D     byte
	E     byte
	F     byte // flags
	H     byte
	L     byte
	SP    uint16 // stack pointer
	PC    uint16 // program counter
	start bool
}

// Emulates a CPU clock cycle
func (c *cpu) tick() {
	// TODO when a clock will be implemented
}

// HL register
func (c *cpu) HL() uint16 {
	return (uint16(c.H) << 8) + uint16(c.L)
}

func (c *cpu) setHL(value uint16) {
	c.H = byte(value >> 8)
	c.L = byte(value)
}

// AF register
func (c *cpu) AF() uint16 {
	return (uint16(c.A) << 8) + uint16(c.F)
}

func (c *cpu) setAF(value uint16) {
	c.A = byte(value >> 8)
	c.F = byte(value)
}

// BC register
func (c *cpu) BC() uint16 {
	return (uint16(c.B) << 8) + uint16(c.C)
}

func (c *cpu) setBC(value uint16) {
	c.B = byte(value >> 8)
	c.C = byte(value)
}

// DE register
func (c *cpu) DE() uint16 {
	return (uint16(c.D) << 8) + uint16(c.E)
}

func (c *cpu) setDE(value uint16) {
	c.D = byte(value >> 8)
	c.E = byte(value)
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

func (c *cpu) setFlags(z, n, h, cy bool) {
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
	high := c.load8(address + 1)
	low := c.load8(address)
	return uint16(high)<<8 + uint16(low)
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

func (c *cpu) write16(address uint16, value uint16) {
	c.write8(address, byte(value))
	c.write8(address+1, byte(value>>8))
}

func (c *cpu) jump(address uint16) {
	c.tick()
	c.PC = address
}

func (c *cpu) push(value uint16) {
	c.SP -= 2
	c.write16(c.SP, value)
	c.tick() // I don't understand why pushes takes 4 cycles more than a pop...
}

func (c *cpu) pop() uint16 {
	value := c.load16(c.SP)
	c.SP += 2
	return value
}

func (c *cpu) call(address uint16) {
	c.push(c.PC)
	c.PC = address // call is 24 cycles, so by using a push we can't also use a jump
}

func (c *cpu) ret() {
	c.jump(c.pop())
}

// Helper that store a value in a register designed by its rank.
// (from 0 to 7) B C D E H L HL A
// Most of the time, the value will be a byte, but we need int for the generic case
func (c *cpu) setReg(register int, value int) {
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
func (c *cpu) getReg(register int) int {
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

func (c *cpu) add(value int) {
	a := int(c.A)
	sum := a + value
	carry := sum > 0xff
	halfCarry := (a&0xf)+(value&0xf) > 0xf
	result := byte(sum)
	c.setFlags(result == 0, false, halfCarry, carry)
	c.A = result
}

func (c *cpu) addc(value int) {
	if c.Cy() {
		value++
	}
	c.add(value)
}

func (c *cpu) sub(value int) {
	c.A = c.cp(value)
}

func (c *cpu) sbc(value int) {
	if c.Cy() {
		value++
	}
	c.sub(value)
}

func (c *cpu) and(value int) {
	c.A &= byte(value)
	c.setFlags(c.A == 0, false, true, false)
}

func (c *cpu) or(value int) {
	c.A |= byte(value)
	c.setFlags(c.A == 0, false, false, false)
}

func (c *cpu) xor(value int) {
	c.A ^= byte(value)
	c.setFlags(c.A == 0, false, false, false)
}

func (c *cpu) cp(value int) byte {
	a := int(c.A)
	diff := a - value
	carry := diff < 0
	result := byte(diff)
	halfCarry := (a&0xf)-(value&0xf) < 0
	c.setFlags(result == 0, true, halfCarry, carry)
	return result
}

func (c *cpu) bit(n uint, reg int) {
	value := c.getReg(reg)
	present := (value & (1 << n)) != 0
	c.setFlags(present, false, true, c.Cy())
}

func (c *cpu) set(n uint, register int) {
	switch register {
	case 0:
		c.B = c.B | (1 << n)
	case 1:
		c.C = c.C | (1 << n)
	case 2:
		c.D = c.D | (1 << n)
	case 3:
		c.E = c.E | (1 << n)
	case 4:
		c.H = c.H | (1 << n)
	case 5:
		c.L = c.L | (1 << n)
	case 6:
		c.write8(c.HL(), c.load8(c.HL())|(1<<n))
	case 7:
		c.A = c.A | (1 << n)
	default:
		panic("Wrong value")
	}
}

func (c *cpu) res(n uint, register int) {
	switch register {
	case 0:
		c.B = c.B &^ (1 << n)
	case 1:
		c.C = c.C &^ (1 << n)
	case 2:
		c.D = c.D &^ (1 << n)
	case 3:
		c.E = c.E &^ (1 << n)
	case 4:
		c.H = c.H &^ (1 << n)
	case 5:
		c.L = c.L &^ (1 << n)
	case 6:
		c.write8(c.HL(), c.load8(c.HL())&^(1<<n))
	case 7:
		c.A = c.A &^ (1 << n)
	default:
		panic("Wrong value")
	}
}

// processCode emulates the fetching and processing
// of an instruction by the CPU
func (c *cpu) processOpcode() {
	opcode := c.load8PC()

	// HALT (must be done before the LD group)
	if opcode == 0x76 {
		// TODO
		return
	}

	// Common LD operations
	if 0x40 <= opcode && opcode < 0x80 {
		src := int((opcode - 0x40) >> 3)
		dst := int(opcode & 7)
		c.setReg(src, dst)
		return
	}

	// Common arithmetic
	if 0x80 <= opcode && opcode <= 0xc0 {
		op := int((opcode - 0x80) >> 3)
		src := int(opcode & 7)
		switch op {
		case 0:
			c.add(c.getReg(src))
		case 1:
			c.addc(c.getReg(src))
		case 2:
			c.sub(c.getReg(src))
		case 3:
			c.sbc(c.getReg(src))
		case 4:
			c.and(c.getReg(src))
		case 5:
			c.xor(c.getReg(src))
		case 6:
			c.or(c.getReg(src))
		case 7:
			c.cp(c.getReg(src))
		}
		return
	}

	// Other opcodes
	switch opcode {
	case 0x00: // NOP
		return
	case 0x01: // LD BC, nn
		c.setBC(c.load16PC())
	case 0x02: // LD (BC), A
		c.write8(c.BC(), c.A)
	case 0x06: // LD B, n
		c.B = c.load8PC()
	case 0x0a: // LD A, (BC)
		c.A = c.load8(c.BC())
	case 0x0c: // INC C
		c.C++
	case 0x0e: // LD C, n
		c.C = c.load8PC()
	case 0x11: // LD DE, nn
		c.setDE(c.load16PC())
	case 0x12: // LD (DE), A
		c.write8(c.DE(), c.A)
	case 0x16: // LD D, n
		c.D = c.load8PC()
	case 0x1a: // LD A, (DE)
		c.A = c.load8(c.DE())
	case 0x1e: // LD E, n
		c.E = c.load8PC()
	case 0x20: // JR NZ, r8
		address := c.PC + uint16(int8(c.load8PC()))
		if !c.Z() {
			c.jump(address)
		}
	case 0x21: // LD HL, nn
		c.setHL(c.load16PC())
	case 0x22: // LD (HL+), A
		hl := c.HL()
		c.write8(hl, c.A)
		c.setHL(hl + 1)
	case 0x26: // LD H, n
		c.H = c.load8PC()
	case 0x2a: // LD A, (HL+)
		hl := c.HL()
		c.A = c.load8(hl)
		c.setHL(hl + 1)
	case 0x2e: // LD L, n
		c.L = c.load8PC()
	case 0x30: // JR NC, r8
		address := c.PC + uint16(int8(c.load8PC()))
		if !c.Cy() {
			c.jump(address)
		}
	case 0x31: // LD SP, nn
		c.SP = c.load16PC()
	case 0x32: // LD (HL-), A
		hl := c.HL()
		c.write8(hl, c.A)
		c.setHL(hl - 1)
	case 0x36: // LD (HL), n
		c.write8(c.HL(), c.load8PC())
	case 0x3a: // LD A, (HL-)
		hl := c.HL()
		c.A = c.load8(hl)
		c.setHL(hl - 1)
	case 0x3e: // LD A, n
		c.A = c.load8PC()
	case 0xc1: // POP BC
		c.setBC(c.pop())
	case 0xc5: // PUSH BC
		c.push(c.BC())
	case 0xcb: // CB prefix
		c.cb()
	case 0xcd: // CALL nn
		destination := c.load16PC()
		c.call(destination)
	case 0xd1: // POP DE
		c.setDE(c.pop())
	case 0xd5: // PUSH DE
		c.push(c.DE())
	case 0xe0: // LDH n A
		c.write8(0xff00+uint16(c.load8PC()), c.A)
	case 0xe1: // POP HL
		c.setHL(c.pop())
	case 0xe2: // LD (C) A
		c.write8(0xff00+uint16(c.C), c.A)
	case 0xe5: // PUSH HL
		c.push(c.HL())
	case 0xf0: // LDH A n
		c.A = c.load8(0xff00 + uint16(c.load8PC()))
	case 0xf1: // POP AF
		c.setAF(c.pop())
	case 0xf2: // LD A (C)
		c.A = c.load8(0xff00 + uint16(c.C))
	case 0xf5: // PUSH AF
		c.push(c.AF())
	default:
		panic("Unknown opcode 0x" + strconv.FormatInt(int64(opcode), 16))
	}
}

// CB-prefixed opcodes
func (c *cpu) cb() {
	opcode := c.load8PC()

	switch {
	case 0x40 <= opcode && opcode < 0x80:
		n := uint((opcode - 0x40) >> 3)
		reg := int(opcode & 7)
		c.bit(n, reg)
	case 0x80 <= opcode && opcode < 0xc0:
		n := uint((opcode - 0x80) >> 3)
		reg := int(opcode & 7)
		c.res(n, reg)
	case 0xc0 <= opcode && opcode < 0xf0:
		n := uint((opcode - 0xc0) >> 3)
		reg := int(opcode & 7)
		c.set(n, reg)
	default:
		panic("Unknown opcode 0xbc 0x" + strconv.FormatInt(int64(opcode), 16))
	}
}
