package grid

type Crawler struct {
	adj map[int][]int
}

// var bufrc = [][]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}}
// var bufrc = [][]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}}

func (gd *Definition) ToCrawler(cardinalOnly bool) *Crawler {
	a := make(map[int][]int, gd.Nact)
	var bufrc [][]int
	if cardinalOnly {
		bufrc = [][]int{{0, -1}, {-1, 0}, {1, 0}, {0, 1}}
	} else {
		bufrc = [][]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}}
	}
	for _, cid := range gd.Sactives {
		aa := make([]int, 0, 8)
		r, c := gd.RowCol(cid)
		for _, drc := range bufrc {
			sc := gd.CellID(r+drc[0], c+drc[1])
			if gd.IsActive(sc) {
				aa = append(aa, sc)
			}
		}
		a[cid] = aa
	}
	return &Crawler{a}
}

func (crwl *Crawler) CrawlByInt(m map[int]int, byKey bool) (groupId map[int]int, borderId map[int][]int, ng int) {
	q := make([]int, 0, len(m))
	groupId = make(map[int]int, len(m)) // assigned group id
	borderId = make(map[int][]int)
	for c := range m {
		groupId[c] = -9999
		q = append(q, c) // push
	}

	var recurs func(int)
	igrp := 0
	recurs = func(c int) {
		groupId[c] = igrp
		for _, ac := range crwl.adj[c] {
			if v, ok := m[ac]; ok {
				if byKey {
					if groupId[ac] < 0 {
						recurs(ac)
					}
				} else {
					if m[c] == v {
						if groupId[ac] < 0 {
							recurs(ac)
						}
					} else {
						borderId[igrp] = append(borderId[igrp], ac)
					}
				}
			} else {
				// borderId[ac] = igrp
				borderId[igrp] = append(borderId[igrp], ac)
			}
		}
	}
	for len(q) > 0 {
		c := q[0] // pop
		q = q[1:]

		if groupId[c] < 0 {
			borderId[igrp] = []int{}
			recurs(c)
			igrp++
		}
	}
	ng = igrp + 1
	return
}

func (crwl *Crawler) CrawlByFloat(m map[int]float64, byKey bool) (groupId map[int]int, borderId map[int][]int, ng int) {
	q := make([]int, 0, len(m))
	groupId = make(map[int]int, len(m)) // assigned group id
	borderId = make(map[int][]int)      // set of cells making up the border to every group (may be overlapping)
	for c := range m {
		groupId[c] = -9999
		q = append(q, c) // push
	}

	var recurs func(int)
	igrp := 0
	recurs = func(c int) {
		groupId[c] = igrp
		for _, ac := range crwl.adj[c] {
			if v, ok := m[ac]; ok {
				if byKey {
					if groupId[ac] < 0 {
						recurs(ac)
					}
				} else {
					if m[c] == v {
						if groupId[ac] < 0 {
							recurs(ac)
						}
					} else {
						borderId[igrp] = append(borderId[igrp], ac)
					}
				}
			} else {
				borderId[igrp] = append(borderId[igrp], ac)
			}
		}
	}
	for len(q) > 0 {
		c := q[0] // pop
		q = q[1:]

		if groupId[c] < 0 {
			borderId[igrp] = []int{}
			recurs(c)
			igrp++
		}
	}
	ng = igrp
	return
}
