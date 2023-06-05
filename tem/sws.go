package tem

import (
	"fmt"
	"math"

	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmaths"
)

// SubwatershedsBifuricate converts a TEM to an index of topologically-ordered subwatersheds of given area [m²]
func (t *TEM) SubwatershedsBifuricate(gd *grid.Definition, area float64) (map[int]int, int) {
	cc := t.ContributingCellCounts()

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

	ws = dissolve(gd, ws, cc, t.Downslopes(), area/2)

	cnt := make(map[int]int)
	for _, g := range ws {
		cnt[g]++
	}
	return ws, len(cnt)
}

// SubwatershedsHeadwater returns a map of mostly equal catchments to a given area, prioritizing headwaters
func (t *TEM) SubwatershedsHeadwater(gd *grid.Definition, area float64) (map[int]int, int) {
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

	ws = dissolve(gd, ws, cc, ds, area/2)

	cnt := make(map[int]int)
	for _, g := range ws {
		cnt[g]++
	}
	return ws, len(cnt)
}

func dissolve(gd *grid.Definition, ws, cc, ds map[int]int, area float64) map[int]int {
	fmt.Println("  dissolving smaller sws with larger")

	thrsh := int(area / gd.Cwidth / gd.Cwidth)         //int(1000 * 1000 / gd.Cwidth / gd.Cwidth)
	minv := func(m map[int]int, n int) map[int][]int { // invert maps
		o := make(map[int][]int, n)
		for c, g := range m {
			if _, ok := o[g]; !ok {
				o[g] = []int{c}
			} else {
				o[g] = append(o[g], c)
			}
		}
		return o
	}
	// crwl := gd.ToCrawler()
	// iws, _, ng := crwl.CrawlByInt(ws, false)
	// // iws = func() map[int]int {
	// // 	o := make(map[int]int, len(iws))
	// // 	for c, g := range iws {
	// // 		o[c] = g
	// // 	}
	// // 	return o
	// // }()
	// msws := minv(iws, ng)
	gcoll := make(map[int]int)
	for _, v := range ws {
		gcoll[v]++
	}
	msws := minv(ws, len(gcoll))
	remap := make(map[int]int)
	for g, a := range msws {
		if len(a) <= thrsh {
			pp := func() int {
				x, cx := 0, -1
				for _, c := range a {
					if cc[c] > x {
						x = cc[c]
						cx = c
					}
				}
				return cx
			}()
			if d, ok := ds[pp]; ok {
				remap[g] = ws[d] // iws[d]
			} else {
				remap[g] = -1
			}
		}
	}

	// iws, mbrd, ng := crwl.CrawlByInt(ws, false)
	// msws := minv(iws, ng)
	// q := []int{}
	// for g, a := range msws {
	// 	if v, ok := mbrd[g]; !ok || len(v) == 0 {
	// 		panic("borderless sws found")
	// 	}
	// 	if len(a) <= thrsh {
	// 		q = append(q, g)
	// 	}
	// }
	// for len(q) > 0 {
	// 	g := q[0]
	// 	q = q[1:]

	// 	gc := make(map[int]int)
	// 	for _, b := range mbrd[g] {
	// 		gc[iws[b]]++
	// 	}
	// 	gx, nx := -1, 0
	// 	for g, n := range gc {
	// 		if n > nx {
	// 			nx = n
	// 			gx = g
	// 		}
	// 	}
	// 	msws[gx] = append(msws[gx], msws[g]...)
	// 	delete(msws, g)
	// }

	dAlt := make(map[int]int)
	for g, a := range msws {
		newg := func(g int) int {
			o := -1
			var recurs func(int)
			recurs = func(g int) {
				if r, ok := remap[g]; ok {
					recurs(r)
				} else {
					o = g
				}
			}
			recurs(g)
			return o
		}(g)
		for _, c := range a {
			dAlt[c] = newg
		}
	}

	// create new indx
	nws := make(map[int]int, len(ws))
	for c, i := range ws {
		if v, ok := dAlt[c]; ok {
			nws[c] = v
		} else {
			nws[c] = i
		}
	}
	return nws
}
