package met

import (
	"fmt"
	"log"
	"sort"
	"time"
)

// Header contains metadata for the .met file
type Header struct {
	Locations  map[int][]interface{}
	v          uint16            // version
	uc, tc     uint8             // unit code, time code, location code
	WBCD       uint64            // waterbalance data code
	wbl        map[uint64]string // waterbalance data map
	prc, lc    int8              // precision, location code
	intvl      uint64            // timestep interval [s]
	dtb, dte   time.Time
	ESPG, nloc uint32
}

// NewHeader returns a header
func NewHeader(dtb, dte time.Time, intvl, nloc, prc int) Header {
	return Header{
		v:  1,
		uc: 1,
		tc: 1,

		dtb:   dtb,
		dte:   dte,
		intvl: uint64(intvl),
		nloc:  uint32(nloc),
		prc:   int8(prc),
	}
}

// IntervalSec returns the time interval of the .met file
func (h *Header) IntervalSec() float64 {
	return float64(h.intvl)
}

// Nloc returns the number of locations in the .met file
func (h *Header) Nloc() int {
	return int(h.nloc)
}

// LocationCode returns the location code of the .met file
func (h *Header) LocationCode() int {
	return int(h.lc)
}

// Nstep returns the number of timesteps in the .met file
func (h *Header) Nstep() int {
	n := h.dte.Add(time.Second*time.Duration(h.intvl)).Sub(h.dtb).Seconds() / float64(h.intvl)
	return int(n)
}

// BeginEndInterval returns the begining and end dates
func (h *Header) BeginEndInterval() (time.Time, time.Time, int64) {
	return h.dtb, h.dte, int64(h.intvl)
}

// Print .met metadata
func (h *Header) Print() {
	fmt.Printf(" Version %04d\n", h.v)
	fmt.Printf(" unit code %d\n", h.uc)
	fmt.Printf(" time code %d\n", h.tc)
	fmt.Printf(" water-balance types: %v\n", h.wbl)
	fmt.Printf(" data precision %d\n", h.prc)
	fmt.Printf(" interval %d\n", h.intvl)
	if h.intvl > 0 {
		fmt.Println(" start date " + h.dtb.Format("2006-01-02"))
		fmt.Println(" end date " + h.dte.Format("2006-01-02")) // 15:04:05"))
		fmt.Printf(" n steps %d\n", h.Nstep())
	}
	fmt.Printf(" location code %d\n", h.lc)
	fmt.Printf(" coordinate projection (ESPG) %d\n", h.ESPG)
	if h.lc == 0 {
		fmt.Printf(" n cells %d\n\n", h.nloc)
	} else if h.lc == 12 {
		fmt.Printf(" n locations %d\n\n", h.nloc)
		for i, c := range h.Locations {
			fmt.Printf("  %d %v\n", i, c)
		}
	} else if h.lc > 0 {
		if h.nloc == 1 {
			fmt.Printf(" outlet cell id %d\n\n", h.Locations[0][0])
		} else {
			fmt.Printf(" n locations %d\n\n", h.nloc)
			for i, c := range h.Locations {
				fmt.Printf("  %d %v\n", i, c)
			}
		}
	}
}

// Copy creates a deep copy of a Header
func (h *Header) Copy() *Header {
	hnew := *h
	newLoc := make(map[int][]interface{}, len(h.Locations))
	for k, v := range h.Locations {
		vs := make([]interface{}, len(v))
		copy(vs, v)
		newLoc[k] = vs
	}
	hnew.Locations = newLoc
	return &hnew
}

// SetWBDC changes the water budget data code
func (h *Header) SetWBDC(wbdc uint64) {
	if wbdc == 0 {
		log.Panicf("waterbalance data code %d currently not supported\n", wbdc)
	}
	h.WBCD = wbdc
	h.wbl = WBcodeToMap(wbdc)
}

// WBDCkeys returns an ordered key index associated with the waterbalance codes
func (h *Header) WBDCkeys() ([]uint64, int) {
	keys := make([]uint64, 0, len(h.wbl))
	for k := range h.wbl {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys, len(keys)
}

// WBDCxr return the zero-order array index of the waterbalance codes
func (h *Header) WBDCxr() map[string]int {
	keys := make([]uint64, 0, len(h.wbl))
	for k := range h.wbl {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	m := make(map[string]int, len(keys))
	for i, k := range keys {
		m[h.wbl[k]] = i
	}
	return m
}

// WBlist return the slice the waterbalance codes
func (h *Header) WBlist() []string {
	keys := make([]uint64, 0, len(h.wbl))
	for k := range h.wbl {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	s := make([]string, len(keys))
	for i, k := range keys {
		s[i] = h.wbl[k]
	}
	return s
}

// AddLocationIndex adds locations of code 1
func (h *Header) AddLocationIndex(iid int) {
	h.lc = 1
	h.nloc = 1
	h.Locations = make(map[int][]interface{}, 1)
	h.Locations[iid] = []interface{}{iid}
}
