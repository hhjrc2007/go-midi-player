package main

import (
	"encoding/binary"
	"errors"
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

func readTrack(r io.Reader) ([]byte, error) {
	for {
		var chunk struct {
			ID     [4]byte
			Length uint32
		}
		if err := binary.Read(r, binary.BigEndian, &chunk); err != nil {
			return nil, err
		}
		if string(chunk.ID[:]) != "MTrk" {
			if _, err := io.CopyN(io.Discard, r, int64(chunk.Length)); err != nil {
				return nil, err
			}
			continue
		}
		data := make([]byte, chunk.Length)
		if _, err := io.ReadFull(r, data); err != nil {
			return nil, err
		}
		return data, nil
	}
}

type Event struct {
	Delta  uint32 // ticks since the previous event in the same track
	Status byte   // 0x80-0xEF channel message, 0xF0/0xF7 sysex, 0xFF meta
	Meta   byte   // meta event type, only set when Status is 0xFF
	Data   []byte
}

// readVarLen decodes a variable-length quantity and reports how many
// bytes it took up.
func readVarLen(data []byte) (value uint32, n int, err error) {
	for n < len(data) && n < 4 {
		b := data[n]
		n++
		value = value<<7 | uint32(b&0x7f)
		if b&0x80 == 0 {
			return value, n, nil
		}
	}
	return 0, 0, errors.New("bad variable-length quantity")
}

func parseTrack(data []byte) ([]Event, error) {
	var events []Event
	var running byte
	pos := 0
	for pos < len(data) {
		delta, n, err := readVarLen(data[pos:])
		if err != nil {
			return nil, err
		}
		pos += n
		if pos >= len(data) {
			return nil, io.ErrUnexpectedEOF
		}

		status := data[pos]
		if status < 0x80 {
			// A data byte where a status should be: running status.
			if running == 0 {
				return nil, fmt.Errorf("data byte %#x with no running status", status)
			}
			status = running
		} else {
			pos++
		}

		ev := Event{Delta: delta, Status: status}
		var length int
		switch {
		case status < 0xF0:
			running = status
			length = 2
			// Program change and channel pressure carry one data byte.
			if kind := status & 0xF0; kind == 0xC0 || kind == 0xD0 {
				length = 1
			}
		case status == 0xFF || status == 0xF0 || status == 0xF7:
			running = 0
			if status == 0xFF {
				if pos >= len(data) {
					return nil, io.ErrUnexpectedEOF
				}
				ev.Meta = data[pos]
				pos++
			}
			l, n, err := readVarLen(data[pos:])
			if err != nil {
				return nil, err
			}
			pos += n
			length = int(l)
		default:
			return nil, fmt.Errorf("unexpected status byte %#x", status)
		}

		if pos+length > len(data) {
			return nil, io.ErrUnexpectedEOF
		}
		ev.Data = data[pos : pos+length]
		pos += length
		events = append(events, ev)
	}
	return events, nil
}
