package grid

import (
	"log"

	"github.com/maseology/mmio"
)

// Real data type array
type Real struct {
	GD *Definition
	A  map[int]float64
}

// New constructor
func (r *Real) New(fp string) {
	r.getGDef(fp + ".gdef")
	r.getBinary(fp)
}

func (r *Real) getGDef(fp string) {
	var err error
	r.GD, err = ReadGDEF(fp, true)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (r *Real) getBinary(fp string) {
	r.A = make(map[int]float64, r.GD.Na)
	b, n, err := mmio.ReadBinaryFloats(fp, 1)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if n != r.GD.Na {
		log.Fatalf(" grid does not match definition length")
	}
	c := 0
	if len(r.GD.Sactives) > 0 {
		for _, i := range r.GD.Sactives {
			r.A[i] = b[0][c]
			c++
		}
	}
}
