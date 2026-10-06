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
      r := bufio.NewReader(f)
      h, err := readHeader(r)
      if err != nil {
              fmt.Fprintln(os.Stderr, err)
              os.Exit(1)
      }
      fmt.Printf("format %d, %d tracks, division %d\n", h.Format, h.NumTracks, h.Division)

      for i := 0; i < int(h.NumTracks); i++ {
              data, err := readTrack(r)
              if err != nil {
                      fmt.Fprintf(os.Stderr, "track %d: %v\n", i, err)
                      os.Exit(1)
              }
              fmt.Printf("track %d: %d bytes, ends with % x\n", i, len(data), data[max(0, len(data)-3):])
      }
}
