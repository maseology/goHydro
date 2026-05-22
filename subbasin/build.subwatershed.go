package subbasin

import (
	"fmt"
	"log"
	"sort"

	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmaths/slice"
)

func (s *Structure) loadSWS(swsfp string) Subwatershed {

	// getting sws IDs mapped to topo-safe array (note asids is altered below once sws IDs are remapped to a 0-based array)
	asids := func(fp string) []int {
		fmt.Printf("   loading: %s\n", fp)
		var g grid.Indx
		g.GD = s.GD
		g.New(fp) //, true)
		aout := make([]int, s.Nc)
		crm := []int{}
		for i, c := range s.Cids { // topo-safe cell order
			if v, ok := g.A[c]; ok {
				if v < 0 {
					crm = append(crm, c)
					continue
				}
				aout[i] = v
			} else {
				panic("loadSWS loadIndx error: " + fp)
			}
		}
		if len(crm) > 0 {
			log.Fatalf("\n    ERROR: %d cells (%.3f%%) were not assigned a positive SWS ID. (Likely due to the trimming of small SWSs.)\n      Re-assign GDEF to the SWS layer to avoid this message.\n      cids: %v", len(crm), float64(len(crm))*100./float64(s.Nc), crm)
		}
		return aout
	}(swsfp)

	// mapping sorted sws IDs to a 0-based array index
	xsws, isws := func() (map[int]int, []int) {
		d := make(map[int]int)
		for i := range s.Cids {
			d[asids[i]]++
		}
		u := make([]int, 0, len(d))
		for k := range d {
			u = append(u, k)
		}
		sort.Ints(u)
		for i, uu := range u {
			if _, ok := d[uu]; !ok {
				panic("loadSWS xsws error")
			}
			d[uu] = i
		}
		return d, u
	}()

	// checking for consistency, remapping asids to the 0-based sws id, and gathering sws sizes
	fnsc := make([]float64, len(xsws))
	for i := range s.Cids {
		if isw, ok := xsws[asids[i]]; ok {
			asids[i] = isw // reset mapped sws IDs to a 0-based array index
			fnsc[isw]++
		} else {
			panic("loadSWS isws error")
		}
	}

	// collecting lists of acids per aswsids
	mcids := make(map[int][]int, len(xsws))
	for i := range s.Cids { // topo-safe cell order
		mcids[asids[i]] = append(mcids[asids[i]], i) // topo-safe cell order on a per-sws basis
	}

	newds := func(scids []int) []int {
		m := make(map[int]int, len(scids))
		dsc := make([]int, len(scids))
		for i, c := range scids {
			dsc[i] = s.Ds[c]
			m[c] = i
		}

		ds := make([]int, len(dsc))
		for i, k := range dsc {
			if ids, ok := m[k]; ok {
				ds[i] = ids
			} else {
				ds[i] = -1 // not draining to cell within current sws, setting to <0 will force rdrr to drain to downslope SWS
			}
		}
		return ds
	}

	// remapping mcids to a list of lists, building downslopes on a per-sws basis
	saids := make([][]int, len(mcids))
	sads := make([][]int, len(mcids))
	for k, v := range mcids {
		sads[k] = newds(v)
		saids[k] = v
	}

	// collecting sws topologies
	dsws, dcid := func() ([]int, []int) {
		dsws, dcid := make([]int, len(saids)), make([]int, len(saids))
		for is, c := range saids {
			oi := len(c) - 1 // final cell id drains to downslope SWS
			if sads[is][oi] != -1 {
				panic("loadSWS SWStopo err")
			}
			di := s.Ds[c[oi]]
			if di > -1 {
				ds := asids[di]
				dc := func() int {
					dsc := saids[ds]
					for j := len(dsc) - 1; j >= 0; j-- {
						if dsc[j] == di {
							return j
						}
					}
					return -1
				}()
				dsws[is] = ds
				dcid[is] = dc
			} else {
				dsws[is] = -1 // model outlet
				dcid[is] = -1
			}
		}
		return dsws, dcid
	}()

	return Subwatershed{
		Sais:   saids, // set of cell indices per sws
		Sads:   sads,  // cell topology per sub-watershed, <0 is routed to down-SWS
		Dsws:   dsws,  // downslope sub-watershed, -1 out of model
		Dcid:   dcid,  // cell ID receiving runoff, -1 out of model
		Sid:    asids, // 0-based cell index to 0-based sws index
		Isws:   isws,  // sws index to sub-watersed ID (needed for forcings)
		Fnsc:   fnsc,  // number of cells per sws
		Islake: make([]bool, len(xsws)),
		Ns:     len(xsws), // number of sws's
	}
}

func (w *Subwatershed) UpdateDS(s *Structure) {
	newds := func(scids []int) []int {
		m := make(map[int]int, len(scids))
		dsc := make([]int, len(scids))
		for i, c := range scids {
			dsc[i] = s.Ds[c]
			m[c] = i
		}

		ds := make([]int, len(dsc))
		for i, k := range dsc {
			if ids, ok := m[k]; ok {
				ds[i] = ids
			} else {
				ds[i] = -1
			}
		}
		return ds
	}

	mcids := make(map[int][]int, w.Ns)
	for i := range s.Cids { // topo-safe cell order
		mcids[w.Sid[i]] = append(mcids[w.Sid[i]], i)
	}
	saids := make([][]int, len(mcids))
	sads := make([][]int, len(mcids))
	for k, v := range mcids {
		sads[k] = newds(v)
		saids[k] = v
	}
	w.Sads = sads
	w.Sais = saids
}

func (w *Subwatershed) BuildReaches(s *Structure, mp *Mapper, flength float64) {
	// collect routing properties
	usws := w.BuildUpSWS()
	rch := make([]Reach, len(w.Isws))
	// cbs, ces := make([]int, len(w.Isws)), make([]int, len(w.Isws))

	drain := func(startcell, endcell int) (float64, float64) {
		if startcell == endcell {
			return flength, s.Dwnslope[startcell]
		}

		cendcell := s.Cids[endcell]
		next, n := s.Ds[startcell], 1.
		rchslp := s.Dwnslope[startcell]
		for {
			rchslp += s.Dwnslope[next]
			if w.Isws[w.Sid[next]] != cendcell {
				break
			}
			if next == endcell {
				break
			}
			next = s.Ds[next]
			if next < 0 {
				break
			}
			n++
		}
		return flength * n, rchslp / n
	}

	for k, ais := range w.Sais {
		startcells := []int{}
		if us, ok := usws[k]; ok {
			for _, uus := range us {
				startcells = append(startcells, ais[w.Dcid[uus]])
			}
		}

		startcell, upn := -1, -1
		for _, sa := range startcells {
			if s.Upcnt[sa] > upn {
				startcell = sa
				upn = s.Upcnt[sa]
			}
		}

		endcell := ais[len(ais)-1]
		// ces[k] = endcell
		if startcell < 0 {
			// cbs[k] = -1
			mslp := func() float64 {
				// sslp := 0.
				// for _, a := range w.Sais[k] {
				// 	sslp += s.Dwnslope[a]
				// }
				// return sslp / float64(len(w.Sais[k]))
				coll := make([]float64, len(w.Sais[k]))
				for i, a := range w.Sais[k] {
					coll[i] = s.Dwnslope[a]
				}
				return slice.Median(coll)
			}()
			rch[k] = Reach{flength, mslp} // headwater reach
		} else {
			// cbs[k] = startcell
			l, s := drain(startcell, endcell)
			rch[k] = Reach{l, s}
		}
	}

	// cb, ce := s.GD.NullInt32(-9999), s.GD.NullInt32(-9999)
	// rchl, rchs := s.GD.NullArray(-9999.), s.GD.NullArray(-9999.)
	// for _, c := range s.GD.Sactives {
	// 	if i, ok := mp.Mx[c]; ok {
	// 		if cbs[w.Sid[i]] < 0 {
	// 			cb[c] = -1
	// 		} else {
	// 			cb[c] = int32(s.Cids[cbs[w.Sid[i]]])
	// 		}
	// 		ce[c] = int32(s.Cids[ces[w.Sid[i]]])
	// 		rchl[c] = rch[w.Sid[i]].Length
	// 		rchs[c] = rch[w.Sid[i]].Slope
	// 	}
	// }
	// writeInts(s.GD, "M:/OWRC23-RDRR/check/sws.startcell.bil", cb)
	// writeInts(s.GD, "M:/OWRC23-RDRR/check/sws.endcell.bil", ce)
	// writeFloats32(s.GD, "M:/OWRC23-RDRR/check/sws.rchlength.bil", rchl) // sub-watershed reach length
	// writeFloats32(s.GD, "M:/OWRC23-RDRR/check/sws.rchslope.bil", rchs)  // slope of sub-watershed reach

	w.Reach = rch
}
