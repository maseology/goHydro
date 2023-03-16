package mesh

import (
	"sort"

	"github.com/maseology/mmio"
)

type HSTRAT struct {
	Nam              string
	Nn, Ne, Nly, Epl int
	Nxyz, Vxyz       [][]float64 // node coordinates, velocities
	Exr, Nxr         [][]int     // element-node & node-element cross-reference
	Nh               []float64   // nodal heads
	Hgeo             []Material  // material properties
}

type Material struct {
	N, Ss, H0 float32
	K         []float32
}

func ReadHSTRAT(fp string, prnt bool) (*HSTRAT, error) {

	b := mmio.OpenBinary(fp)
	if mmio.ReadString(b) != "prism" {
		panic("mesh ReadHSTRAT todo")
	}

	nam := mmio.ReadString(b)
	nly := int(mmio.ReadInt32(b))
	nps := int(mmio.ReadInt32(b))
	nnds := nps * (nly + 1)
	epl := int(mmio.ReadInt32(b))
	nels := epl * nly
	typ := int(mmio.ReadInt32(b))
	typ2 := typ / 2
	if typ2 != 3 {
		panic("mesh ReadHSTRAT todo")
	}
	minthick := mmio.ReadFloat64(b)
	_ = minthick

	nds := make([][]float64, nnds)
	for i := 0; i < nnds; i++ {
		nds[i] = []float64{mmio.ReadFloat64(b), mmio.ReadFloat64(b), mmio.ReadFloat64(b)}
	}

	els := make([][]int, nels)
	for i := 0; i < nels; i++ {
		ii := make([]int, typ)
		for j := 0; j < typ; j++ {
			ii[j] = int(mmio.ReadInt32(b)) - 1
		}
		els[i] = ii
	}

	nxr := make([][]int, nels)
	for eid, nids := range els {
		for _, nid := range nids {
			nxr[nid] = append(nxr[nid], eid)
		}
	}

	hgeo := make([]Material, nels)
	for i := 0; i < nels; i++ {
		kdim := int(mmio.ReadInt8(b))
		k := make([]float32, kdim)
		for j := 0; j < kdim; j++ {
			k[j] = mmio.ReadFloat32(b)
		}
		hgeo[i] = Material{
			K:  k,
			N:  mmio.ReadFloat32(b),
			Ss: mmio.ReadFloat32(b),
		}
	}

	if int(mmio.ReadInt32(b)) != len(nds) {
		panic("mesh ReadHSTRAT todo: zones")
	}

	nh0 := make([]float64, nnds)
	for i := 0; i < nnds; i++ {
		// hgeo[i].H0 = float32(mmio.ReadFloat64(b))
		nh0[i] = mmio.ReadFloat64(b)
	}

	if int(mmio.ReadInt32(b)) != len(nds)*3 {
		panic("mesh ReadHSTRAT fluxes must be included")
	}
	v := make([][]float64, nnds)
	for i := 0; i < nnds; i++ {
		vv := make([]float64, 3)
		for j := 0; j < 3; j++ {
			vv[j] = mmio.ReadFloat64(b)
		}
		v[i] = []float64{vv[0], vv[1], vv[2]}
	}

	if !mmio.ReachedEOF(b) {
		panic("mesh ReadHSTRAT todo: EOF not reached")
	}

	// // check
	// func() {
	// 	ncw, nccw := 0, 0
	// 	for _, nids := range els {
	// 		xs, ys := make([]float64, len(nids)), make([]float64, len(nids))
	// 		for i, nid := range nids {
	// 			nxyz := nds[nid]
	// 			xs[i] = nxyz[0]
	// 			ys[i] = nxyz[1]
	// 		}
	// 		if spatial.IsClockwise(xs, ys) {
	// 			ncw++
	// 		} else {
	// 			nccw++
	// 		}
	// 	}
	// 	if nccw > 0 && ncw > 0 {
	// 		panic("dis-ordered nodes")
	// 	}
	// }()

	return &HSTRAT{
		Nam:  nam,
		Nxyz: nds,
		Vxyz: v,
		Nh:   nh0,
		Exr:  els,
		Nxr:  nxr,
		Nn:   nnds,
		Ne:   nels,
		Epl:  epl,
		// Nsl:    nly + 1,
		Nly:  nly,
		Hgeo: hgeo,
	}, nil
}

func (h *HSTRAT) BuildElementalConnectivity(cardinalOnly bool) map[int][]int {
	o := make(map[int][]int, h.Ne)

	unique := func(intSlice []int) []int { // https://www.golangprograms.com/remove-duplicate-values-from-slice.html
		keys := make(map[int]bool)
		list := []int{}
		for _, entry := range intSlice {
			if _, value := keys[entry]; !value {
				keys[entry] = true
				list = append(list, entry)
			}
		}
		return list
	}

	if !cardinalOnly {
		// version 1: all adjacent elements, no ordering
		for eid, nids := range h.Exr {
			var l []int
			for _, nid := range nids {
				l = append(l, h.Nxr[nid]...)
			}
			u := unique(l)
			o[eid] = make([]int, 0, len(u)-1)
			for _, uu := range u {
				if uu != eid {
					o[eid] = append(o[eid], uu)
				}
			}
		}
	} else {
		// version 2: cardinal elements, ordered: [laterals]-bottom-top
		for eid, nids := range h.Exr {
			d := make(map[int]int)
			for _, nid := range nids {
				for _, eid2 := range h.Nxr[nid] {
					d[eid2] += 1
				}
			}
			delete(d, eid)
			for k, v := range d {
				if v < 3 {
					delete(d, k)
				}
			}
			u := make([]int, 0, len(d))
			for k := range d {
				u = append(u, k)
			}

			// ordering [laterals]-bottom-top
			sort.Ints(u)
			hastop := eid+h.Epl == u[len(u)-1]
			hasbot := eid-h.Epl == u[0]
			if hasbot && hastop {
				u = append(u[1:len(u)-1], u[0], u[len(u)-1])
			} else if hasbot {
				u = append(u[1:], u[0], -1)
			} else if hastop { // top only
				u = append(u[:len(u)-1], -1, u[len(u)-1])
			} else {
				panic("shouldn't occur unless this is a 1-layered model")
			}
			o[eid] = u
		}
	}
	return o
}
