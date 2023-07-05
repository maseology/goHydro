package tem

import (
	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmaths"
)

const small = 1e-11 // "have ~15 significant digits"; using: 1e-11 decimals allow 4-digit elvation <9999 in 64-bit floats

func (t *TEM) FillDepressions(gd *grid.Definition) {
	// ref: Wang, L., H. Liu, 2006. An efficient method for identifying and filling surface depressions in digital elevation models for hydrologic analysis and modelling. International Journal of Geographical Information Science 20(2): 193-213.
	// NOTE: Zhou etal. (2016) is supposed to be faster than Wang and Liu (2006) but doesn't appear to be the case as coded here.
	println("  building priority queue")
	pq := mmaths.NewPriorityQueue()
	zs := make(map[int]float64)
	bufs := gd.Buffers(false, true)
	for _, c := range t.Outlets() {
		if _, ok := bufs[c]; !ok {
			panic("FillDepressions err2.1")
		}
		if func() bool { // edge detection
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
			} else {
				panic("FillDepressions err2.2")
			}
		}
	}

	println("  running priority queue, filling depressions")
	flat := make(map[int]float64)
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
				if tc, ok := t.TEC[bc]; ok {
					if tc.Z > zs[c] {
						zs[bc] = tc.Z
					} else {
						zs[bc] = zs[c]   // + small // (initial/simple) flat areas solution
						flat[bc] = zs[c] // collecting original elevations
						flat[c] = zs[c]  // needed for flat region code
					}
					pq.Push(bc, zs[bc])
				} else {
					panic("FillDepressions err4")
				}
			}
		}
	}

	if true {
		zs = fixflatregions(gd, zs, flat, bufs) // I don't see much improvement
	}

	println("  re-building flowpaths")
	for c, z := range zs {
		if tt, ok := t.TEC[c]; ok {
			tt.Z = z
			t.TEC[c] = tt
		} else {
			panic("FillDepressions.buildflowpaths err5")
		}
	}
	t.BuildUpslopes(t.buildDsFromNeighbours(bufs))
}

func fixflatregions(gd *grid.Definition, zs, flat map[int]float64, bufs map[int][]int) map[int]float64 {
	// after: Garbrecht Martz 1997 The assignment of drainage direction over flat surfaces in raster digital elevation models

	// odir := "E:/Sync/@dev/pages_owrc/interpolants/interpolation/calc/hydroDEM/test/"

	println("  fixing flat regions..")
	crwl := gd.ToCrawler()
	println("    locating flat regions")
	iflat, mbrd, ng := crwl.CrawlByFloat(flat, false)
	// writeInts(odir+"iflat.indx", iflat, gd.Ncells())

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
	removeinternalboundaries := func(bndry map[int][]int) map[int][]int {
		println("    removing internal boundaries")
		o := make(map[int][]int, len(bndry))
		for g, a := range bndry {
			m := make(map[int]float64, len(a))
			for _, aa := range a {
				m[aa] = 0.
			}
			ibn, _, nb := crwl.CrawlByFloat(m, true)
			if nb == 1 {
				o[g] = a
			} else {
				ax, ix := 0, -1
				vbn := minv(ibn, nb)
				for gg, aa := range vbn { // assuming the larges boundary is the outside boundary
					if len(aa) > ax {
						ax = len(aa)
						ix = gg
					}
				}
				o[g] = append(o[g], vbn[ix]...)
			}
		}
		return o
	}
	mbrd = removeinternalboundaries(mbrd)

	ibrd := make(map[int]int)
	for g, a := range mbrd {
		for _, aa := range a {
			ibrd[aa] = g
		}
	}
	// writeInts(odir+"ibrd.indx", ibrd, gd.Ncells())

	println("    applying fixes..")
	olev1, olev2, olev3 := make(map[int]int, len(flat)), make(map[int]int, len(flat)), make(map[int]int, len(flat))
	for g, aflat := range minv(iflat, ng) {
		// fmt.Print(g, aflat[0])
		if len(aflat) == 1 {
			bmin := 1e10
			for _, bc := range bufs[aflat[0]] {
				if bmin > zs[bc] {
					bmin = zs[bc]
				}
			}
			zs[aflat[0]] = bmin + small
			// println(" - single cell")
		} else if b, ok := mbrd[g]; ok {
			amin, ma, aouts, bouts := 1e10, make(map[int]int, len(aflat)), make(map[int]int), make(map[int]int)
			for _, c := range aflat {
				ma[c] = 1
				if amin > zs[c] {
					amin = zs[c]
				}
			}
			for _, c := range b {
				if amin >= zs[c] {
					bouts[c]++ // finding flat reigion outlets
				}
			}
			if len(bouts) == 0 {
				func() {
					for _, c := range aflat {
						for _, bc := range bufs[c] {
							if bc < 0 {
								aouts[c]++ // flat region draining to farfield
								return
							}
						}
					}
					panic("filldrepressions flat regaion farfield expected")
				}()
			} else {
				for bout := range bouts {
					for _, bc := range bufs[bout] {
						if _, ok := ma[bc]; ok {
							aouts[bc]++
						}
					}
				}
			}

			// Step 1: gradient towards lower terrain
			func() {
				// print(" - step1")
				q := make([]int, 0, len(aouts))
				alev := make(map[int]int)
				for a := range aouts {
					alev[a] = 0
					q = append(q, a)
				}
				for len(q) > 0 {
					a := q[0]
					q = q[1:]
					if _, ok := alev[a]; !ok {
						panic("fixflatregions step 1 err")
					}
					for _, bc := range bufs[a] {
						if _, ok := ma[bc]; ok {
							if v, ok := alev[bc]; ok {
								if v > alev[a]+1 {
									q = append(q, bc)
									alev[bc] = alev[a] + 1
								}
							} else {
								q = append(q, bc)
								alev[bc] = alev[a] + 1
							}
						}
					}
				}
				for c, v := range alev {
					olev1[c] = v
				}
			}()

			// Step 2: gradient away from higher terrain
			func() {
				// print(" - step2")
				blev, aq := make(map[int]int), []int{}
				for _, bb := range b {
					for _, bc := range bufs[bb] {
						if _, ok := ma[bc]; ok {
							blev[bc] = 0
							aq = append(aq, bc)
						}
					}
				}
				for len(aq) > 0 {
					b := aq[0]
					aq = aq[1:]
					if _, ok := blev[b]; !ok {
						panic("fixflatregions step 2 err")
					}
					for _, bc := range bufs[b] {
						if _, ok := ma[bc]; ok {
							if v, ok := blev[bc]; ok {
								if v > blev[b]+1 {
									aq = append(aq, bc)
									blev[bc] = blev[b] + 1
								}
							} else {
								aq = append(aq, bc)
								blev[bc] = blev[b] + 1
							}
						}
					}
				}
				vx := 0
				for _, v := range blev {
					if v > vx {
						vx = v
					}
				}
				vx++
				for c, v := range blev {
					if olev1[c] > 0 {
						olev2[c] = vx - v
					} else {
						olev2[c] = 0 // flat area outlet cells are excluded
					}
				}
			}()
			// println(" - complete")
		} else {
			panic("fixflatregions err2")
		}
	}

	// Step 3: combined gradient and final drainage pattern
	println("    correcting elevations")
	for c := range flat {
		olev3[c] = 3*olev1[c] + 2*olev2[c]
	}
	// writeInts(odir+"olev1.indx", olev1, gd.Ncells())
	// writeInts(odir+"olev2.indx", olev2, gd.Ncells())
	// writeInts(odir+"olev3.indx", olev3, gd.Ncells())
	for c, v := range olev3 {
		zs[c] += float64(v) * small
	}
	for c := range iflat {
		func() { // last fix for "rather exceptional situations" (pg.208)
			for _, bc := range bufs[c] {
				if zs[c] > zs[bc] {
					return
				}
			}
			zs[c] += small
		}()
	}

	return zs
}

// func writeInts(fp string, m map[int]int, nc int) {
// 	i32 := make([]int32, nc)
// 	for c := 0; c < nc; c++ {
// 		if v, ok := m[c]; ok {
// 			i32[c] = int32(v)
// 		} else {
// 			i32[c] = -9999
// 		}
// 	}
// 	buf := new(bytes.Buffer)
// 	if err := binary.Write(buf, binary.LittleEndian, i32); err != nil {
// 		panic(err)
// 	}
// 	if err := ioutil.WriteFile(fp, buf.Bytes(), 0644); err != nil { // see: https://en.wikipedia.org/wiki/File_system_permissions
// 		panic(err)
// 	}
// }

// func writeFloats(fp string, m map[int]float64, nc int) {
// 	f32 := make([]float32, nc)
// 	for c := 0; c < nc; c++ {
// 		if v, ok := m[c]; ok {
// 			f32[c] = float32(v)
// 		} else {
// 			f32[c] = -9999.
// 		}
// 	}
// 	buf := new(bytes.Buffer)
// 	if err := binary.Write(buf, binary.LittleEndian, f32); err != nil {
// 		panic(err)
// 	}
// 	if err := ioutil.WriteFile(fp, buf.Bytes(), 0644); err != nil { // see: https://en.wikipedia.org/wiki/File_system_permissions
// 		panic(err)
// 	}
// }
