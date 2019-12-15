package lia// VERSION 2

import (
	"log"
	"math"
)

func (d *Domain) addGhostNodes(n int) {
	if d.gns != nil {
		log.Fatalln("addGhostNodes error: ghost nodes already exist")
	}
	d.gns = make(map[int]*node, n)
	wkr0 := make([]worker, len(d.WKR)+n)
	copy(wkr0, d.WKR)
	d.WKR = wkr0
}

// SetHeadBC sets specified head (type 1) boundary conditions
func (d *Domain) SetHeadBC(m map[int]float64) {
	wid := len(d.WKR)
	d.addGhostNodes(len(m))
	for k, v := range m {
		fnid := d.GF.FaceCell[k]
		if fnid[0] == -1 && fnid[1] == -1 {
			log.Fatalf("LIA.SetHeadBC() error 2")
		} else if fnid[0] == -1 {
			nb := node{h: v, z: d.ns[fnid[1]].z, n: d.ns[fnid[1]].n} // ghost node
			d.gns[k] = &nb
			d.WKR[wid] = worker{
				q:  &d.qs[k],
				nb: &nb,
				nf: &d.ns[fnid[1]],
				zx: d.ns[fnid[1]].z,
				n2: math.Pow(d.ns[fnid[1]].n, 2.),
			}
		} else if fnid[1] == -1 {
			nf := node{h: v, z: d.ns[fnid[0]].z, n: d.ns[fnid[0]].n} // ghost node
			d.gns[k] = &nf
			d.WKR[wid] = worker{
				q:  &d.qs[k],
				nb: &d.ns[fnid[0]],
				nf: &nf,
				zx: d.ns[fnid[0]].z,
				n2: math.Pow(d.ns[fnid[0]].n, 2.),
			}
		} else {
			log.Fatalf("LIA.SetHeadBC() error: boundary face (possibly) already pointing to ghost node")
		}
		wid++
	}
	return
}

// SetFluxBC sets specified flux (type 2) boundary conditions
func (d *Domain) SetFluxBC(m map[int]float64) {
	for k, f := range m {
		fnid := d.GF.FaceCell[k]
		if fnid[0] == -1 && fnid[1] == -1 {
			log.Fatalf("LIA.SetFluxBC() error 2")
		} else if fnid[0] == -1 {
			d.qs[k] = flux(f)
		} else if fnid[1] == -1 {
			d.qs[k] = flux(-f)
		} else {
			log.Fatalf("LIA.SetFluxBC() error 3")
		}
	}
	return
}
