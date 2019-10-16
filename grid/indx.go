package grid

import (
	"fmt"
	"log"
	"sort"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

// Indx data type array of integers
type Indx struct {
	gd *Definition
	a  map[int]int
}

// New constructor
func (r *Indx) New(fp string, rowmajor bool) {
	if r.gd == nil {
		if _, b := mmio.FileExists(fp + ".gdef"); b {
			r.getGDef(fp + ".gdef")
		} else {
			log.Fatalf("getGDef: no grid definition loaded")
		}
	}
	r.getBinary(fp, rowmajor)
}

// NewGD constructor
func (r *Indx) NewGD(bfp, gdfp string) {
	r.getGDef(gdfp)
	r.getBinary(bfp, true)
}

// NewShort constructor
func (r *Indx) NewShort(fp string, rowmajor bool) {
	if r.gd == nil {
		if _, b := mmio.FileExists(fp + ".gdef"); b {
			r.getGDef(fp + ".gdef")
		} else {
			log.Fatalf("getGDef: no grid definition loaded")
		}
	}
	r.getBinaryShort(fp, rowmajor)
}

// NewShortGD constructor
func (r *Indx) NewShortGD(bfp, gdfp string, rowmajor bool) {
	r.getGDef(gdfp)
	r.getBinaryShort(bfp, rowmajor)
}

// NewIMAP constructor
func (r *Indx) NewIMAP(imap map[int]int) {
	if r.gd == nil {
		log.Fatalf("grid.Indx.NewIMAP: grid definition needs defining\n")
	}
	r.a = make(map[int]int, len(imap))
	for k, v := range imap {
		r.a[k] = v
	}
}

// LoadGDef loads grid definition
func (r *Indx) LoadGDef(gd *Definition) {
	r.gd = gd
}

// Nvalues returns the size of the Indx
func (r *Indx) Nvalues() int {
	return len(r.a)
}

// Value returns the value of a given cell ID
func (r *Indx) Value(cid int) int {
	if v, ok := r.a[cid]; ok {
		return v
	}
	log.Fatalf("Indx.Value: no value asigned to cell ID %d", cid)
	return -1
}

// UniqueValues returns the value of a given cell ID
func (r *Indx) UniqueValues() []int {
	c, i := make([]int, len(r.a)), 0
	for _, v := range r.a {
		c[i] = v
		i++
	}
	return mmaths.UniqueInts(c)
}

// Values returns the mapped grid values
func (r *Indx) Values() map[int]int {
	return r.a
}

func (r *Indx) getGDef(fp string) {
	var err error
	r.gd, err = ReadGDEF(fp, true)
	if err != nil {
		log.Fatalf("getGDef: %v", err)
	}
}

func (r *Indx) getBinaryShort(fp string, rowmajor bool) {
	b, n, err := mmio.ReadBinaryShorts(fp, 1)
	if err != nil {
		log.Fatalf(" Indx.getBinary(): %v", err)
	}
	switch n {
	case r.gd.na:
		r.a = make(map[int]int, r.gd.na)
		log.Fatalln(" Indx.getBinary: active grids not yet supported (TODO)")
	case r.gd.nr * r.gd.nc:
		r.a = make(map[int]int, r.gd.nr*r.gd.nc)
		if rowmajor {
			for i := 0; i < n; i++ {
				r.a[i] = int(b[0][i])
			}
		} else {
			c, nr, nc := 0, r.gd.nr, r.gd.nc
			for j := 0; j < nc; j++ {
				for i := 0; i < nr; i++ {
					r.a[i*nc+j] = int(b[0][c])
					c++
				}
			}
		}
	case 2 * r.gd.nr * r.gd.nc, 2 * r.gd.na:
		// log.Fatalf(" Indx.getBinaryShort: %s is not of type short", fp)
		r.getBinary(fp, rowmajor)
	default:
		log.Fatalf(" Indx.getBinaryShort: %s does not match definition length", fp)
	}
}

func (r *Indx) getBinary(fp string, rowmajor bool) {
	b, n, err := mmio.ReadBinaryInts(fp, 1)
	if err != nil {
		log.Fatalf(" Indx.getBinary(): %v", err)
	}
	switch n {
	case r.gd.na:
		r.a = make(map[int]int, r.gd.na)
		i, lst := 0, make([]int, len(r.gd.act))
		for cid := range r.gd.act {
			lst[i] = cid
			i++
		}
		sort.Ints(lst)
		for k, cid := range lst {
			r.a[cid] = int(b[0][k])
		}
	case r.gd.nr * r.gd.nc:
		r.a = make(map[int]int, r.gd.nr*r.gd.nc)
		if rowmajor {
			for i := 0; i < n; i++ {
				r.a[i] = int(b[0][i])
			}
		} else {
			c, nr, nc := 0, r.gd.nr, r.gd.nc
			for j := 0; j < nc; j++ {
				for i := 0; i < nr; i++ {
					r.a[i*nc+j] = int(b[0][c])
					c++
				}
			}
		}
	default:
		log.Fatalf(" Indx.getBinary: grid does not match definition length")
	}
}

// ToASC creates an ascii-grid of Indx.
func (r *Indx) ToASC(fp string, ignoreActives bool) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf("Indx ToASC: %v", err)
	}
	defer t.Close()
	r.gd.ToASCheader(t)
	if r.gd.na > 0 && ignoreActives {
		c := 0
		for i := 0; i < r.gd.nr; i++ {
			for j := 0; j < r.gd.nc; j++ {
				if r.gd.act[c] {
					t.Write(fmt.Sprintf("%d ", r.a[c]))
				} else {
					t.Write("-9999 ")
				}
				c++
			}
			t.Write("\n")
		}
	} else {
		c := 0
		for i := 0; i < r.gd.nr; i++ {
			for j := 0; j < r.gd.nc; j++ {
				if v, ok := r.a[c]; ok {
					t.Write(fmt.Sprintf("%d ", v))
				} else {
					t.Write("-9999 ")
				}
				c++
			}
			t.Write("\n")
		}
	}
	return nil
}
