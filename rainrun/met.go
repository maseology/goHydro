package rainrun

import (
	"encoding/gob"
	"log"
	"os"
	"sort"
	"time"

	"github.com/maseology/goHydro/met"
	"github.com/maseology/mmio"
)

// HDR holds header info
var HDR *met.Header

// FRC holds forcing data
var FRC [][]float64

// Ndt number of timesteps
var Ndt int

// DT holds dates
var DT []time.Time

// DOY hold day of year
var DOY []int

// Timestep timestep in seconds
var Timestep float64

// Loc contains location info (coordinates, catchment properties, etc.)
var Loc []float64

// LoadMET collect the climate data, set to a global variable
func LoadMET(fp string, print bool) {
	switch mmio.GetExtension(fp) {
	case ".met":
		loadMet(fp, print)
	case ".gob":
		loadGob(fp)
	default:
		log.Fatalf("unknown input data file %s", fp)
	}

	Ndt = len(DT)
	DOY = make([]int, Ndt)
	for i, t := range DT {
		DOY[i] = t.YearDay()
	}
}

func loadGob(fp string) {
	f, err := os.Open(fp)
	defer f.Close()
	if err != nil {
		log.Fatalf("met.go loadGob error: %v", err)
	}
	enc := gob.NewDecoder(f)
	err = enc.Decode(&FRC)
	if err != nil {
		log.Fatalf("met.go loadGob error: %v", err)
	}
	err = enc.Decode(&DT)
	if err != nil {
		log.Fatalf("met.go loadGob error: %v", err)
	}
}

func loadMet(fp string, print bool) {
	Ndt, FRC, HDR = func() (int, [][]float64, *met.Header) {
		h, c, err := met.ReadMET(fp, print)
		if err != nil {
			log.Fatalln(err)
		}
		if h.Nloc() != 1 {
			log.Fatalln("error: currently on simgle-location .met files supported")
		}

		Timestep = h.IntervalSec()
		DT = make([]time.Time, 0, len(c.T))
		for _, t := range c.T {
			DT = append(DT, t)
		}
		sort.Slice(DT, func(i, j int) bool { return DT[i].Before(DT[j]) })

		afrc := make([][]float64, 0, len(DT))
		switch h.WBCD {
		case 33554486:
			for i := range DT {
				afrc = append(afrc, []float64{c.D[i][0][0], c.D[i][0][1], c.D[i][0][2], c.D[i][0][3], c.D[i][0][4]})
			}
		case 33555968:
			log.Fatalf("met.go LoadMET(): FIX CODE\n")
			// chk := func(d time.Time, i int, p string) float64 {
			// 	if _, ok := dc[d]; !ok {
			// 		log.Fatalln(d, "not included in met file")
			// 	}
			// 	if v1, ok := dc[d][i]; ok {
			// 		return v1
			// 	}
			// 	log.Fatalln(p, "not included in met file")
			// 	return math.NaN()
			// }

			// for _, d := range dt { // [timeID][0][TypeID]
			// 	v := make([]float64, 3)
			// 	v[0] = chk(d, met.AtmosphericYield, "AtmosphericYield")
			// 	v[1] = chk(d, met.AtmosphericDemand, "AtmosphericDemand")
			// 	v[2] = chk(d, met.UnitDischarge, "UnitDischarge")
			// 	afrc = append(afrc, v)
			// }
		default:
			log.Fatalf("rainrun/inout/met.go LoadMET() error: WBCD code not supported: %d\n", h.WBCD)
		}

		switch h.LocationCode() {
		case 1:
			Loc = []float64{h.Locations[0][0].(float64)}
		case 16:
			for k, v := range h.Locations {
				Loc = make([]float64, 7)
				Loc[0] = float64(k) // cell id
				for i := 1; i < 7; i++ {
					Loc[i] = v[i-1].(float64) // x,y,z,gradient,aspect,area
				}
			}
		default:
			log.Fatalf("rainrun/inout/met.go LoadMET() error: location code not supported: %d\n", h.LocationCode())
		}

		return len(afrc), afrc, h
	}()
}
