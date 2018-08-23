package goboy

import "testing"

func TestBoot(t *testing.T) {
	mem := Memory{}
	data := make([]byte, 0x4000)
	mem.LoadRom(&data)
	cpu := CPU{}
	cpu.mem = &mem

	cpu.run()
}
