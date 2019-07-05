package swat

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/maseology/mmio"
)

// temporary data loaders
type lbsn struct{ SUBKM, SLSUBBSN, CHL, CHS, CHN, SURLAG, GWDELAY, ALPHABF float64 }
type lrte struct{ CHL, CHS, CHW, CHD, CHN float64 }
type lhru struct {
	SWSID                                                        int
	HRUFR, HRUSLP, OVN, CN2, CV, ESCO, CLAY, SOLBD, SOLAWC, SOLK float64
}

// Load a set of .csv files to build a SWAT model structure
func Load(fbsn, fhru, frte, ftopo string) (WaterShed, []int) {

	ibsn, err := mmio.ReadCSV(fbsn) // SWSID,SUBKM,SLSUBBSN,CHL,CHS,CHN,SURLAG,GWDELAY,ALPHABF
	if err != nil {
		log.Fatalf("main: Error reading %s: %v\n", fbsn, err)
	}
	sbsn := make(map[int]lbsn, len(ibsn))
	for _, s := range ibsn {
		sb := lbsn{
			SUBKM:    s[1],
			SLSUBBSN: s[2],
			CHL:      s[3],
			CHS:      s[4],
			CHN:      s[5],
			SURLAG:   s[6],
			GWDELAY:  s[7],
			ALPHABF:  s[8],
		}
		sbsn[int(s[0])] = sb
	}
	fmt.Printf("%d sub-basins read\n", len(ibsn))

	irte, err := mmio.ReadCSV(frte) // SWSID,CHL,CHS,CHW,CHD,CHN
	if err != nil {
		log.Fatalf("main: Error reading %s: %v\n", frte, err)
	}
	srte := make(map[int]lrte, len(irte))
	for _, s := range irte {
		sr := lrte{
			CHL: s[1],
			CHS: s[2],
			CHW: s[3],
			CHD: s[4],
			CHN: s[5],
		}
		srte[int(s[0])] = sr
	}
	fmt.Printf("%d channels read\n", len(irte))

	ihru, err := mmio.ReadCSV(fhru) // SWSID,HRUFR,HRUSLP,OVN,CN2,CV,CLAY,SOLBD,SOLAWC,SOLK
	if err != nil {
		log.Fatalf("main: Error reading %s: %v\n", fhru, err)
	}
	shru := make(map[int]lhru, len(ihru))
	for i, s := range ihru {
		sh := lhru{
			SWSID:  int(s[0]),
			HRUFR:  s[1],
			HRUSLP: s[2],
			OVN:    s[3],
			CN2:    s[4],
			CV:     s[5],
			ESCO:   s[6],
			CLAY:   s[7],
			SOLBD:  s[8],
			SOLAWC: s[9],
			SOLK:   s[10],
		}
		shru[i] = sh
	}
	fmt.Printf("%d HRUs read\n", len(ihru))

	itopo, err := mmio.ReadTextLines(ftopo) // sub-basin topology
	if err != nil {
		log.Fatalf("main: Error (1) reading %s: %v\n", ftopo, err)
	}
	topo = make(map[int][]int, len(itopo)) // SubBasin topology {to:[]from}
	for _, s := range itopo {
		s1 := strings.Split(s, ",")
		s2 := strings.Split(s1[1], " ")
		icoll := make([]int, 0, len(s2))
		i1, err := strconv.Atoi(strings.TrimSpace(s1[0]))
		if err != nil {
			log.Fatalf("main: Error (2) reading %s: %v\n", ftopo, err)
		}
		for _, s := range s2 {
			if len(s) == 0 {
				continue
			}
			i2, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				log.Fatalf("main: Error (3) reading %s: %v\n", ftopo, err)
			}
			icoll = append(icoll, i2)
		}
		topo[i1] = icoll
	}
	fmt.Printf("sub-basin topology read\n")

	// create model
	hrus := make(map[int][]*HRU, len(sbsn))
	for _, s := range shru {
		var sl SoilLayer
		sl.New(s.CLAY, s.SOLBD, s.SOLAWC, s.SOLK)
		var hru HRU
		hru.New(sl, s.HRUFR, s.HRUSLP, s.OVN, s.CN2, s.CV, s.ESCO, false)
		if _, ok := hrus[s.SWSID]; !ok {
			hrus[s.SWSID] = make([]*HRU, 0)
		}
		hrus[s.SWSID] = append(hrus[s.SWSID], &hru)
	}
	chns := make(map[int]*Channel, len(sbsn))
	for i, s := range srte {
		var chn Channel
		chn.New(s.CHW, s.CHD, s.CHL, s.CHS, s.CHN)
		chns[i] = &chn
	}
	sbsns := make(map[int]*SubBasin, len(sbsn))
	for i, s := range sbsn {
		var bsn SubBasin
		bsn.New(hrus[i], chns[i], s.SUBKM, s.SLSUBBSN, s.CHL, s.CHS, s.CHN, s.SURLAG, s.GWDELAY, s.ALPHABF)
		sbsns[i] = &bsn
	}

	// set subbasin topology
	for to, froms := range topo {
		for _, from := range froms {
			if _, ok := sbsns[from]; ok {
				sbsns[from].Outflow = to
			}
		}
	}

	return sbsns, topoOrder()
}
