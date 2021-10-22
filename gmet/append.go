package apigmet

import (
	"fmt"
	"log"
	"time"
)

func (g *GMET) Append(g1 *GMET) bool {
	// compare station list order
	for i, s := range g.Sids {
		if g1.Sids[i] != s {
			log.Fatalf("GMET.append somparision error: not the same station list")
		}
	}

	switch g1.Ts[0].Sub(g.Ts[0]) {
	case 0: // same starting point
		if g1.Nts == g.Nts {
			if g1.Ts[g1.Nts-1].Sub(g.Ts[g1.Nts-1]) == 0 {
				fmt.Println("  GMET.append: same data")
			} else {
				log.Fatalf("GMET.append error 1")
			}
		}
		return false
	default:
		g.check()
		g1.check()

		ndays := func(d time.Duration) int { return int(d.Seconds() / 86400.) }

		offset := ndays(g1.Ts[0].Sub(g.Ts[0]))
		if g.Ts[offset] != g1.Ts[0] {
			log.Fatalf("date error (g.Ts[offset] != g1.Ts[0]); %v %v", g.Ts[offset], g1.Ts[0])
		}

		switch nd := ndays(g1.Ts[g1.Nts-1].Sub(g.Ts[g.Nts-1])); {
		case nd <= 0: // update only
			for i := 0; i < g.Nsta; i++ {
				for j, v := range g1.Dat[i] {
					g.Dat[i][j+offset] = v
				}
			}
		default: // new data found
			ntsnew := ndays(g1.Ts[g1.Nts-1].Sub(g.Ts[0])) + 1
			tsnew := make([]time.Time, ntsnew)
			copy(tsnew, g.Ts)
			if tsnew[offset] != g1.Ts[0] {
				log.Fatalf("date error (tsnew[offset] != g1.Ts[0]); %v %v", g.Ts[offset], g1.Ts[0])
			}
			for i, t := range g1.Ts {
				tsnew[i+offset] = t
			}

			datnew := make([][]DSet, g.Nsta)
			copy(tsnew, g.Ts)
			for i := 0; i < g.Nsta; i++ {
				d := make([]DSet, ntsnew)
				copy(d, g.Dat[i])
				for j, v := range g1.Dat[i] {
					d[j+offset] = v
				}
				datnew[i] = d
			}

			g.Nts = ntsnew
			g.Ts = tsnew
			g.Dat = datnew
		}
	}
	return true
}
