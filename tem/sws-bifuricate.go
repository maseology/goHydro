package tem

import (
	"math"

	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmaths"
)

// SubwatershedsBifuricate converts a TEM to an index of topologically-ordered subwatersheds of given area [m²]
func (t *TEM) SubwatershedsBifuricate(gd *grid.Definition, cc map[int]int, area float64) map[int]int {
	// cc := t.ContributingCellCounts()

	thrsh := int(area / gd.Cwidth / gd.Cwidth)

	cs := func() []int {
		bpnts := make(map[int]int)
		var climb func(int)
		croot := -1
		climb = func(cid int) {
			p := make(map[int]int)
			for _, c := range t.USlp[cid] {
				if cc[c] > thrsh {
					p[c] = cc[cid] - cc[c]
				}
				climb(c)
			}
			if len(p) > 1 { // condition 1: Bifuricate when more than one branch points to a contributing are greater than specified
				if cc[croot]-cc[cid] < thrsh { // condition 2: continue with smallest
					px, cx := math.MaxInt, -1
					for c, v := range p {
						if v < px {
							px = v
							cx = c
						}
					}
					for c := range p {
						if c == cx {
							continue
						}
						bpnts[c] = cc[c]
					}
				} else {
					for c := range p {
						bpnts[c] = cc[c]
					}
				}
			}
		}
		for _, c := range t.Outlets() {
			bpnts[c] = cc[c]
			croot = c
			climb(c)
		}
		cs, _ := mmaths.SortMapInt(bpnts)
		return cs
	}()

	ws := func() map[int]int {
		ws := make(map[int]int, len(t.TEC))
		var climb func(int)
		cid0 := -1
		climb = func(cid int) {
			ws[cid] = cid0
			for _, u := range t.USlp[cid] {
				if _, ok := ws[u]; !ok {
					climb(u)
				}
			}
		}
		for _, c := range cs {
			cid0 = c
			climb(c)
		}
		return ws
	}()

	return ws
}
