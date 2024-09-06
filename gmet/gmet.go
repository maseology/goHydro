package gmet

import (
	"fmt"
	"log"
	"time"
)

type DSet struct { // daily met set
	Date string
	Dat  []float64
}

type XYZ struct{ X, Y, Z float64 }

type GMET struct {
	Dat       [][]DSet    // [station][]
	Nts, Nsta int         // number timesteps/stations
	Ts        []time.Time // timesteps
	Sids      []int       // station IDs
	Sxy       []XYZ       // station coordinates
	Snams     []string    // station Names (or 'Dat' parameter names)
}

func (g *GMET) CheckAndPrint(timestepsec float64) {
	g.check(timestepsec)
	fmt.Printf("    n stations %d\n", g.Nsta)
	fmt.Printf("    n timesteps %d\n", g.Nts)
	fmt.Printf("    startdate: %v\n", g.Ts[0])
	fmt.Printf("    end date: %v\n", g.Ts[g.Nts-1])
}

func (g *GMET) check(timestepsec float64) bool {
	if len(g.Sids) != g.Nsta {
		log.Fatalf("GMET.check Error: nsta\n")
	}
	if len(g.Ts) != g.Nts {
		log.Fatalf("GMET.check Error: nts\n")
	}
	ndays := g.Ts[g.Nts-1].Sub(g.Ts[0]).Seconds()/timestepsec + 1
	if g.Nts != int(ndays) {
		log.Fatalf("GMET.check Error: nts!=ndays\n")
	}

	for i := 0; i < g.Nts-1; i++ {
		if g.Ts[i+1].Sub(g.Ts[i]).Seconds() != timestepsec {
			log.Fatalf("GMET.check consecutive date error: %s %s\n", g.Ts[i], g.Ts[i+1])
		}
	}

	tnow := time.Now()
	for {
		if g.Ts[g.Nts-1].After(tnow) {
			g.Nts--
			g.Ts = g.Ts[:g.Nts]
			for i := range g.Nsta {
				g.Dat[i] = g.Dat[i][:g.Nts]
			}
		} else {
			break
		}
	}
	return true
}
