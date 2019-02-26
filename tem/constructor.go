package tem

import (
	"encoding/binary"
	"fmt"
	"path/filepath"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

// New contructor
func (t *TEM) New(fp string) (map[int]mmaths.Point, error) {
	fmt.Printf(" loading: %s\n", fp)

	var coord map[int]mmaths.Point
	var err error
	switch filepath.Ext(fp) {
	case ".uhdem", ".bin":
		coord, err = t.loadUHDEM(fp)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf(" error: unknown TEM file type used")
	}

	t.buildUpslopes()
	return coord, nil
}

func (t *TEM) loadUHDEM(fp string) (map[int]mmaths.Point, error) {
	// load file
	buf := mmio.OpenBinary(fp)

	// check file type
	switch mmio.ReadString(buf) {
	case "unstructured":
		// do nothing
	default:
		return nil, fmt.Errorf("Fatal error: unsupported UHDEM filetype")
	}

	// read dem data
	var nc int32
	binary.Read(buf, binary.LittleEndian, &nc) // number of cells
	t.TEC = make(map[int]TEC, nc)
	coord := make(map[int]mmaths.Point, nc)
	// for i := int32(0); i < nc; i++ {
	// 	u := uhdemReader{}
	// 	u.uhdemRead(buf)
	// 	ii := int(u.I)
	// 	coord[ii], t.TEC[ii] = u.toTEC()
	// }
	uc := make([]uhdemReader, nc)
	if err := binary.Read(buf, binary.LittleEndian, uc); err != nil {
		return nil, fmt.Errorf("Fatal error: loadUHDEM uhdem read failed: %v", err)
	}
	for _, u := range uc {
		ii := int(u.I)
		coord[ii], t.TEC[ii] = u.toTEC()
	}

	// read flowpaths
	var nfp int32
	binary.Read(buf, binary.LittleEndian, &nfp) // number of flowpaths
	// for i := int32(0); i < nfp; i++ {
	// 	f := fpReader{}
	// 	f.fpRead(buf)
	// 	var x = t.TEC[int(f.I)]
	// 	x.Ds = int(f.Ids)
	// 	t.TEC[int(f.I)] = x
	// }
	fc := make([]fpReader, nfp)
	if err := binary.Read(buf, binary.LittleEndian, fc); err != nil {
		return nil, fmt.Errorf("Fatal error: loadUHDEM flowpath read failed: %v", err)
	}
	for _, f := range fc {
		ii := int(f.I)
		var x = t.TEC[ii]
		x.Ds = int(f.Ids)
		t.TEC[ii] = x
	}

	if mmio.ReachedEOF(buf) {
		return coord, nil
	}
	return nil, fmt.Errorf("Fatal error: UHDEM file contains extra data")
}

func (t *TEM) buildUpslopes() {
	t.us = make(map[int][]int)
	for i, v := range t.TEC {
		if v.Ds >= 0 {
			t.us[v.Ds] = append(t.us[v.Ds], i)
		}
	}
}