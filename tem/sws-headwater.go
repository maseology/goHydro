package tem

import (
	"fmt"

	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmaths"
)

// SubwatershedsHeadwater returns a map of mostly equal catchments to a given area, prioritizing headwaters
func (t *TEM) SubwatershedsHeadwater(gd *grid.Definition, area float64) map[int]int {
	cc := t.ContributingCellCounts()
	ds := t.Downslopes()

	ws := func() map[int]int {
		cxc := make(map[int]int, len(cc))
		for k, v := range cc {
			cxc[k] = v // making copy as this gets adjusted below
		}
		cntx := int(area / gd.Cwidth / gd.Cwidth)
		opk := func() []int {
			fmt.Println("  getting ordered peaks")
			ct := t.concentrationTime()
			pct := make(map[int]int)
			for _, p := range t.Peaks(-1) {
				pct[p] = ct[p]
			}
			ii, oo := mmaths.SortMapInt(pct)
			_ = oo
			for i, j := 0, len(ii)-1; i < j; i, j = i+1, j-1 {
				ii[i], ii[j] = ii[j], ii[i]
			}
			return ii
		}()

		drain := func(i int) []int {
			var recurs func(int)
			c := make(map[int]int)
			recurs = func(i int) {
				c[i] = 1
				if d, ok := ds[i]; ok && d > -1 {
					recurs(d)
				}
			}
			o := make([]int, 0, len(c))
			for i := range c {
				o = append(o, i)
			}
			return o
		}

		var watershed func(int)
		col := make(map[int]int)
		watershed = func(i int) {
			cc0 := cxc[i]
			for us := range t.climb(i) {
				if _, ok := col[us]; !ok {
					col[us] = i
				}
			}
			for _, ds := range drain(i) {
				if _, ok := col[ds]; ok {
					panic("Subwatersheds.watershed err1")
				}
				cxc[ds] -= cc0
			}
		}

		fmt.Println("  building sws")
		for _, p := range opk {
			if _, ok := col[p]; ok {
				continue
			}
			cid := p
			// println(cid)
			for {
				if cxc[cid] < cntx {
					if _, ok := ds[cid]; !ok { // farfield
						break
					}
					if ds[cid] < 0 { // farfield
						break
					}
					cid = ds[cid]
					continue
				}
				break
			}
			watershed(cid)
		}
		return col
	}()

	return ws
}
