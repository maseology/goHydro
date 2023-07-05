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
	r.getGDEF(fp + ".gdef")
	r.getBinary(fp)
}

// NewGD32 constructor
func (r *Real) NewGD32(fp string, gd *Definition) {
	r.GD = gd
	r.getBinary32(fp, true)
}

func (r *Real) getGDEF(fp string) {
	var err error
	r.GD, err = ReadGDEF(fp, true)
	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (r *Real) getBinary(fp string) {
	r.A = make(map[int]float64, r.GD.Nact)
	b, n, err := mmio.ReadBinaryFloat64s(fp, 1)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if n != r.GD.Nact {
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
		log.Fatalf(" Real.getBinary(): %v", err)
	}
	switch n {
	case r.GD.Nact:
		r.A = make(map[int]float64, r.GD.Nact)
		for i, cid := range r.GD.Sactives {
			r.A[cid] = float64(b[0][i])
		}
	case r.GD.Nrow * r.GD.Ncol:
		r.A = make(map[int]float64, r.GD.Nrow*r.GD.Ncol)
		if rowmajor {
			for i := 0; i < n; i++ {
				r.A[i] = float64(b[0][i])
			}
		} else {
			c, nr, nc := 0, r.GD.Nrow, r.GD.Ncol
			for j := 0; j < nc; j++ {
				for i := 0; i < nr; i++ {
					r.A[i*nc+j] = float64(b[0][c])
					c++
				}
			}
		}
	default:
		log.Fatalf(" Real.getBinary: grid does not match definition length")
	}
}

func (r *Real) ResetToGDEF(gdeffp string) {
	var err error
	r.GD, err = ReadGDEF(gdeffp, true)
	if err != nil {
		log.Fatalf("%v", err)
	}
	newa := make(map[int]float64, r.GD.Nact)
	for _, c := range r.GD.Sactives {
		newa[c] = r.A[c]
	}
	r.A = newa
}
