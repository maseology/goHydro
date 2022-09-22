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

func (g *GMET) CheckAndPrint() {
	fmt.Printf("\n  N stations %d\n", g.Nsta)
	fmt.Printf("  N timesteps %d\n", g.Nts)
	fmt.Printf("  startdate: %v\n", g.Ts[0])
	fmt.Printf("  end date: %v\n\n", g.Ts[g.Nts-1])
	g.check()
}

func (g *GMET) check() bool {
	if len(g.Sids) != g.Nsta {
		log.Fatalf("GMET.check Error: nsta\n")
	}
	if len(g.Ts) != g.Nts {
		log.Fatalf("GMET.check Error: nts\n")
	}
	ndays := g.Ts[g.Nts-1].Sub(g.Ts[0]).Seconds()/86400. + 1
	if g.Nts != int(ndays) {
		log.Fatalf("GMET.check Error: nts!=ndays\n")
	}

	for i := 0; i < g.Nts-1; i++ {
		if g.Ts[i+1].Sub(g.Ts[i]).Seconds() != 86400. {
			log.Fatalf("GMET.check consecutive date error: %s %s\n", g.Ts[i], g.Ts[i+1])
		}
	}
	return true
}
