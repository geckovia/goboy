package goboy

// MBC is an interface for memory banks
type MBC interface {
	read(address uint16) byte
	write(address uint16, value byte)
}

// MBC0 emulates no MBC (32kB ROM only)
type MBC0 struct {
	rom [0x8000]byte
}

// MBC1 emulates the Memory Bank Controller 1
type MBC1 struct {
	rom       []byte
	ram       [0x8000]byte
	ramMode   bool
	writeable bool
	romBank   uint8
	ramBank   uint8
}

// No MBC (or "MBC0")
func loadMBC0(rom *[]byte) *MBC0 {
	m := &MBC0{}
	copy(m.rom[:], *rom)
	return m
}

func (m *MBC0) read(address uint16) byte {
	return m.rom[address]
}

func (m *MBC0) write(address uint16, value byte) {
	panic("Attempt to write in ROM")
}

// MBC1
func loadMBC1(rom *[]byte) *MBC1 {
	m := &MBC1{}
	m.rom = *rom
	m.romBank = 1
	return m
}

func (m *MBC1) read(address uint16) byte {
	switch {
	case address < 0x4000: // ROM bank 0
		return m.rom[address]
	case 0x4000 <= address && address < 0x8000: // Other ROM bank
		return m.rom[address+uint16(m.romBank-1)*0x4000]
	case 0xa000 <= address && address < 0xc000: // RAM bank
		return m.ram[address-0xa000+uint16(m.ramBank)*0x2000]
	default:
		panic("Invalid address to read")
	}
}

func (m *MBC1) write(address uint16, value byte) {
	switch {
	// Writeable RAM mode
	case address < 0x2000:
		mode := value & 0xf
		m.writeable = (mode == 10)
	// Change ROM bank
	case 0x2000 <= address && address < 0x4000:
		bank := value & 0x1f // keep last five bits
		if bank == 0 {       // special case
			bank = 1
		}
		m.romBank = (m.romBank & 0x60) + bank
	// Change RAM bank
	case 0x4000 <= address && address < 0x6000:
		bank := value & 0x3 // keep last two bits
		if m.ramMode {
			m.ramBank = bank
		} else { // The most significant bits of the ROM bank have changed.
			m.romBank = bank*0x20 + (m.romBank & 0x1f)
		}
	// Select RAM mode
	case 0x6000 <= address && address < 0x8000:
		m.ramMode = (value & 1) == 1
	// Write in RAM
	case 0xa000 <= address && address < 0xc000:
		if m.writeable {
			m.ram[address-0xa000+uint16(m.ramBank)*0x2000] = value
		} else {
			panic("Tried to write in protected RAM")
		}
	default:
		panic("Invalid address to write to")
	}
}
