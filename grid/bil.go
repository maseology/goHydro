package grid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/maseology/mmaths"
)

func (r *Real) ImportBil(fp string) error {
	b, err := os.ReadFile(fp)
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
			if (math.Log10(math.Abs(v64)) > 30 && math.Log10(math.Abs(nd)) > 30) || v64 == nd || (math.IsNaN(nd) && math.IsNaN(v64)) {
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

func (x *Indx) ImportBil(fp string) error {
	b, err := os.ReadFile(fp)
	if err != nil {
		return fmt.Errorf("ImportBil failed: %v", err)
	}
	buf := bytes.NewReader(b)
	n := len(b) / 4
	v := make([]int32, n)
	if err := binary.Read(buf, binary.LittleEndian, v); err != nil {
		return fmt.Errorf("ImportBil failed: %v", err)
	}

	// read grid into
	var nd float64
	x.GD, nd, err = ReadHdr(strings.ReplaceAll(fp, ".bil", ".hdr"))
	if err != nil {
		return fmt.Errorf("ImportBil failed: %v", err)
	}

	// build grid mapping
	cid := -1
	cids := make([]int, 0, n)
	x.A = make(map[int]int, n)
	x.GD.Coord = make(map[int]mmaths.Point)
	for i := 0; i < x.GD.Nrow; i++ {
		for j := 0; j < x.GD.Ncol; j++ {
			cid++
			v64 := float64(v[cid])
			if (math.Log10(math.Abs(v64)) > 30 && math.Log10(math.Abs(nd)) > 30) || v64 == nd || (math.IsNaN(nd) && math.IsNaN(v64)) {
				continue
			}
			x.A[cid] = int(v64)
			p := mmaths.Point{X: x.GD.Eorig + x.GD.Cwidth*(float64(j)+0.5), Y: x.GD.Norig - x.GD.Cwidth*(float64(i)+0.5)}
			x.GD.Coord[cid] = p
			cids = append(cids, cid)
		}
	}
	x.GD.ResetActives(cids)
	return nil
}

func (x *Real) ToBil(fp string) error {
	a, c := make([]float32, x.GD.Ncells()), 0
	for i := 0; i < x.GD.Nrow; i++ {
		for j := 0; j < x.GD.Ncol; j++ {
			cid := x.GD.CellID(i, j)
			if xac, ok := x.A[cid]; ok {
				a[c] = float32(xac)
			} else {
				a[c] = -9999.
			}
			c++
		}
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, a); err != nil {
		return fmt.Errorf("Real.ToBil() failed1: %v", err)
	}
	if err := os.WriteFile(fp, buf.Bytes(), 0644); err != nil { // see: https://en.wikipedia.org/wiki/File_system_permissions
		return fmt.Errorf("Real.ToBil() failed2: %v", err)
	}
	return nil
}

func (x *Indx) ToBil(fp string) error {
	a, c := make([]int32, x.GD.Ncells()), 0
	for i := 0; i < x.GD.Nrow; i++ {
		for j := 0; j < x.GD.Ncol; j++ {
			cid := x.GD.CellID(i, j)
			if xac, ok := x.A[cid]; ok {
				a[c] = int32(xac)
			} else {
				a[c] = -9999
			}
			c++
		}
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, a); err != nil {
		return fmt.Errorf("Indx.ToBil() failed1: %v", err)
	}
	if err := os.WriteFile(fp, buf.Bytes(), 0644); err != nil { // see: https://en.wikipedia.org/wiki/File_system_permissions
		return fmt.Errorf("Indx.ToBil() failed2: %v", err)
	}
	return nil
}
