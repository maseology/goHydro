package tem

import (
	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmaths"
)

func (t *TEM) SubwatershedSplit(gd *grid.Definition, ws, ds map[int]int, area float64) map[int]int {
	thrsh := int(area / gd.Cwidth / gd.Cwidth)

	// collect watersheds that require splitting
	gs := func(areafact int) map[int][]int {
		cnt := make(map[int]int)
		for _, g := range ws {
			cnt[g]++
		}
		gs := make(map[int][]int)
		for g, n := range cnt {
			if n > areafact*thrsh {
				gs[g] = make([]int, 0, n)
			}
		}
		for c, g := range ws {
			if _, ok := gs[g]; ok {
				gs[g] = append(gs[g], c)
			}
		}
		return gs
	}(3)

	contributingCellCounts := func(g0 int) map[int]int {
		mcnt := make(map[int]int, len(t.TEC))
		var climb func(int) int
		climb = func(cid int) int {
			mcnt[cid] = 1
			if us, ok := t.USlp[cid]; ok {
				for _, p := range us {
					if ws[p] != g0 {
						continue
					}
					if _, ok := mcnt[p]; ok {
						continue
					}
					mcnt[cid] += climb(p)
				}
			}
			return mcnt[cid]
		}
		climb(g0)
		return mcnt
	}
	concentrationTime := func(g0 int) map[int]int {
		cnt := make(map[int]int, len(t.TEC))
		var climb func(int, int)
		climb = func(i, c int) {
			cnt[i] = c
			if us, ok := t.USlp[i]; ok {
				for _, p := range us {
					if ws[p] != g0 {
						continue
					}
					if _, ok := cnt[p]; ok {
						continue
					}
					climb(p, c+1)
				}
			}
		}
		climb(g0, 1)
		return cnt
	}

	newsws := make(map[int]int, len(ws))
	for k, v := range ws {
		newsws[k] = v
	}
	for gOrig, a := range gs {
		if gOrig < 0 {
			continue
		}
		sz := len(a)
		newthrsh := sz / (sz / thrsh)
		newa := make(map[int]int, sz)
		ic, _ := mmaths.SortMapInt(concentrationTime(gOrig))
		cc := contributingCellCounts(gOrig)
		watershed := func(gNew int) {
			var climb func(int)
			climb = func(cid int) {
				if _, ok := newa[cid]; !ok {
					newa[cid] = gNew
					for _, us := range t.USlp[cid] {
						if ws[us] == gOrig {
							climb(us)
						}
					}
				}
			}
			climb(gNew)

			if ccadj, ok := cc[gNew]; ok {
				var drain func(int)
				drain = func(c int) {
					if _, ok := newa[c]; !ok {
						if _, ok := cc[c]; !ok {
							panic("tem.split err3.2")
						}
						cc[c] -= ccadj
						if d, ok := ds[c]; ok {
							if ws[d] == gOrig {
								drain(d)
							}
						}
					}
				}
				drain(gNew)
			} else {
				panic("tem.split err3.1")
			}
		}

		for i := sz - 1; i >= 0; i-- {
			cid := ic[i]
			if _, ok := newa[cid]; !ok {
				if ccc, ok := cc[cid]; ok {
					if ccc > 2*newthrsh {
						if us, ok := t.USlp[cid]; ok {
							for _, p := range us {
								if ug, ok := ws[p]; ok {
									if ug != gOrig {
										continue
									}
								} else {
									panic("tem.split err6")
								}
								if ccp, ok := cc[p]; ok {
									if ccp > newthrsh/3 {
										watershed(p)
									}
								} else {
									panic("tem.split err4")
								}
							}
						}
					} else if ccc > thrsh {
						watershed(cid)
					}
				} else {
					panic("tem.split err2")
				}
			}
		}

		for c, g := range newa {
			if _, ok := newsws[c]; !ok {
				panic("tem.split err5")
			}
			newsws[c] = g
		}
	}

	// for gOrig, a := range gs {

	// 	sz := len(a)
	// 	nsws := sz / thrsh
	// 	newthrsh := sz / nsws

	// 	newa := make(map[int]int, sz)

	// 	if c0 := func() int {
	// 		gcnt := make(map[int]int, sz)
	// 		var climb func(int) int
	// 		climb = func(cid int) int {
	// 			gcnt[cid] = 1
	// 			if us, ok := t.USlp[cid]; ok {
	// 				for _, p := range us {
	// 					if ws[p] != gOrig {
	// 						continue
	// 					}
	// 					if _, ok := gcnt[p]; ok {
	// 						continue
	// 					}
	// 					gcnt[cid] += climb(p)
	// 				}
	// 			}
	// 			return gcnt[cid]
	// 		}
	// 		climb(gOrig)

	// 		ic, oc := mmaths.SortMapInt(gcnt)
	// 		for i, c := range ic {
	// 			if oc[i] > newthrsh {
	// 				return c
	// 			}
	// 		}
	// 		return -1
	// 	}(); c0 < 0 {
	// 		panic("tem.split err1")
	// 	} else {
	// 		watershed(c0)
	// 	}

	// }

	// var climb func(int)
	// for g0 := range gs {
	// 	nc := 0
	// 	climb = func(cid int) {
	// 		if us, ok := t.USlp[cid]; ok {
	// 			sort.Ints(us)

	// 			for _, u := range t.USlp[cid] {
	// 				if ws[u] == g0 {

	// 				}
	// 			}
	// 		}
	// 	}
	// 	climb(g0)
	// }

	// splitter := func(a []int, g0 int) {
	// 	// sz := len(a)
	// 	// acc := make(map[int]int, sz)
	// 	// for _, c := range a {
	// 	// 	acc[c] = cc[c]
	// 	// }
	// 	// icc, ncc := mmaths.SortMapInt(acc)
	// 	// for i := sz - 2; i >= 0; i-- {

	// 	// }
	// 	var climb func(int)

	// }

	// for g, a := range gs {

	// 	splitter(a, g)

	// }

	// for {
	// 	gs := make(map[int][]int)
	// 	for c, g := range ws {
	// 		if _, ok := gs[g]; !ok {
	// 			gs[g] = []int{}
	// 		}
	// 		gs[g] = append(gs[g], c)
	// 	}
	// 	cnt := 0
	// 	var climb func(int)
	// 	for g, a := range gs {
	// 		sz := len(a)
	// 		if sz > thrsh {
	// 			col := make(map[int]int, sz)
	// 			for _, c := range a {
	// 				col[c] = cc[c]
	// 			}
	// 			cs, ccnt := mmaths.SortMapInt(col)
	// 			for i := sz - 2; i >= 0; i-- {
	// 				if ccnt[sz-1]-ccnt[i] < sz/2 {
	// 					if dc, ok := ds[cs[i]]; ok {
	// 						if dc == g {
	// 							continue
	// 						}
	// 						climb = func(cid int) {
	// 							for _, u := range t.USlp[cid] {
	// 								if ws[u] == g {
	// 									ws[u] = dc
	// 									climb(u)
	// 								}
	// 							}
	// 						}
	// 						climb(cs[i])
	// 					} else {
	// 						panic("split err1")
	// 					}
	// 					break
	// 				}
	// 			}
	// 			cnt++
	// 		}
	// 	}
	// 	println(cnt)
	// 	if cnt == 0 {
	// 		break
	// 	}

	// 	// cnt := make(map[int]int)
	// 	// for _, g := range ws {
	// 	// 	cnt[g]++
	// 	// }
	// 	// gs := make(map[int][]int)
	// 	// for g, c := range cnt {
	// 	// 	if c > thrsh {
	// 	// 		gs[g] = make([]int, 0, c)
	// 	// 	}
	// 	// }
	// 	// if len(gs) == 0 {
	// 	// 	break
	// 	// }

	// 	// for c, g := range ws {
	// 	// 	gs[g] = append(gs[g], c)
	// 	// }
	// }

	return newsws
}
