package main

import (
      "bufio"
      "fmt"
      "os"
)

func main() {
      if len(os.Args) < 2 {
              fmt.Fprintln(os.Stderr, "usage: go-midi-player <file.mid>")
              os.Exit(1)
      }

      f, err := os.Open(os.Args[1])
      if err != nil {
              fmt.Fprintln(os.Stderr, err)
              os.Exit(1)
      }
      defer f.Close()

      h, err := readHeader(bufio.NewReader(f))
      if err != nil {
              fmt.Fprintln(os.Stderr, err)
              os.Exit(1)
      }
      fmt.Printf("format %d, %d tracks, division %d\n", h.Format, h.NumTracks, h.Division)
}
