package met

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/maseology/goHydro/grid"

	"github.com/maseology/mmio"
)

// ReadHead collect met header information
func (h *Header) readHead(b *bytes.Reader) {
	// version 0001
	h.v = mmio.ReadUInt16(b)
	h.uc = mmio.ReadUInt8(b)
	h.tc = mmio.ReadUInt8(b)
	h.wbdc = mmio.ReadUInt64(b)
	if h.wbdc == 0 {
		log.Panicf("waterbalance data code %d currently not supported\n", h.wbdc)
	}
	h.wbl = WBcodeToMap(h.wbdc)
	h.prc = mmio.ReadInt8(b)
	h.intvl = mmio.ReadUInt64(b)
	if h.intvl > 0 {
		h.dtb = time.Unix(mmio.ReadInt64(b), 0).UTC()
		h.dte = time.Unix(mmio.ReadInt64(b), 0).UTC()
	}
	h.lc = mmio.ReadInt8(b)
	h.ESPG = mmio.ReadUInt32(b)
}

// ReadLoc collects met location information
func (h *Header) readLoc(b *bytes.Reader) {
	if h.lc == 0 {
		h.nloc = mmio.ReadUInt32(b)
	} else if h.lc > 0 {
		h.nloc = mmio.ReadUInt32(b)
		h.Locations = make(map[int][]interface{})
		if h.lc == 1 {
			if h.nloc == 1 {
				h.Locations[0] = []interface{}{mmio.ReadInt32(b)}
			} else {
				for i := 0; i < int(h.nloc); i++ {
					h.Locations[int(mmio.ReadInt32(b))] = []interface{}{mmio.ReadInt32(b)}
				}
			}
		} else if h.lc == 12 {
			log.Panicln("h.locations-build: CODE NOT CHECKED YET")
			for i := 0; i < int(h.nloc); i++ {
				h.Locations[int(mmio.ReadInt32(b))] = []interface{}{mmio.ReadFloat64(b), mmio.ReadFloat64(b)}
			}
		} else {
			log.Panicf("location code %d currently not supported\n", h.lc)
		}
	} else {
		log.Panicf("location code %d currently not supported\n", h.lc)
	}
}

func (h *Header) check() error {
	cver := uint16(1) // current version
	if h.v != cver {
		return fmt.Errorf("Error: not the current supported .met version--found: %04d; want: %04d", h.v, cver)
	}
	return nil
}

// ReadMET reads a .met blob
func ReadMET(fp string, print bool) (*Header, *Coll, error) {
	b := mmio.OpenBinary(fp)
	var h Header
	h.readHead(b)
	h.readLoc(b)
	if print {
		fmt.Printf("\n File: %s\n", filepath.Base(fp))
		h.Print()
	}
	if err := h.check(); err != nil {
		return nil, nil, err
	}

	// read data
	iwbl, nwbl := h.WBDCkeys()
	nan := func(v float64) float64 { // no data handler
		if v == -9999.0 {
			return math.NaN()
		}
		return v
	}

	ts := time.Second * time.Duration(h.intvl)
	col := Coll{T: make([]time.Time, h.Nstep()), D: make([][][]float64, h.Nstep())}
	// dc := make(map[time.Time]map[int]map[int]float64, h.Nstep())
	if h.prc == 8 {
		// for d := h.dtb; !d.After(h.dte); d = d.Add(ts) {
		// 	dc[d] = make(map[int]map[int]float64, h.nloc)
		// 	for i := 0; i < int(h.nloc); i++ {
		// 		dc[d][i] = make(map[int]float64, len(iwbl))
		// 		for _, j := range iwbl {
		// 			dc[d][i][int(j)] = nan(mmio.ReadFloat64(b))
		// 		}
		// 	}
		// }
		k := 0
		for dt := h.dtb; !dt.After(h.dte); dt = dt.Add(ts) {
			col.T[k] = dt
			a := make([]float64, int(h.nloc)*len(iwbl))
			if err := binary.Read(b, binary.LittleEndian, &a); err != nil {
				log.Fatalf(" met.ReadMET failed: %v", err)
			}
			col.D[k] = make([][]float64, int(h.nloc))
			c := 0
			for i := 0; i < int(h.nloc); i++ {
				col.D[k][i] = make([]float64, nwbl)
				for j := 0; j < nwbl; j++ {
					col.D[k][i][j] = nan(a[c]) // [date ID][cell ID][type ID]
					c++
				}
			}
			k++
		}
	} else if h.prc == 4 {
		k := 0
		for dt := h.dtb; !dt.After(h.dte); dt = dt.Add(ts) {
			// fmt.Println(d)
			// dc[d] = make(map[int]map[int]float64, h.nloc)
			// for i := 0; i < int(h.nloc); i++ {
			// 	dc[d][i] = make(map[int]float64, len(iwbl))
			// 	for _, j := range iwbl {
			// 		dc[d][i][int(j)] = nan(float64(mmio.ReadFloat32(b)))
			// 		cnt++
			// 	}
			// }
			// fmt.Println(cnt)
			col.T[k] = dt
			a := make([]float32, int(h.nloc)*len(iwbl))
			if err := binary.Read(b, binary.LittleEndian, &a); err != nil {
				log.Fatalf(" met.ReadMET failed: %v", err)
			}
			col.D[k] = make([][]float64, int(h.nloc))
			c := 0
			for i := 0; i < int(h.nloc); i++ {
				col.D[k][i] = make([]float64, nwbl)
				for j := 0; j < nwbl; j++ {
					col.D[k][i][j] = nan(float64(a[c])) // [date ID][cell ID][type ID]
					c++
				}
			}
			k++
		}
	} else {
		return nil, nil, fmt.Errorf(" met.ReadMET error: unknown data type")
	}
	return &h, &col, nil
}

// ReadRaw reads raw binary, returning a map
func ReadRaw(fp string, print bool) (*Header, map[time.Time]map[int]float64, error) {
	var h Header

	gdefToHeader := func(gd *grid.Definition) (Header, error) {
		var h Header
		h.v = 1

		// Locations  map[int][]interface{}
		// v          uint16            // version
		// uc, tc     uint8             // unit code, time code, location code
		// wbdc       uint64            // waterbalance data code
		// wbl        map[uint64]string // waterbalance data map
		// prc, lc    int8              // precision, location code
		// intvl      uint64            // timestep interval [s]
		// dtb, dte   time.Time
		// ESPG, nloc uint32
		return h, nil
	}

	switch ext := mmio.GetExtension(fp); ext {
	case ".f16":
		gd, err := grid.ReadGDEF(fp+".gdef", print)
		if err != nil {
			return nil, nil, fmt.Errorf("MET.ReadRaw: ReadGDEF error: %v", err)
		}
		h, err = gdefToHeader(gd)
		if err != nil {
			return nil, nil, fmt.Errorf("MET.gdefToHeader: grid definition error: %v", err)
		}
	default:
		return nil, nil, fmt.Errorf("MET.ReadRaw: unrecognized file type ext: %s", ext)
	}
	if _, ok := mmio.FileExists(fp + ".gdef"); !ok {
		return nil, nil, fmt.Errorf("MET.ReadRaw: required file not found: %s", fp+".gdef")
	}

	if err := h.check(); err != nil {
		return nil, nil, err
	}
	if print {
		fmt.Printf("\n File: %s\n", filepath.Base(fp))
		h.Print()
	}
	// b := mmio.OpenBinary(fp)

	// // read data
	// dt := time.Second * time.Duration(h.intvl)
	dc := make(map[time.Time]map[int]float64)
	// iwbl := func() []uint64 {
	// 	keys := make([]uint64, 0, len(h.wbl))
	// 	for k := range h.wbl {
	// 		keys = append(keys, k)
	// 	}
	// 	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	// 	return keys
	// }()

	// nan := func(v float64) float64 { // no data handler
	// 	if v == -9999.0 {
	// 		return math.NaN()
	// 	}
	// 	return v
	// }
	// if h.prc == 8 {
	// 	for d := h.dtb; !d.After(h.dte); d = d.Add(dt) {
	// 		dc[d] = make(map[int]float64)
	// 		for _, i := range iwbl {
	// 			dc[d][int(i)] = nan(mmio.ReadFloat64(b))
	// 		}
	// 	}
	// } else if h.prc == 4 {
	// 	for d := h.dtb; !d.After(h.dte); d = d.Add(dt) {
	// 		dc[d] = make(map[int]float64)
	// 		for _, i := range iwbl {
	// 			dc[d][int(i)] = nan(float64(mmio.ReadFloat32(b)))
	// 		}
	// 	}
	// } else {
	// 	log.Fatalf(" met.ReadMET error: unknown data type\n")
	// }

	return &h, dc, nil
}

// ReadBigMET reads a .met blob in chunks to be more memory-conservative
func ReadBigMET(fp string, print bool) (*Header, *Coll, error) {
	f, err := os.Open(fp)
	if err != nil {
		log.Fatalf("ReadBigMET failed to open file, error: %v\n", err)
	}
	defer f.Close()

	var h Header
	bh := make([]byte, 42)
	if _, err := f.Read(bh); err != nil {
		log.Fatalf("ReadBigMET failed to read file header, error: %v\n", err)
	}
	h.readHead(bytes.NewReader(bh))
	if h.intvl == 0 {
		log.Fatalf("ReadBigMET error: un-specified interval not supported\n")
	}

	bh = make([]byte, h.locationSize())
	if _, err := f.Read(bh); err != nil {
		log.Fatalf("ReadBigMET failed to read file locations, error: %v\n", err)
	}
	h.readLoc(bytes.NewReader(bh))
	if print {
		fmt.Printf("\n Reading .met File: %s\n", filepath.Base(fp))
		h.Print()
	}
	if err := h.check(); err != nil {
		return nil, nil, err
	}

	// read data
	nan := func(v float64) float64 { // no data handler
		if v == -9999.0 {
			return math.NaN()
		}
		return v
	}

	nwbl := len(h.wbl)
	ts := time.Second * time.Duration(h.intvl)
	col := Coll{T: make([]time.Time, h.Nstep()), D: make([][][]float64, h.Nstep())}
	if h.prc == 8 {
		fmt.Print("TODO")
	} else if h.prc == 4 {
		k := 0
		for dt := h.dtb; !dt.After(h.dte); dt = dt.Add(ts) {
			col.T[k] = dt
			bd := make([]byte, int(h.nloc)*nwbl*4)
			if _, err := f.Read(bd); err != nil {
				log.Fatalf("ReadBigMET failed to read data, error: %v\n", err)
			}
			rd := bytes.NewReader(bd)

			a, c := make([]float32, int(h.nloc)*nwbl), 0
			if err := binary.Read(rd, binary.LittleEndian, &a); err != nil {
				log.Fatalf("ReadBigMET failed (float32): %v", err)
			}
			col.D[k] = make([][]float64, int(h.nloc))
			for i := 0; i < int(h.nloc); i++ {
				col.D[k][i] = make([]float64, nwbl)
				for j := 0; j < nwbl; j++ {
					col.D[k][i][j] = nan(float64(a[c]))
					c++
				}
			}
			k++
		}
	} else {
		return nil, nil, fmt.Errorf(" met.ReadMET error: unknown data type")
	}
	return &h, &col, nil
}
