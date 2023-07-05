package tem

import "github.com/maseology/goHydro/grid"

func SubwatershedDissolve(gd *grid.Definition, ws, cc, ds map[int]int, area float64) map[int]int {
	// fmt.Println("  dissolving smaller sws with larger")

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
