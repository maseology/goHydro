package grid

import (
	"fmt"
	"log"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

// Indx data type array of integers
type Indx struct {
	GD *Definition
	A  map[int]int
}

// New constructor
func (x *Indx) New(fp string, rowmajor bool) {
	if x.GD == nil {
		if _, b := mmio.FileExists(fp + ".gdef"); b {
			fmt.Println(" loading: " + fp + ".gdef")
			x.getGDef(fp + ".gdef")
		} else {
			log.Fatalf(" Indx.New: no grid definition loaded %s", fp)
		}
	}
	x.getBinary(fp, rowmajor)
}

// NewGD constructor
func (x *Indx) NewGD(bfp, gdfp string) {
	x.getGDef(gdfp)
	x.getBinary(bfp, true)
}

// NewShort constructor
func (x *Indx) NewShort(fp string, rowmajor bool) {
	if x.GD == nil {
		if _, b := mmio.FileExists(fp + ".gdef"); b {
			x.getGDef(fp + ".gdef")
		} else {
			log.Fatalf(" Indx.NewShort: no grid definition loaded")
		}
	}
	x.getBinaryShort(fp, rowmajor)
}

// NewShortGD constructor
func (x *Indx) NewShortGD(bfp, gdfp string, rowmajor bool) {
	x.getGDef(gdfp)
	x.getBinaryShort(bfp, rowmajor)
}

// // NewIMAP constructor
// func (x *Indx) NewIMAP(imap map[int]int) {
// 	if x.GD == nil {
// 		log.Fatalf(" Indx.NewIMAP: grid definition needs defining\n")
// 	}
// 	x.A = make(map[int]int, len(imap))
// 	for k, v := range imap {
// 		x.A[k] = v
// 	}
// }

// ToIndx
func (gd *Definition) ToIndx(imap map[int]int) *Indx {
	x := Indx{GD: gd, A: make(map[int]int, len(imap))}
	for k, v := range imap {
		x.A[k] = v
	}
	return &x
}

// LoadGDef loads grid definition
func (x *Indx) LoadGDef(gd *Definition) {
	x.GD = gd
}

// Nvalues returns the size of the Indx
func (x *Indx) Nvalues() int {
	return len(x.A)
}

// Value returns the value of a given cell ID
func (x *Indx) Value(cid int) int {
	if v, ok := x.A[cid]; ok {
		return v
	}
	log.Fatalf("Indx.Value: no value assigned to cell ID %d", cid)
	return -1
}

// UniqueValues returns the value of a given cell ID
func (x *Indx) UniqueValues() []int {
	c, i := make([]int, len(x.A)), 0
	for _, v := range x.A {
		c[i] = v
		i++
	}
	return mmaths.UniqueInts(c)
}

func (x *Indx) getGDef(fp string) {
	var err error
	x.GD, err = ReadGDEF(fp, true)
	if err != nil {
		log.Fatalf("getGDef: %v", err)
	}
}

func (x *Indx) getBinaryShort(fp string, rowmajor bool) {
	b, n, err := mmio.ReadBinaryShorts(fp, 1)
	if err != nil {
		log.Fatalf(" Indx.getBinary(): %v", err)
	}
	switch n {
	case x.GD.Nact:
		x.A = make(map[int]int, x.GD.Nact)
		log.Fatalln(" Indx.getBinary: active grids not yet supported (TODO)")
	case x.GD.Nrow * x.GD.Ncol:
		x.A = make(map[int]int, x.GD.Nrow*x.GD.Ncol)
		if rowmajor {
			for i := 0; i < n; i++ {
				x.A[i] = int(b[0][i])
			}
		} else {
			c, nr, nc := 0, x.GD.Nrow, x.GD.Ncol
			for j := 0; j < nc; j++ {
				for i := 0; i < nr; i++ {
					x.A[i*nc+j] = int(b[0][c])
					c++
				}
			}
		}
	case 2 * x.GD.Nrow * x.GD.Ncol, 2 * x.GD.Nact:
		// log.Fatalf(" Indx.getBinaryShort: %s is not of type short", fp)
		x.getBinary(fp, rowmajor)
	default:
		fmt.Printf("   %d %d %d", n, x.GD.Nrow*x.GD.Ncol, x.GD.Nact)
		log.Fatalf(" Indx.getBinaryShort: %s does not match definition length", fp)
	}
}

func (x *Indx) getBinary(fp string, rowmajor bool) {
	b, n, err := mmio.ReadBinaryInts(fp, 1)
	if err != nil {
		log.Fatalf(" Indx.getBinary(): %v", err)
	}
	switch n {
	case x.GD.Nact:
		x.A = make(map[int]int, x.GD.Nact)
		for i, cid := range x.GD.Sactives {
			x.A[cid] = int(b[0][i])
		}
	case x.GD.Nrow * x.GD.Ncol:
		x.A = make(map[int]int, x.GD.Nrow*x.GD.Ncol)
		if rowmajor {
			for i := 0; i < n; i++ {
				x.A[i] = int(b[0][i])
			}
		} else {
			c, nr, nc := 0, x.GD.Nrow, x.GD.Ncol
			for j := 0; j < nc; j++ {
				for i := 0; i < nr; i++ {
					x.A[i*nc+j] = int(b[0][c])
					c++
				}
			}
		}
	default:
		fmt.Println(x.GD.Nact, x.GD.Nrow*x.GD.Ncol, x.GD.Nact*4, x.GD.Nrow*x.GD.Ncol*4)
		log.Fatalf(" Indx.getBinary: grid does not match definition length %d", n)
	}
}
