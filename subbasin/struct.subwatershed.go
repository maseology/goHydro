package subbasin

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/maseology/goHydro/grid"
)

type Subwatershed struct {
	Outer, Sais, Sads          [][]int
	Reach                      []Reach
	Sid, Dsws, Dcid, Isws, Sgw []int
	Fnsc                       []float64
	Islake                     []bool
	Ns                         int
}

func (w *Subwatershed) Checkandprint(gd *grid.Definition, cids []int, fnc float64, chkdirprfx string, crop bool) {

	var gd2 *grid.Definition
	xr := make(map[int]int)
	if crop {
		gd2, xr = gd.CropToActives()
	} else {
		gd2 = gd
		for _, c := range gd.Sactives {
			xr[c] = c
		}
	}

	// summarize
	fmt.Printf("   %d sub-watersheds in %d rounds, computionally ordered:\n", w.Ns, len(w.Outer))
	if len(w.Outer) > 10 {
		for k, inner := range w.Outer {
			if k < 3 || k == len(w.Outer)-1 {
				fmt.Printf("        round %d (%d)\n", k+1, len(inner))
			} else if k == 3 {
				print("         ...\n")
			}
		}
	} else {
		println("        ID          SWSID        n cells  (%%of domain)")
		for k, inner := range w.Outer {
			fmt.Printf("    round %d (%d)\n", k+1, len(inner))
			for _, isw := range inner {
				fmt.Printf("%10d%15d%15d  (%d %%)\n", isw, w.Isws[isw], int(w.Fnsc[isw]), int(100*w.Fnsc[isw]/fnc))
			}
		}
	}

	mx := make(map[int]int, len(cids))
	wcis := make(map[int][]int, w.Ns)
	for i, c := range cids {
		wcis[w.Sid[i]] = append(wcis[w.Sid[i]], i)
		mx[c] = i
	}

	si, sids, saids, dsws, dcid, sads, sgw, islak := gd2.NullInt32(-9999), gd2.NullInt32(-9999), gd2.NullInt32(-9999), gd2.NullInt32(-9999), gd2.NullInt32(-9999), gd2.NullInt32(-9999), gd2.NullInt32(-9999), gd2.NullInt32(-9999)
	rchl, rchs := gd2.NullArray(-9999.), gd2.NullArray(-9999.)
	hassgw := w.Sgw != nil
	for _, c := range gd.Sactives {
		if i, ok := mx[c]; ok {
			c2 := xr[c]
			si[c2] = int32(w.Sid[i])
			sids[c2] = int32(w.Isws[w.Sid[i]])
			dsws[c2] = int32(w.Dsws[w.Sid[i]])
			dcid[c2] = int32(w.Dcid[w.Sid[i]])
			rchl[c2] = w.Reach[w.Sid[i]].Length
			rchs[c2] = w.Reach[w.Sid[i]].Slope

			if w.Islake[w.Sid[i]] {
				islak[c2] = 1
			}
			if hassgw {
				sgw[c2] = int32(w.Sgw[w.Sid[i]])
			}
		}
	}
	for k, sa := range w.Sais {
		for i, a := range sa {
			c := xr[cids[a]]
			saids[c] = int32(a)
			sads[c] = int32(w.Sads[k][i])
		}
	}

	sord := gd2.NullInt32(-9999)
	for k, inner := range w.Outer {
		for _, isw := range inner {
			for _, i := range wcis[isw] {
				sord[xr[cids[i]]] = int32(k + 1)
			}
		}
	}

	writeInts(gd2, chkdirprfx+"sws.aid.bil", si)     // zero-based index
	writeInts(gd2, chkdirprfx+"sws.sid.bil", sids)   // original index
	writeInts(gd2, chkdirprfx+"sws.said.bil", saids) // zero-based index, per sws
	writeInts(gd2, chkdirprfx+"sws.sads.bil", sads)  // cell topology per sub-watershed, <0 is routed to down-SWS
	if hassgw {
		writeInts(gd2, chkdirprfx+"sws.sgw.bil", sgw) // groundwater index, now projected to sws
	}
	writeInts(gd2, chkdirprfx+"sws.dsws.bil", dsws)    // downslope sws index
	writeInts(gd2, chkdirprfx+"sws.dcid.bil", dcid)    // downslope cell id from sws outlet
	writeInts(gd2, chkdirprfx+"sws.order.bil", sord)   // computational sws ordering
	writeInts(gd2, chkdirprfx+"sws.islake.bil", islak) // shows which sws is deemed a lake

	writeFloats32(gd2, chkdirprfx+"sws.rchlength.bil", rchl) // sub-watershed reach length
	writeFloats32(gd2, chkdirprfx+"sws.rchslope.bil", rchs)  // slope of sub-watershed reach
}

func (w *Subwatershed) SaveGob(fp string) error {
	f, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf(" mapper.SaveGob %v", err)
	}
	if err := gob.NewEncoder(f).Encode(w); err != nil {
		return fmt.Errorf(" mapper.SaveGob %v", err)
	}
	f.Close()
	return nil
}

func LoadGobSubwatershed(fp string) (*Subwatershed, error) {
	var rtr Subwatershed
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	enc := gob.NewDecoder(f)
	err = enc.Decode(&rtr)
	if err != nil {
		return nil, err
	}
	f.Close()
	return &rtr, nil
}

func (w *Subwatershed) BuildUpSWS() map[int][]int {
	o := make(map[int][]int, len(w.Sid))
	for i := range w.Sais {
		if _, ok := o[w.Dsws[i]]; !ok {
			o[w.Dsws[i]] = []int{}
		}
		o[w.Dsws[i]] = append(o[w.Dsws[i]], i)
	}
	return o
}
