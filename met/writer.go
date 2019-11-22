package met

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// Writer is a struct used to write .met files
type Writer struct {
	f *os.File
}

// NewWriter creates a new writer struct
func NewWriter(filepath string, h *Header) (*Writer, error) {
	f, err := os.OpenFile(filepath, os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	w := Writer{
		f: f,
	}
	if err := w.writeHead(h); err != nil {
		return nil, err
	}
	return &w, nil
}

// Close met.Writer
func (w *Writer) Close() { w.f.Close() }

func (w *Writer) writeHead(h *Header) error {
	// version 0001
	chk := func(err error) error {
		return fmt.Errorf("met.writeHead failed: %v", err)
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, h.v); err != nil {
		return chk(err)
	}
	if err := binary.Write(buf, binary.LittleEndian, h.uc); err != nil {
		return chk(err)
	}
	if err := binary.Write(buf, binary.LittleEndian, h.tc); err != nil {
		return chk(err)
	}
	if err := binary.Write(buf, binary.LittleEndian, h.wbdc); err != nil {
		return chk(err)
	}
	if err := binary.Write(buf, binary.LittleEndian, h.prc); err != nil {
		return chk(err)
	}
	if err := binary.Write(buf, binary.LittleEndian, h.intvl); err != nil {
		return chk(err)
	}
	if h.intvl > 0 {
		if err := binary.Write(buf, binary.LittleEndian, h.dtb.Unix()); err != nil {
			return chk(err)
		}
		if err := binary.Write(buf, binary.LittleEndian, h.dte.Unix()); err != nil {
			return chk(err)
		}
	}
	if err := binary.Write(buf, binary.LittleEndian, h.lc); err != nil {
		return chk(err)
	}
	if err := binary.Write(buf, binary.LittleEndian, h.ESPG); err != nil {
		return chk(err)
	}

	if h.nloc > 0 {
		if err := binary.Write(buf, binary.LittleEndian, h.nloc); err != nil {
			return chk(err)
		}
		if h.lc == 1 {
			for k := range h.Locations {
				if err := binary.Write(buf, binary.LittleEndian, int32(k)); err != nil {
					return chk(err)
				}
			}
		} else {
			return fmt.Errorf("writer.go wrtieHead TODO")
		}
	}

	if _, err := w.f.Write(buf.Bytes()); err != nil {
		return chk(err)
	}
	return nil
}

// Add adds data to file
func (w *Writer) Add(data ...interface{}) error {
	buf := new(bytes.Buffer)
	for _, v := range data {
		if err := binary.Write(buf, binary.LittleEndian, v); err != nil {
			return fmt.Errorf("met.Add failed: %v", err)
		}
	}
	if _, err := w.f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("met.Add write failed: %v", err)
	}
	return nil
}
