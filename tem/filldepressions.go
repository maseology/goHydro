package tem

import (
	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmaths"
)

func (t *TEM) FillDepressions(GD *grid.Definition) map[int]int {
	// ref: Wang, L., H. Liu, 2006. An efficient method for identifying and filling surface depressions in digital elevation models for hydrologic analysis and modelling. International Journal of Geographical Information Science 20(2): 193-213.
	// NOTE: Zhou etal. (2016) is supposed to be faster than Wang and Liu (2006) but doesn't appear to be the case as coded here.
	println("  building priority queue")
	pq := mmaths.NewPriorityQueue()
	zs := make(map[int]float64)
	bufs := GD.Buffers(false, true)
	o, oi := make(map[int]int), 0
	for _, c := range t.Outlets() {
		if _, ok := bufs[c]; !ok {
			panic("FillDepressions err2.1")
		}
		if func() bool {
			for _, c := range bufs[c] {
				if c < 0 {
					return true
				}
			}
			return false
		}() {
			if t, ok := t.TEC[c]; ok {
				zs[c] = t.Z
				pq.Push(c, zs[c])
				o[c] = oi
				oi++
			} else {
				panic("FillDepressions err2.2")
			}
		}

	}

	println("  running priority queue")
	for pq.Len() > 0 {
		ic, err := pq.Pop()
		if err != nil {
			panic(err)
		}
		c := ic.(int)

		if _, ok := bufs[c]; !ok {
			panic("FillDepressions err3")
		}
		for _, bc := range bufs[c] {
			if bc < 0 || bc == c {
				continue
			}
			if _, ok := zs[bc]; !ok {
				if t, ok := t.TEC[bc]; ok {
					if t.Z > zs[c] {
						zs[bc] = t.Z
					} else {
						zs[bc] = zs[c] + .00001
					}
					pq.Push(bc, zs[bc])
					o[bc] = oi
					oi++
				} else {
					panic("FillDepressions err4")
				}
			}
		}
	}

	println("  re-building flowpaths")
	for c, z := range zs {
		if tt, ok := t.TEC[c]; ok {
			tt.Z = z
			t.TEC[c] = tt
		} else {
			panic("FillDepressions err5")
		}
	}

	t.BuildUpslopes(t.buildDs(bufs))
	return o
}

// zmin := 1e9
// for _, c := range GD.Sactives {
// 	if t, ok := t.TEC[c]; ok {
// 		if t.Z < zmin {
// 			zmin = t.Z
// 		}
// 	} else {
// 		panic("FillDepressions err1")
// 	}
// }
// zmin -= 1.

// ss := make(map[int]float64)
// for _, s := range t.Outlets() {
// 	if t, ok := t.TEC[s]; ok {
// 		ss[s] = t.Z - zmin
// 		pq.Push(s, (t.Z - zmin))
// 	} else {
// 		panic("FillDepressions err2")
// 	}
// }

// println("  running priority queue")
// bufs := GD.Buffers(false, true)
// for pq.Len() > 0 {
// 	ic, err := pq.Pop()
// 	if err != nil {
// 		panic(err)
// 	}
// 	c := ic.(int)
// 	fmt.Println(ss[c])

// 	if _, ok := bufs[c]; !ok {
// 		panic("FillDepressions err3")
// 	}
// 	for _, bc := range bufs[c] {
// 		if bc < 0 {
// 			continue
// 		}
// 		if _, ok := ss[bc]; !ok {
// 			if t, ok := t.TEC[bc]; ok {
// 				if t.Z-zmin > ss[c] {
// 					ss[bc] = t.Z - zmin
// 				} else {
// 					ss[bc] = ss[c]
// 				}
// 				pq.Push(bc, 1./ss[bc])
// 			} else {
// 				panic("FillDepressions err4")
// 			}
// 		}
// 	}
// }

// println("  re-building flowpaths")
// for i, s := range ss {
// 	if tt, ok := t.TEC[i]; ok {
// 		tt.Z = s + zmin
// 		t.TEC[i] = tt
// 	} else {
// 		panic("FillDepressions err5")
// 	}
// }

// t.BuildUpslopes(t.buildDs(bufs))
// }
