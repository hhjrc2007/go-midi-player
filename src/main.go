package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "usage: go-midi-player file.mid")
        os.Exit(1)
    }
    fmt.Println("would play:", os.Args[1])
}
