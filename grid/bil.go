package grid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/maseology/mmaths"
)

func (r *Real) ImportBil(fp string) error {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("ImportBil failed: %v", err)
	}
	buf := bytes.NewReader(b)
	n := len(b) / 4
	v := make([]float32, n)
	if err := binary.Read(buf, binary.LittleEndian, v); err != nil {
		return fmt.Errorf("ImportBil failed: %v", err)
	}

	// read grid into
	var nd float64
	r.GD, nd, err = ReadHdr(strings.ReplaceAll(fp, ".bil", ".hdr"))
	if err != nil {
		return fmt.Errorf("ImportBil failed: %v", err)
	}

	// build grid mapping
	cid := -1
	cids := make([]int, 0, n)
	r.A = make(map[int]float64, n)
	r.GD.Coord = make(map[int]mmaths.Point)
	for i := 0; i < r.GD.Nrow; i++ {
		for j := 0; j < r.GD.Ncol; j++ {
			cid++
			v64 := float64(v[cid])
			if v64 == nd {
				continue
			}
			r.A[cid] = v64
			p := mmaths.Point{X: r.GD.Eorig + r.GD.Cwidth*(float64(j)+0.5), Y: r.GD.Norig - r.GD.Cwidth*(float64(i)+0.5)}
			r.GD.Coord[cid] = p
			cids = append(cids, cid)
		}
	}
	r.GD.ResetActives(cids)
	return nil
}
