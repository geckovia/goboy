package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "goboy"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
    data, err := ioutil.ReadFile(os.Args[1])
    check(err)
    m := goboy.Memory{}
    m.LoadRom(&data)
    fmt.Println(m.Read(0x100))
    m.Write(0x100, 42)
    fmt.Println(m.Read(0x100))
}
