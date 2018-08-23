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

func title(rom []byte) string {
	return string(rom[0x134:0x143])
}

func isColor(rom []byte) bool {
	return rom[0x143] == 0x80
}

func license(rom []byte) string {
	license := rom[0x14b]
	switch license {
	case 0x79:
		return "Accolade"
	case 0xa4:
		return "Konami"
	case 0x33:
		return string(rom[0x144:0x146])
	default:
		return string("Unknown")
	}
}

func destination(rom []byte) string {
	if rom[0x14a] != 0 {
		return "Japanese"
	}
	return "Non-Japanese"
}

func cartridgeType(rom []byte) string {
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

func romSize(rom []byte) int {
	value := int(rom[0x148])
	size := 32 * 1024
	for i := 0; i < value; i++ {
		size *= 2
	}
	return size
}

func ramSize(rom []byte) int {
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
	fmt.Println("Title:\t\"" + title(dat) + "\"")
	fmt.Println("GameBoy Color:\t", isColor(dat))
	fmt.Println("License:\t", license(dat))
	fmt.Println("Destination:\t", destination(dat))
	fmt.Println("Cartridge type:\t", cartridgeType(dat))
	fmt.Println("ROM size:\t", romSize(dat)/1024, "kB")
	fmt.Println("RAM size:\t", ramSize(dat)/1024, "kB")
}
