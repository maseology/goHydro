package grid

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

func (r *Real) ImportAsc(fp string) error {
	// read grid into
	var nd float64 // no data value
	var err error
	r.GD, nd, err = ReadHdr(fp)
	if err != nil {
		return fmt.Errorf("ImportAsc header read fail: %v", err)
	}

	// read lines
	lns, err := mmio.ReadTextLines(fp)
	if err != nil {
		return fmt.Errorf("ImportAsc data read fail: %v", err)
	}
	d := make([]float64, r.GD.Ncells())
	nc, na, f := 0, 0, true
	for _, ln := range lns {
		sp := strings.Split(mmio.RemoveWhiteSpaces(strings.TrimSpace(ln)), " ")
		if f {
			if _, err := strconv.ParseFloat(sp[0], 64); err != nil {
				continue // skip header lines
			}
			f = false
		}

		for k, s := range sp {
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				println(k)
			}
			d[nc] = v
			nc++
			if v != nd {
				na++
			}
		}
	}
	if nc != r.GD.Ncells() {
		return fmt.Errorf("ImportAsc data read fail: ndat = %d, ncells = %d", nc, r.GD.Ncells())
	}

	// build grid mapping
	cid, naa := -1, 0
	r.A, r.GD.Sactives = make(map[int]float64, na), make([]int, na)
	r.GD.Coord = make(map[int]mmaths.Point)
	r.GD.Nact = na
	r.GD.act = make(map[int]bool, na)
	for i := 0; i < r.GD.Nrow; i++ {
		for j := 0; j < r.GD.Ncol; j++ {
			cid++
			if d[cid] == nd {
				continue
			}
			r.A[cid] = d[cid]
			r.GD.Coord[cid] = mmaths.Point{X: r.GD.Eorig + r.GD.Cwidth*(float64(j)+0.5), Y: r.GD.Norig - r.GD.Cwidth*(float64(i)+0.5)}
			r.GD.Sactives[naa] = cid
			r.GD.act[cid] = true
			naa++
		}
	}
	if na != naa {
		return fmt.Errorf("ImportAsc data collect fail: na = %d, naa = %d", na, naa)
	}
	return nil
}

func (r *Real) ToAsc(fp string) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf(" Real.ToASC: %v", err)
	}
	defer t.Close()
	r.GD.ToASCheader(t)
	c := 0
	for i := 0; i < r.GD.Nrow; i++ {
		for j := 0; j < r.GD.Ncol; j++ {
			if v, ok := r.A[c]; ok {
				t.Write(fmt.Sprintf("%f ", v))
			} else {
				t.Write("-9999 ")
			}
			c++
		}
		t.Write("\n")
	}
	return nil
}
