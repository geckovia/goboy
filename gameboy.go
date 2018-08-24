package goboy

// gameBoy is the structure emulating a gameBoy console
type gameBoy struct {
	CPU    *cpu
	Memory *memory
}

// NewGameBoy constructs a GameBoy
func NewGameBoy(rom *[]byte) gameBoy {
	gb := gameBoy{}

	c := cpu{}
	gb.CPU = &c

	mem := memory{}
	mem.LoadRom(rom)
	c.mem = &mem
	gb.Memory = &mem

	return gb
}

// Run the emulator
func (gb *gameBoy) Run() {
	for {
		gb.CPU.processOpcode()
	}
}
