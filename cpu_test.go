package goboy

import (
	"testing"
)

func TestBoot(t *testing.T) {
	mem := memory{}
	data := make([]byte, 0x4000)
	mem.loadRom(&data)
	processor := cpu{}
	processor.mem = &mem

	for processor.PC != 0xe0 {
		processor.processOpcode()
	}
}
