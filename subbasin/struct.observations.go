package subbasin

import (
	"encoding/gob"
	"fmt"
	"os"
	"slices"

	"github.com/maseology/mmaths/slice"
	"github.com/maseology/mmio"
)

// Observations holds locations where the model is to provide model outputs
type Observations struct {
	SWtem, GWtem map[int][]float64 // observation timeseries; [cellID]timeseries
	SWnam, GWnam map[int]string    // SW montior names, GW monitor names; [cellID]name
	SWmon        [][]int           // [sws index] to SWmid indices
	SWmid        []int             // SW monitor IDs (cell ids); GW monitor IDs (cell ids)
}

func (ob *Observations) SaveGob(fp string) error {
	f, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf(" Observations.SaveGob %v", err)
	}
	if err := gob.NewEncoder(f).Encode(ob); err != nil {
		return fmt.Errorf(" Observations.SaveGob %v", err)
	}
	f.Close()
	return nil
}

func LoadGobObservations(fp string) (*Observations, error) {
	var ob Observations
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	enc := gob.NewDecoder(f)
	err = enc.Decode(&ob)
	if err != nil {
		return nil, err
	}
	f.Close()
	return &ob, nil
}

func SubsetFromObservation(s *Structure, w *Subwatershed, nt int, fp string) (*Observations, error) {
	var err error
	var oin *Observations
	switch mmio.GetExtension(fp) {
	case ".gob":
		oin, err = LoadGobObservations(fp)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unrecognized observation file type: %s", fp)
	}

	mx := make(map[int]int, s.Nc)
	dmons := make(map[int]int)
	for i, c := range s.Cids {
		mx[c] = i
	}
	shyd, snam := make(map[int][]float64), make(map[int]string)
	for c, a := range oin.SWtem {
		if i, ok := mx[c]; ok {
			shyd[c] = make([]float64, nt)
			copy(shyd[c], a)
			snam[c] = oin.SWnam[c]
			dmons[i] = c
		}
	}

	swsmons, mids := BuildQmonTopology(s, w, snam, dmons)

	ghyd, gnam := make(map[int][]float64), make(map[int]string)
	for c, a := range oin.GWtem {
		ghyd[c] = make([]float64, nt)
		copy(ghyd[c], a)
		gnam[c] = oin.GWnam[c]
	}

	return &Observations{
		SWtem: shyd,    // surface water [cellID]timeseries
		GWtem: ghyd,    // groundwater [cellID]timeseries
		SWmon: swsmons, // [swsID]<monitor index> ; <0 no monitor
		SWmid: mids,    // list of monitor IDs []cellID
		SWnam: snam,    // SW monitor names; [cellID]name
		GWnam: gnam,    // GW monitor names; [cellID]name
	}, nil
}

func BuildQmonTopology(s *Structure, w *Subwatershed, nmons map[int]string, amons map[int]int) ([][]int, []int) {
	fmt.Printf("  build upslope topology..\n")

	fixObsCatchmentTopography := func(upcswska, upswsk []int) []int {
		nucsws := len(upcswska)
		nusws := len(upswsk)
		if nucsws > nusws {
			panic("getObservation topology error 1: more cell-based upslope SWSs than sws-based upslope SWSs")
		}
		if nucsws > 0 {
			for _, sc := range upcswska {
				if !slices.Contains(upswsk, sc) {
					panic("getObservation topology error 2: cell-based upslope SWS is not an sws-based upslope SWSs")
				}
			}
			return upcswska
		} else {
			return nil
		}
	}

	upsws := w.BuildUpSWS()
	us := func() map[int][]int {
		o := make(map[int][]int)
		for a, d := range s.Ds {
			if _, ok := o[d]; !ok {
				o[d] = []int{}
			}
			o[d] = append(o[d], a)
		}
		return o
	}()
	upcsws := func(k, a int) []int { // determines upslope sws by climbing cells
		coll := []int{}
		var climb func(int)
		climb = func(c int) {
			if _, ok := us[c]; !ok {
				return
			}
			for _, u := range us[c] {
				if w.Sid[u] == k {
					climb(u)
				} else {
					coll = append(coll, w.Sid[u])
				}
			}
		}
		climb(a)
		return slice.Distinct(coll)
	}

	cmon := make([]int, 0, len(nmons))
	for c := range nmons {
		cmon = append(cmon, c)
	}
	smns := make([][]int, len(w.Sais))
	coll := make([]int, 0, len(amons))
	nRemoved := 0
	for k, ais := range w.Sais {
		for _, a := range ais {
			if c, ok := amons[a]; ok {
				for i, ci := range cmon {
					if ci == c {
						// smns[k] = []int{i} // naive: assumes sub-watersheds are created by SW monitoring locations
						s := fixObsCatchmentTopography(upcsws(k, a), upsws[k])
						if s == nil {
							nRemoved++
						} else {
							for _, kk := range s {
								smns[kk] = append(smns[kk], i)
							}
						}
					}
				}
				coll = append(coll, a)
			}
		}
	}

	if nRemoved > 0 {
		fmt.Printf(" Monitoring runoff at %d locations (%d SW monitoring locations removed for small (<10km²) catchment areas)\n", len(coll)-nRemoved, nRemoved)
	} else {
		fmt.Printf(" Monitoring runoff at %d locations\n", len(coll))
	}
	return smns, cmon
}
