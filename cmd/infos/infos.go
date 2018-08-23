package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Get the ROM Title
func Title(rom []byte) string {
	return string(rom[0x134:0x143])
}

// Is the ROM a GameBoy Color game?
func IsColor(rom []byte) bool {
	return rom[0x143] == 0x80
}

// License information
func License(rom []byte) string {
	license := rom[0x14b]
	switch license {
	case 0x79:
		return "Accolade"
	case 0xa4:
		return "Konami"
	case 0x33:
		return string(rom[0x144:0x146])
	}
	panic("Unexpected license")
}

// The cardtridge destination: Japan or not
func Destination(rom []byte) string {
	if rom[0x14a] != 0 {
		return "Japanese"
	}
	return "Non-Japanese"
}

// Cartdridge type. I will add values as I meet them.
func CartridgeType(rom []byte) string {
	switch rom[0x147] {
	case 0:
		return "ROM ONLY"
	case 1:
		return "ROM+MBC1"
	case 0x13:
		return "ROM+MBC3+RAM+BATT"
	default:
		return string(rom[0x147])
	}
}

// Size in bytes of the ROM (the memory) of the ROM (the game).
func ROMSize(rom []byte) int {
	value := int(rom[0x148])
	size := 32 * 1024
	for i := 0; i < value; i++ {
		size *= 2
	}
	return size
}

// Size in bytes of the RAM.
func RAMSize(rom []byte) int {
	value := int(rom[0x149])
	if value == 0 {
		return 0
	}
	size := 2048
	for i := 0; i < value-1; i++ {
		size *= 4
	}
	return size
}

func main() {
	dat, err := ioutil.ReadFile(os.Args[1])
	check(err)
	fmt.Println("ROM information\n---------------")
	fmt.Println("Title:\t\"" + Title(dat) + "\"")
	fmt.Println("GameBoy Color:\t", IsColor(dat))
	fmt.Println("License:\t", License(dat))
	fmt.Println("Destination:\t", Destination(dat))
	fmt.Println("Cartridge type:\t", CartridgeType(dat))
	fmt.Println("ROM size:\t", ROMSize(dat)/1024, "kB")
	fmt.Println("RAM size:\t", RAMSize(dat)/1024, "kB")
}
