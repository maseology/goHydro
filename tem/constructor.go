package tem

import (
	"encoding/binary"
	"fmt"
	"path/filepath"

	"github.com/maseology/mmio"
)

// New contructor
func (t *TEM) New(fp string) error {
	fmt.Printf(" loading: %s\n", fp)

	switch filepath.Ext(fp) {
	case ".uhdem", ".bin":
		if err := t.loadUHDEM(fp); err != nil {
			return err
		}
	default:
		return fmt.Errorf(" error: unknown TEM file type used")
	}

	t.buildUpslopes()
	return nil
}

func (t *TEM) loadUHDEM(fp string) error {
	// load file
	buf := mmio.OpenBinary(fp)

	// check file type
	switch mmio.ReadString(buf) {
	case "unstructured":
		// do nothing
	default:
		return fmt.Errorf("Fatal error: unsupported UHDEM filetype")
	}

	// read dem data
	var nc int32
	binary.Read(buf, binary.LittleEndian, &nc) // number of cells
	t.TECs = make(map[int]TEC, nc)
	for i := int32(0); i < nc; i++ {
		u := uhdemReader{}
		u.uhdemRead(buf)
		t.TECs[int(u.I)] = u.toTEC()
	}

	// read flowpaths
	var nfp int32
	binary.Read(buf, binary.LittleEndian, &nfp) // number of flowpaths
	for i := int32(0); i < nfp; i++ {
		f := fpReader{}
		f.fpRead(buf)
		var x = t.TECs[int(f.I)]
		x.Ds = int(f.Ids)
		t.TECs[int(f.I)] = x
	}

	if mmio.ReachedEOF(buf) {
		return nil
	}
	return fmt.Errorf("Fatal error: UHDEM file contains extra data")
}

func (t *TEM) buildUpslopes() {
	t.us = make(map[int][]int)
	for i, v := range t.TECs {
		if v.Ds >= 0 {
			t.us[v.Ds] = append(t.us[v.Ds], i)
		}
	}
}
