package met

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"github.com/maseology/mmio"
)

type header struct {
	locations  map[int][]interface{}
	v          uint16            // version
	uc, tc, lc uint8             // unit code, time code, location code
	wbdc       uint64            // waterbalance data code
	wbl        map[uint64]string // waterbalance data map
	prc        int8              // precision
	intvl      uint64
	dtb, dte   time.Time
	espg, nloc uint32
}

func (h *header) Read(b *bytes.Reader) {
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
	h.lc = mmio.ReadUInt8(b)
	h.espg = mmio.ReadUInt32(b)
	if h.lc == 0 {
		h.nloc = mmio.ReadUInt32(b)
	} else if h.lc > 0 {
		h.nloc = mmio.ReadUInt32(b)
		h.locations = make(map[int][]interface{})
		if h.lc == 12 {
			log.Panicln("h.locations-build: CODE NOT CHECKED YET")
			for i := 0; i < int(h.nloc); i++ {
				h.locations[int(mmio.ReadInt32(b))] = []interface{}{mmio.ReadFloat64(b), mmio.ReadFloat64(b)}
			}
		} else {
			log.Panicf("location code %d currently not supported\n", h.lc)
		}
	}
}

func (h *header) Check() error {
	cver := uint16(1) // current version
	if h.v != cver {
		return fmt.Errorf("Error: not the current supported .met version--found: %04d; want: %04d", h.v, cver)
	}
	return nil
}

func (h *header) Print() {
	fmt.Printf("\nVersion %04d\n", h.v)
	fmt.Printf("unit code %d\n", h.uc)
	fmt.Printf("time code %d\n", h.tc)
	fmt.Printf("water-balance types: %v\n", h.wbl)
	fmt.Printf("data precision %d\n", h.prc)
	fmt.Printf("interval %d\n", h.intvl)
	if h.intvl > 0 {
		fmt.Println("start date " + h.dtb.Format("2006-01-02"))
		fmt.Println("end date " + h.dte.Format("2006-01-02")) // 15:04:05"))
	}
	fmt.Printf("location code %d\n", h.lc)
	fmt.Printf("coordinate projection (ESPG) %d\n", h.espg)
	if h.lc == 0 {
		fmt.Printf("n cells %d\n\n", h.nloc)
	} else if h.lc > 0 {
		fmt.Printf("n locations %d\n", h.nloc)
		for i, c := range h.locations {
			fmt.Println(i, c)
		}
	}
}

// ReadMET reads a .met blob, returning a map
func ReadMET(fp string) (map[time.Time]map[int]float64, error) {
	b := mmio.OpenBinary(fp)
	var h header
	h.Read(b)
	h.Print()
	if err := h.Check(); err != nil {
		return nil, err
	}

	// read data
	dt := time.Second * time.Duration(h.intvl)
	dc := make(map[time.Time]map[int]float64)
	iwbl := func() []uint64 {
		keys := make([]uint64, 0, len(h.wbl))
		for k := range h.wbl {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
		return keys
	}()

	nan := func(v float64) float64 { // no data handler
		if v == -9999.0 {
			return math.NaN()
		}
		return v
	}
	if h.prc == 8 {
		for d := h.dtb; !d.After(h.dte); d = d.Add(dt) {
			dc[d] = make(map[int]float64)
			for _, i := range iwbl {
				dc[d][int(i)] = nan(mmio.ReadFloat64(b))
			}
		}
	} else if h.prc == 4 {
		for d := h.dtb; !d.After(h.dte); d = d.Add(dt) {
			dc[d] = make(map[int]float64)
			for _, i := range iwbl {
				dc[d][int(i)] = nan(float64(mmio.ReadFloat32(b)))
			}
		}
	} else {
		panic("unknown data type")
	}

	return dc, nil
}
