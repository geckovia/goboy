package goboy

type Memory struct {
	mem [0xffff]byte // TODO: optimize space, just because we could
    mbc MBC
}


func (m *Memory) LoadRom(rom *[]byte) {	
    switch (*rom)[0x147] {
    case 0:
        m.mbc = loadMBC0(rom)
    case 1:
        m.mbc = loadMBC1(rom)
    default:
        panic("Unknown cartridge type")
    }
}

func (m *Memory) Read(address uint16) byte {
	switch {
    case 0xa000 <= address && address < 0xc000:  // 8kB Switchable RAM bank
        return m.mbc.read(address)
    case address < 0x8000:  // 32kB Cartridge
        return m.mbc.read(address)
    default:
        return m.mem[address]
    }
}

func (m *Memory) Write(address uint16, value byte) {
    switch {
    case 0xe000 <= address && address < 0xfe00:  // Echo of 8kB Internal RAM
        m.mem[address] = value
        m.mem[address - 0x2000] = value
    case 0xc000 <= address && address < 0xe000:  // 8kB Internal RAM
        m.mem[address] = value
        m.mem[address + 0x2000] = value
    case 0xa000 <= address && address < 0xc000:  // 8kB Switchable RAM bank
        fallthrough
    case address < 0x8000:  // 32kB Cartridge
        m.mbc.write(address, value)
    default:
        m.mem[address] = value
    }
}
