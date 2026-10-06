package main

import (
      "encoding/binary"
      "fmt"
      "io"
)

type Header struct {
      Format    uint16
      NumTracks uint16
      Division  uint16
}

func readHeader(r io.Reader) (Header, error) {
      var chunk struct {
              ID     [4]byte
              Length uint32
      }
      if err := binary.Read(r, binary.BigEndian, &chunk); err != nil {
              return Header{}, err
      }
      if string(chunk.ID[:]) != "MThd" {
              return Header{}, fmt.Errorf("not a MIDI file: got chunk %q", chunk.ID[:])
      }
      if chunk.Length < 6 {
              return Header{}, fmt.Errorf("header too short: %d bytes", chunk.Length)
      }

      var h Header
      if err := binary.Read(r, binary.BigEndian, &h); err != nil {
              return Header{}, err
      }
      if _, err := io.CopyN(io.Discard, r, int64(chunk.Length)-6); err != nil {
              return Header{}, err
      }
      return h, nil
}
