package tem

import (
	"encoding/binary"
	"fmt"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

func (t *TEM) loadUHDEM(fp string) (map[int]mmaths.Point, map[int]int, error) {
	// load file
	buf := mmio.OpenBinary(fp)

	// check file type
	switch mmio.ReadString(buf) {
	case "unstructured":
		// do nothing
	default:
		return nil, nil, fmt.Errorf("fatal error: unsupported UHDEM filetype")
	}

	// read dem data
	var nc int32
	if err := binary.Read(buf, binary.LittleEndian, &nc); err != nil { // number of cells
		return nil, nil, fmt.Errorf("fatal error: loadUHDEM uhdem read failed: %v", err)
	}
	t.TEC = make(map[int]TEC, nc)
	coord := make(map[int]mmaths.Point, nc)
	uc := make([]uhdemReader, nc)
	if err := binary.Read(buf, binary.LittleEndian, uc); err != nil {
		return nil, nil, fmt.Errorf("fatal error: loadUHDEM uhdem read failed: %v", err)
	}
	for _, u := range uc {
		ii := int(u.I)
		coord[ii], t.TEC[ii] = u.toTEC()
	}

	// read flowpaths
	var nfp int32
	if err := binary.Read(buf, binary.LittleEndian, &nfp); err != nil { // number of flowpaths
		return nil, nil, fmt.Errorf("fatal error: loadUHDEM flowpath read failed: %v", err)
	}
	fc := make([]fpReader, nfp)
	if err := binary.Read(buf, binary.LittleEndian, fc); err != nil {
		return nil, nil, fmt.Errorf("fatal error: loadUHDEM flowpath read failed: %v", err)
	}
	dwnSlps := make(map[int]int, len(fc))
	for _, f := range fc {
		if f.Nds != 1 {
			return nil, nil, fmt.Errorf("fatal error: loadUHDEM TODO: many-to-one only allowed")
		}
		ii := int(f.I)
		dwnSlps[ii] = int(f.Ids)
		// var x = t.TEC[ii]
		// x.Ds = int(f.Ids)
		// t.TEC[ii] = x
	}

	if mmio.ReachedEOF(buf) {
		return coord, dwnSlps, nil
	}
	return nil, nil, fmt.Errorf("fatal error: UHDEM file contains extra data")
}

func (t *TEM) loadHDEM(fp string) (map[int]mmaths.Point, map[int]int, error) {
	// load file
	buf := mmio.OpenBinary(fp)

	// check file type
	typ := mmio.ReadString(buf)
	switch typ {
	case "grid":
		// read dem data
		var nc int32
		binary.Read(buf, binary.LittleEndian, &nc) // number of cells
		t.TEC = make(map[int]TEC, nc)
		coord := make(map[int]mmaths.Point, nc)
		uc := make([]uhdemReader, nc)
		if err := binary.Read(buf, binary.LittleEndian, uc); err != nil {
			return nil, nil, fmt.Errorf("fatal error: loadHDEM uhdem read failed: %v", err)
		}
		for _, u := range uc {
			ii := int(u.I)
			coord[ii], t.TEC[ii] = u.toTEC()
		}

		// read flowpaths
		var nfp int32
		binary.Read(buf, binary.LittleEndian, &nfp) // number of flowpaths
		fc := make([]fpReader, nfp)
		if err := binary.Read(buf, binary.LittleEndian, fc); err != nil {
			return nil, nil, fmt.Errorf("fatal error: loadHDEM flowpath read failed: %v", err)
		}
		ds := make(map[int]int, len(fc))
		for _, f := range fc {
			ii := int(f.I)
			ds[ii] = int(f.Ids)
			// var x = t.TEC[ii]
			// x.Ds = int(f.Ids)
			// t.TEC[ii] = x
		}

		if mmio.ReachedEOF(buf) {
			return coord, ds, nil
		}

	default:
		return nil, nil, fmt.Errorf("fatal error: unsupported HDEM filetype: '%s'", typ)

		// default:
		// 	// case "":
		// 	// old/raw version --- grid-based hdem's not working, use uhdem
		// 	if _, ok := mmio.FileExists(mmio.RemoveExtension(fp) + ".gdef"); !ok {
		// 		return nil, fmt.Errorf("fatal error: gdef required to read %s", fp)
		// 	}
		// 	gd, err := grid.ReadGDEF(mmio.RemoveExtension(fp)+".gdef", false)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	// nc := gd.Ncells()
		// 	nc := gd.Na
		// 	t.TEC = make(map[int]TEC, nc)
		// 	coord := make(map[int]mmaths.Point, nc)
		// 	hc := make([]hdemReader, nc)
		// 	buf = mmio.OpenBinary(fp) // re-open buf
		// 	if err := binary.Read(buf, binary.LittleEndian, hc); err != nil {
		// 		return nil, fmt.Errorf("fatal error: loadHDEM hdem read failed: %v", err)
		// 	}
		// 	for i, h := range hc {
		// 		if i < 0 {
		// 			print()
		// 		}
		// 		t.TEC[i] = h.toTEC()
		// 		coord[i] = gd.Coord[i]
		// 	}

		// 	if mmio.ReachedEOF(buf) {
		// 		return coord, nil
		// 	}

		// 	// read flowpaths
		// 	for {
		// 		var uid int32
		// 		var n uint8
		// 		binary.Read(buf, binary.LittleEndian, &uid) // flowpath upslope (from) cell
		// 		binary.Read(buf, binary.LittleEndian, &n)   // number of flowpaths
		// 		if uid < 0 {
		// 			fmt.Print()
		// 		}
		// 		for ii := 0; ii < int(n); ii++ {
		// 			var did int32
		// 			var frac float64

		// 			binary.Read(buf, binary.LittleEndian, &did)  // flowpath downslope (to) cell
		// 			binary.Read(buf, binary.LittleEndian, &frac) // flow fraction
		// 			if did < 0 {
		// 				fmt.Print()
		// 			}

		// 			var x = t.TEC[int(uid)]
		// 			x.Ds = int(did)
		// 			t.TEC[int(uid)] = x
		// 		}

		// 		if mmio.ReachedEOF(buf) {
		// 			return coord, nil
		// 		}
		// 	}
	}
	return nil, nil, fmt.Errorf("fatal error: HDEM file contains extra data")
}
