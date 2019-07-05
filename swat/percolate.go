package swat

import (
	"fmt"
	"math"
)

// percolate from soil zone (pg.151)
func (m *HRU) percolate() float64 {
	for i, ly := range m.sz {
		if !ly.frz {
			if m.iwt && i < nsl-1 && m.sz[i+1].sw <= m.sz[i+1].fc+(m.sz[i+1].sat-m.sz[i+1].fc)/2. {
				fmt.Println("high water table")
				// high water table, no percolation allowed
			} else {
				if ly.sw > ly.fc {
					swex := ly.sw - ly.fc                           // [mm]
					w := swex * (1. - math.Exp(-hoursperday/ly.tt)) // hard-coded to daily simulations
					if i < nsl-1 {
						if !m.sz[i+1].frz {
							m.sz[i].sw -= w
							m.sz[i+1].sw += w
						}
					} else {
						m.sz[i].sw -= w
						return w
					}
				}
			}
		}
	}
	return 0.
}
