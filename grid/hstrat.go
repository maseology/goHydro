package grid

import (
	"fmt"
	"log"

	"github.com/maseology/mmio"
)

type HSTRAT struct {
	Nam   string
	Nlay  int
	Cells map[int]Cell
}

type Cell struct {
	Top, Bottom, N, Ss, Sy, H0 float32
	K                          []float32
}

func ReadHSTRAT(fp string, print bool) (*HSTRAT, error) {
	big := func(i int) string {
		return mmio.Thousands(int64(i))
	}

	gd, err := ReadGDEF(mmio.RemoveExtension(fp)+".gdef", true)
	if err != nil {
		log.Fatalf(" ReadHSTRAT gdef read error: %v", err)
	}
	fmt.Printf(" HSTRAT-GDEF read: %s cells (%d rows, %d columns), %s actives\n", big(gd.Ncells()), gd.Nrow, gd.Ncol, big(gd.Nact))

	b := mmio.OpenBinary(fp)
	if mmio.ReadString(b) != "grid" {
		panic("todo")
	}

	nam := mmio.ReadString(b)
	cs, nlay := func() (map[int]Cell, int) {
		cs := make(map[int]Cell)
		dnlay := make(map[int]bool)
		nc := gd.Ncells()
		for {
			cid, ok := mmio.ReadInt32check(b)
			if !ok {
				break // EOF
			}
			top := mmio.ReadFloat32(b)
			nl := int(mmio.ReadInt8(b))
			for ly := 0; ly < nl; ly++ {
				lcid := int(cid) + ly*nc
				bottom := mmio.ReadFloat32(b)
				c := Cell{
					Top:    top,
					Bottom: bottom,
					N:      mmio.ReadFloat32(b),
					Ss:     mmio.ReadFloat32(b),
					Sy:     mmio.ReadFloat32(b),
					H0:     mmio.ReadFloat32(b),
				}
				top = bottom
				nk := int(mmio.ReadInt8(b))
				c.K = make([]float32, nk)
				for i := 0; i < nk; i++ {
					c.K[i] = mmio.ReadFloat32(b) // k[d]
				}
				cs[lcid] = c
			}
			// fmt.Println(top, nl)
			dnlay[nl] = true
		}

		// check uniform layering
		nlys := make([]int, 0, len(dnlay))
		for k := range dnlay {
			nlys = append(nlys, k)
		}
		if len(nlys) == 1 {
			return cs, nlys[0]
		}
		return cs, -1

	}()

	return &HSTRAT{
		Nam:   nam,
		Nlay:  nlay,
		Cells: cs,
	}, nil
}
