package grid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"

	"github.com/maseology/mmaths"
)

func (r *Real) ImportBil(fp string) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		fmt.Printf("ImportBil failed: %v\n", err)
		return
	}
	buf := bytes.NewReader(b)
	n := len(b) / 4
	v := make([]float32, n)
	if err := binary.Read(buf, binary.LittleEndian, v); err != nil {
		fmt.Printf("ImportBil failed: %v\n", err)
		return
	}

	r.A = make(map[int]float64, n)
	for i := 0; i < n; i++ {
		r.A[i] = float64(v[i])
	}

	// read grid into
	var nd float64
	r.GD, nd, err = ReadHdr(fp)
	if err != nil {
		fmt.Printf("ImportBil failed: %v\n", err)
		return
	}

	// build grid mapping
	cid, na := -1, 0
	r.GD.Coord = make(map[int]mmaths.Point)
	for i := 0; i < r.GD.Nrow; i++ {
		for j := 0; j < r.GD.Ncol; j++ {
			cid++
			if float64(v[cid]) == nd {
				continue
			}
			p := mmaths.Point{X: r.GD.Eorig + r.GD.Cwidth*(float64(j)+0.5), Y: r.GD.Norig - r.GD.Cwidth*(float64(i)+0.5)}
			r.GD.Coord[cid] = p
			r.GD.Sactives = append(r.GD.Sactives, cid)
			na++
		}
	}
	r.GD.Nact = na
}
