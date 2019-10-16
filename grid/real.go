package grid

import (
	"log"

	"github.com/maseology/mmio"
)

// Real data type array
type Real struct {
	gd *Definition
	a  map[int]float64
}

// New constructor
func (r *Real) New(fp string) {
	r.getGDef(fp + ".gdef")
	r.getBinary(fp)
}

func (r *Real) getGDef(fp string) {
	var err error
	r.gd, err = ReadGDEF(fp, true)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (r *Real) getBinary(fp string) {
	r.a = make(map[int]float64, r.gd.na)
	b, n, err := mmio.ReadBinaryFloats(fp, 1)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if n != r.gd.na {
		log.Fatalf(" grid does not match definition length")
	}
	c := 0
	if len(r.gd.act) > 0 {
		for i, a := range r.gd.act {
			if a {
				r.a[i] = b[0][c]
				c++
			}
		}
	}
}
