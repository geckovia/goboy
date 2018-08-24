package goboy

import "testing"

// A Helper to check memory
func (m *memory) Assert(address uint16, value byte, t *testing.T) {
	result := m.Read(address)
	if result != value {
		t.Error("Expected", value, "got", result)
	}
}

// Test that we can write and read a simple byte in memory
func TestSimpleWrite(t *testing.T) {
	m := memory{}
	m.Assert(0xc100, 0, t)
	m.Write(0xc100, 42)
	m.Assert(0xc100, 42, t)
}

// Test a ROM loading
func TestRomLoading(t *testing.T) {
	m := memory{}
	m.Write(0xc100, 42)
	data := make([]byte, 100000)
	for i := 0; i < 100000; i++ {
		data[i] = byte(i/1000 + 1)
	}
	m.LoadRom(&data)
	m.bootDisabled = true
	m.Assert(0, 1, t)
	m.Assert(0x3fff, 17, t)
	m.Assert(0x4000, 17, t)
	m.Assert(0xc100, 42, t)
}

// Test Echo memory segment
func TestEchoMemory(t *testing.T) {
	m := memory{}
	m.Write(0xc000, 42)
	m.Write(0xe001, 69)
	m.Assert(0xc000, 42, t)
	m.Assert(0xe000, 42, t)
	m.Assert(0xc001, 69, t)
	m.Assert(0xe001, 69, t)
}

func TestMBC1ChangeROM(t *testing.T) {
	m := memory{}
	data := make([]byte, 100000)
	for i := 0; i < 100000; i++ {
		data[i] = byte(i/1024 + 1)
	}
	data[0x147] = 1 // MBC1
	m.LoadRom(&data)
	m.bootDisabled = true
	m.Assert(0, 1, t)
	m.Assert(0x4000, 17, t)
	m.Write(0x2000, 3) // Change to ROM bank 3
	m.Assert(0x4000, 49, t)
}
