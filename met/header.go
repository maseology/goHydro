package met

import (
	"fmt"
	"time"
)

// Header contains metadata for the .met file
type Header struct {
	Locations  map[int][]interface{}
	v          uint16            // version
	uc, tc     uint8             // unit code, time code, location code
	wbdc       uint64            // waterbalance data code
	wbl        map[uint64]string // waterbalance data map
	prc, lc    int8              // precision, location code
	intvl      uint64            // timestep interval [s]
	dtb, dte   time.Time
	ESPG, nloc uint32
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
	}
	fmt.Printf(" location code %d\n", h.lc)
	fmt.Printf(" coordinate projection (ESPG) %d\n", h.ESPG)
	if h.lc == 0 {
		fmt.Printf(" n cells %d\n\n", h.nloc)
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
