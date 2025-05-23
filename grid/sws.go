package grid

import (
	"fmt"
	"log"

	"github.com/maseology/mmio"
)

type SWS struct {
	GD         *Definition
	SwsC, Usws map[int][]int // SwsC: [swsID]cellID; Usws: [downstream_swsID]upstream_swsID
	CSws, Dsws map[int]int   // CSws: [cellID]swsID; Dsws: [upstream_swsID]downstream_swsID
}

func CollectSWS(swsFP string, gd *Definition) *SWS {

	var gsws Indx
	if gd == nil {
		gsws.ImportBil(swsFP)
		gd = gsws.GD
	} else {
		gsws.GD = gd
		gsws.New(swsFP) //, false)
	}

	cs := gsws.A
	sc := make(map[int][]int, len(gsws.UniqueValues()))
	for c, s := range cs {
		sc[s] = append(sc[s], c)
		// if _, ok := sc[s]; ok {
		// 	sc[s] = append(sc[s], c)
		// } else {
		// 	sc[s] = []int{c}
		// }
	}

	// collect topology
	var dsws map[int]int
	var usws map[int][]int
	topoFP := mmio.RemoveExtension(swsFP) + "-topo.csv"
	nsws := len(sc)
	if _, ok := mmio.FileExists(topoFP); ok {
		d, err := mmio.ReadCSV(topoFP, 1)
		if err != nil {
			log.Fatalf(" grid.CollectSWS: error reading %s: %v\n", topoFP, err)
		}
		// OLD
		// // dsws = make(map[int]int, len(d)) // note: swsids not contained within dsws drain to farfield
		// // for _, ln := range d {
		// // 	dsws[int(ln[1])] = int(ln[2]) // linkID,upstream_swsID,downstream_swsID
		// // }
		// dsws = make(map[int]int, nsws) // note: swsids not contained within dsws drain to farfield
		// usws = make(map[int][]int, nsws)
		// for _, ln := range d {
		// 	if _, ok := sc[int(ln[1])]; ok {
		// 		if _, ok := sc[int(ln[2])]; ok {
		// 			dsws[int(ln[1])] = int(ln[2]) // linkID,upstream_swsID,downstream_swsID
		// 			usws[int(ln[2])] = append(usws[int(ln[2])], int(ln[1]))
		// 		}
		// 	}
		// }
		dsws = make(map[int]int, nsws) // note: swsids not contained within dsws drain to farfield
		usws = make(map[int][]int, nsws)
		for _, ln := range d {
			if _, ok := sc[int(ln[0])]; ok {
				if _, ok := sc[int(ln[1])]; ok {
					dsws[int(ln[0])] = int(ln[1]) // sid,dsid,dcid
					usws[int(ln[1])] = append(usws[int(ln[1])], int(ln[0]))
				}
			}
		}
	} else {
		// fmt.Printf(" warning: sws topology (*.topo) not found\n")
		log.Fatalf(" grid.CollectSWS error: sws topology (*.topo) not found: %s", topoFP)
	}

	return &SWS{
		GD:   gd,
		CSws: cs,
		SwsC: sc,
		Dsws: dsws,
		Usws: usws,
	}
}

func (s *SWS) ClimbFrom(swsID int) []int {
	var i []int
	var recurs func(int)
	recurs = func(sid int) {
		i = append(i, sid)
		if a, ok := s.Usws[sid]; ok {
			for _, ii := range a {
				recurs(ii)
			}
		}
	}
	recurs(swsID)
	return i
}

func (s *SWS) Print() {
	fmt.Printf("subwatersheds loaded\n nCells %d\n nSws %d\n nDwnSws %d\n nUpSws %d\n\n", len(s.CSws), len(s.SwsC), len(s.Dsws), len(s.Usws))
}
