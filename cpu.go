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

type registerOperation func(c *cpu, value int) int

// Helper that load a common value specified by its rank.
// (from 0 to 7) B C D E H L (HL) A
func (c *cpu) getReg(register int) byte {
	switch register {
	case 0:
		return c.B
	case 1:
		return c.C
	case 2:
		return c.D
	case 3:
		return c.E
	case 4:
		return c.H
	case 5:
		return c.L
	case 6:
		return c.load8(c.HL())
	case 7:
		return c.A
	default:
		panic("Wrong value")
	}
}

// Helper that apply an operation to a register and affect the result to a register
// specified by its rank: (from 0 to 7) B C D E H L (HL) A.
// Most of the time, the value will be a byte, but we need int for the generic case
func (c *cpu) applyOp(destRank int, srcRank int, op registerOperation) {
	result := byte(op(c, int(c.getReg(srcRank))))
	switch destRank {
	case 0:
		c.B = result
	case 1:
		c.C = result
	case 2:
		c.D = result
	case 3:
		c.E = result
	case 4:
		c.H = result
	case 5:
		c.L = result
	case 6:
		c.write8(c.HL(), result)
	case 7:
		c.A = result
	default:
		panic("Wrong value")
	}
}

func ld(c *cpu, value int) int {
	return value
}

func add(c *cpu, value int) int {
	a := int(c.A)
	sum := a + value
	carry := sum > 0xff
	halfCarry := (a&0xf)+(value&0xf) > 0xf
	c.setFlags(sum == 0, false, halfCarry, carry)
	return sum
}

func addc(c *cpu, value int) int {
	if c.Cy() {
		value++
	}
	return add(c, value)
}

func sub(c *cpu, value int) int {
	a := int(c.A)
	diff := a - value
	carry := diff < 0
	halfCarry := (a&0xf)-(value&0xf) < 0
	c.setFlags(diff == 0, true, halfCarry, carry)
	return diff
}

func sbc(c *cpu, value int) int {
	if c.Cy() {
		value++
	}
	return sub(c, value)
}

func and(c *cpu, value int) int {
	result := int(c.A & byte(value))
	c.setFlags(result == 0, false, true, false)
	return result
}

func or(c *cpu, value int) int {
	result := int(c.A | byte(value))
	c.setFlags(result == 0, false, false, false)
	return result
}

func xor(c *cpu, value int) int {
	result := int(c.A ^ byte(value))
	c.setFlags(c.A == 0, false, false, false)
	return result
}

func cp(c *cpu, value int) int {
	sub(c, value)
	return value
}

func bit(n uint) registerOperation {
	return func(c *cpu, value int) int {
		present := (value & (1 << n)) != 0
		c.setFlags(present, false, true, c.Cy())
		return value
	}
}

func set(n uint) registerOperation {
	return func(c *cpu, value int) int {
		return value | (1 << n)
	}
}

func res(n uint) registerOperation {
	return func(c *cpu, value int) int {
		return value &^ (1 << n)
	}
}

func rlc(c *cpu, value int) int {
	carry := byte(value) >> 7
	result := byte(value) << 1
	result += carry
	c.setFlags(result == 0, false, false, carry == 1)
	return int(result)
}

var arithmeticOps = []registerOperation{add, addc, sub, sbc, and, xor, or, cp}

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
		dst := int((opcode - 0x40) >> 3)
		src := int(opcode & 7)
		c.applyOp(dst, src, ld)
		return
	}

	// Common arithmetic
	if 0x80 <= opcode && opcode <= 0xc0 {
		op := int((opcode - 0x80) >> 3)
		src := int(opcode & 7)
		c.applyOp(7, src, arithmeticOps[op])
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

//var cbOps = []registerOperation{}

// CB-prefixed opcodes
func (c *cpu) cb() {
	opcode := c.load8PC()

	switch {
	case 0x40 <= opcode && opcode < 0x80:
		n := uint((opcode - 0x40) >> 3)
		reg := int(opcode & 7)
		c.applyOp(reg, reg, bit(n))
	case 0x80 <= opcode && opcode < 0xc0:
		n := uint((opcode - 0x80) >> 3)
		reg := int(opcode & 7)
		c.applyOp(reg, reg, res(n))
	case 0xc0 <= opcode && opcode < 0xf0:
		n := uint((opcode - 0xc0) >> 3)
		reg := int(opcode & 7)
		c.applyOp(reg, reg, set(n))
	default:
		panic("Unknown opcode 0xbc 0x" + strconv.FormatInt(int64(opcode), 16))
	}
}
