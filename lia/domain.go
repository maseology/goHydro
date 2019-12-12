package lia

import (
	"log"
	"math"
	"sync"

	"github.com/maseology/goHydro/grid"
)

const (
	alpha  = 0.7
	theta  = 0.7
	tresid = 0.00001
)

// Domain is a model domain for a local inertial approximation to the 2D SWE solver
// ref: de Almeida, G.A.M., P. Bates, 2013. Applicability of the local intertial approximation of the shallow water equations to flood modeling. Water Resources Research 49: 4833-4844.
// see also: - de Almeida Bates Freer Souvignet 2012 Improving the stability of a simple formulation of the shallow water equations for 2-D flood modeling
//           - Sampson etal 2012 An automated routing methodology to enable direct rainfall in high resolution shallow water models
// similar in theory to LISFLOOD-FP
type Domain struct {
	gf  *grid.Face
	gxr map[int]int // cell id to index
	fs  []face
	ns  []node
	st  []state
	r   []float64 // vertical influx [m/s]
	fxr [][]int   // face to node
	bf  []bool    // boundary face
	// fs            map[int]*face
	// ns            map[int]*node
	// st            map[int]*state
	// r             map[int]float64 // vertical influx [m/s]
	// fxr map[int][]int // face to node
	// bf            map[int]bool    // boundary face
	tcum, dt, dx float64 // cumulative time, time-step, space-step
	nc, nf, nfi  int     // number of cells, number of faces, number of internal faces
}

// Build a new global LIA model
func (d *Domain) Build(gd *grid.Definition, z, h0, n map[int]float64) {
	d.dx = gd.CellWidth() // goHydro grid.Definition (currently) only supports square-uniform grids
	d.gf = grid.NewFace(gd)
	d.gxr = gd.CellIndexXR() // cell id to index cross-ref
	d.nc = gd.Nactives()

	// build nodes
	d.ns = make([]node, d.nc)
	for _, c := range gd.Sactives {
		if _, ok := z[c]; !ok { // surface elevation
			log.Fatalf("Elevation not provided for cell ID %d", c)
		}
		if _, ok := h0[c]; !ok { // initial head (same datum as surface elevation)
			log.Fatalf("Initial head not provided for cell ID %d", c)
		}
		if _, ok := n[c]; !ok { // Manning's n roughness coef.
			log.Fatalf("Manning n not provided for cell ID %d", c)
		}
		d.ns[c] = node{z: z[c], h: h0[c], n: n[c], fid: d.gf.CellFace[c]}
	}

	// build faces
	d.nf, d.nfi = d.gf.Nfaces, 0
	d.fs = make([]face, d.nf)
	d.bf = make([]bool, d.nf)
	d.fxr = make([][]int, d.nf)
	d.st = make([]state, d.nf)
	for i := 0; i < d.nf; i++ {
		fc1 := newFace(d.gf, i)
		if fc1.isInactive() {
			continue
		}
		d.bf[i] = fc1.isBoundary()
		d.fs[i] = *fc1
		if !fc1.isBoundary() {
			d.fxr[i] = fc1.idColl()
			d.st[i] = state{}
			d.nfi++
		}
	}

	for k, f := range d.fs {
		if d.bf[k] {
			continue
		}
		n0, n1 := f.nodeIDs()
		f.initialize(&d.ns[n0], &d.ns[n1], d.dx)
		d.fs[k] = f
	}
}

func (d *Domain) setCurrentState() {
	dmax := 0. //-math.MaxFloat64 // the maximum water depth over the domain
	for _, n := range d.ns {
		if n.h-n.z > dmax {
			dmax = (n.h - n.z)
		}
	}
	if dmax > 0. {
		d.dt = alpha * d.dx / math.Sqrt(9.80665*dmax) // eq.12
	} else {
		print("setCurrentState check: domain is dry.")
	}

	type kv struct {
		i int
		s state
	}
	chst := make(chan kv, d.nfi)
	for i, s := range d.st {
		if d.fxr[i] == nil { // boundary or inactive face
			continue
		}
		go func(i int, s state) {
			s.n0h = d.ns[d.fxr[i][0]].h
			s.n1h = d.ns[d.fxr[i][1]].h
			s.bflux = d.fs[d.fxr[i][2]].q
			if len(d.fxr[i]) == 3 { // ghost node boundary condition
				s.avgOrthoFlux = 0.
			} else {
				s.fflux = d.fs[d.fxr[i][3]].q
				qorth := 0.
				for j := 4; j < 8; j++ {
					qorth += d.fs[d.fxr[i][j]].q
				}
				s.avgOrthoFlux = qorth / 4. // eq. 9-10 average orthogonal flux
			}
			chst <- kv{i, s}
		}(i, s)
	}
	for i := 0; i < d.nf; i++ {
		if d.fxr[i] != nil { // internal face
			kv := <-chst
			d.st[kv.i] = kv.s
		}
	}
	close(chst)
}

func (d *Domain) updateFluxes() {
	var wg sync.WaitGroup
	for k := range d.fs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			d.fs[i].updateFlux(&d.st[i], d.dt)
		}(k)
	}
	wg.Wait()
}

func (d *Domain) updateHeads() float64 {
	d1 := d.dt / d.dx // math.Pow(d.dt/d.dx,2.) // error in equation 11, see eq. 20 in de Almeda etal 2012
	type kv struct {
		i  int
		dh float64
	}
	ch := make(chan kv, d.nc)
	for i := 0; i < d.nc; i++ {
		go func(i int, n *node) {
			dh := d1 * (d.fs[n.fid[2]].q - d.fs[n.fid[0]].q + d.fs[n.fid[3]].q - d.fs[n.fid[1]].q) // eq.11
			if d.r != nil {
				print("updateHeads TODO")
				// if v, ok := d.r[k]; ok {
				// 	dh += v
				// }
			}
			ch <- kv{i, dh}
		}(i, &d.ns[i])
	}
	if len(d.ns) > d.nc {
		log.Fatalf("updateHeads todo: ghostnodes")
		// if k >= d.nc { // ghost node boundary condition   ////////////////////////////////////  assumes orderd map ///////////////////////
		// 	break
		// }
	}

	resid, aresid := 0., 0.
	for i := 0; i < d.nc; i++ {
		kv := <-ch
		adh := math.Abs(kv.dh)
		if adh > aresid {
			aresid = adh
			resid = kv.dh
		}
		d.ns[kv.i].h += kv.dh
	}
	close(ch)

	return resid
}
