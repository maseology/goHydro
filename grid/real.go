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

// NewGD32 constructor
func (r *Real) NewGD32(fp string, gd *Definition) {
	r.GD = gd
	r.getBinary32(fp, true)
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

func (r *Real) getBinary32(fp string, rowmajor bool) {
	b, n, err := mmio.ReadBinaryFloat32s(fp, 1)
	if err != nil {
		log.Fatalf(" Indx.getBinary(): %v", err)
	}
	switch n {
	case r.GD.Na:
		r.A = make(map[int]float64, r.GD.Na)
		for i, cid := range r.GD.Sactives {
			r.A[cid] = float64(b[0][i])
		}
	case r.GD.Nr * r.GD.Nc:
		r.A = make(map[int]float64, r.GD.Nr*r.GD.Nc)
		if rowmajor {
			for i := 0; i < n; i++ {
				r.A[i] = float64(b[0][i])
			}
		} else {
			c, nr, nc := 0, r.GD.Nr, r.GD.Nc
			for j := 0; j < nc; j++ {
				for i := 0; i < nr; i++ {
					r.A[i*nc+j] = float64(b[0][c])
					c++
				}
			}
		}
	default:
		log.Fatalf(" Indx.getBinary: grid does not match definition length")
	}
}
