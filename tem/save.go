package tem

import (
	"log"

	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmio"
)

func (t *TEM) SaveUHDEM(fp string, gd *grid.Definition) error {
	nc := len(t.TEC)
	uc := make([]uhdemReader, nc)
	mc := make(map[int]bool, nc)
	ii := 0
	for i, t := range t.TEC {
		cntrd := gd.CellCentroid(i)
		mc[i] = true
		uc[ii] = uhdemReader{
			I: int32(i),
			A: t.A,
			G: t.G,
			Z: t.Z,
			// X: -9999.,
			// Y: -9999.,
			X: cntrd[0],
			Y: cntrd[1],
		}
		ii++
	}

	type ft struct{ f, t int }
	m := make([]ft, 0, nc)
	for i, a := range t.USlp {
		for _, c := range a {
			if _, ok := mc[c]; !ok {
				log.Fatalf("SaveUHDEM error, TEC ID [%d] (upslope of %d) not found in model\n", c, i)
			}
			mc[c] = false
			m = append(m, ft{c, i})
		}
	}
	for c, b := range mc {
		if b {
			m = append(m, ft{c, -1}) // farfield
		}
	}
	fc, ii := make([]fpReader, len(m)), 0
	for _, f := range m {
		fc[ii] = fpReader{
			I:   int32(f.f),
			Ids: int32(f.t),
			Nds: int32(1),
			F:   1.,
		}
		ii++
	}

	return mmio.WriteBinary(fp, int8(12), []byte("unstructured"), int32(nc), uc, int32(len(m)), fc)
	// return mmio.WriteBinary(fp, "unstructured", int32(nc), uc, int32(len(m)), fc)
}
