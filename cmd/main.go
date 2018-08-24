package main

import (
	"io/ioutil"
	"os"

	"github.com/geckovia/goboy"
)

func main() {
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	gb := goboy.NewGameBoy(&data)
	gb.Run()
}
