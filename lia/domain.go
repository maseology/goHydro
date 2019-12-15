package lia // VERSION 2

import (
	"log"
	"math"

	"github.com/maseology/goHydro/grid"
)

const (
	dhmin = 0.00001 // minimum of maximum global head change threshold (for steady-state simulations only)
	g     = 9.80665
	dmin  = 0.001 // minimum depth for timestep calculation
)

// Domain is a model domain for a local inertial approximation to the 2D SWE solver
// ref: de Almeida, G.A.M., P. Bates, 2013. Applicability of the local intertial approximation of the shallow water equations to flood modeling. Water Resources Research 49: 4833-4844.
// see also: - de Almeida Bates Freer Souvignet 2012 Improving the stability of a simple formulation of the shallow water equations for 2-D flood modeling
//           - Sampson etal 2012 An automated routing methodology to enable direct rainfall in high resolution shallow water models
// similar in theory to LISFLOOD-FP
type Domain struct {
	GF                         *grid.Face
	WKR                        []worker
	qs                         []flux
	ns                         []node
	gns                        map[int]*node // ghost nodes (indexed by face id)
	tcum, dt, dx, Alpha, Theta float64       // cumulative time; time-step; space-step; timestep limiter; numerical diffusion (Theta<1)
	// fgxr                       map[int]int // face to ghost node xr

	// // gxr                        map[int]int // cell id to index
	r []float64 // vertical influx [m/s]
	// fxr                        [][]int   // face to node
	// bf                         []bool    // boundary face
	// nn, ngn, nf, nfi           int       // number of nodes/cells, number of ghost/sacrificial nodes, number of faces, number of internal faces
}

// Build a new global LIA model
func (d *Domain) Build(gd *grid.Definition, z, h0, n map[int]float64) {
	d.dx = gd.Cw // goHydro grid.Definition (currently) only supports square-uniform grids
	d.GF = grid.NewFace(gd)

	// d.nn = gd.Na

	// build nodes
	d.ns = make([]node, gd.Na)
	gxr := gd.CellIndexXR() // cell id to index cross-ref
	for _, c := range gd.Sactives {
		i := gxr[c]
		if _, ok := z[c]; !ok { // surface elevation
			log.Fatalf("Elevation not provided for cell ID %d", c)
		}
		if _, ok := h0[c]; !ok { // initial head (same datum as surface elevation)
			log.Fatalf("Initial head not provided for cell ID %d", c)
		}
		if _, ok := n[c]; !ok { // Manning's n roughness coef.
			log.Fatalf("Manning n not provided for cell ID %d", c)
		}
		d.ns[i] = node{z: z[c], h: h0[c], n: n[c]}
	}

	// build faces
	nf := d.GF.Nfaces
	d.qs = make([]flux, nf)
	fxr := make(map[int]int, nf)
	nfi := 0
	for fid := 0; fid < nf; fid++ {
		fc := d.GF.FaceCell[fid]
		nidb := fc[0]
		nidf := fc[1]
		var q flux
		d.qs[fid] = q
		if nidb == -1 || nidf == -1 { // boundary
			continue
		}
		fxr[nfi] = fid
		nfi++
	}

	// link to workers
	d.WKR = make([]worker, nfi)
	for i := 0; i < nfi; i++ {
		fid := fxr[i]
		fc := d.GF.FaceCell[fid]
		nidb := fc[0]
		nidf := fc[1]
		// if nidb == -1 || nidf == -1 {
		// 	continue // boundary node
		// }
		// if !gd.IsActive(nidb) || gd.IsActive(nidf) {
		// 	continue // boundary node
		// }
		//   1
		// 2   0
		//   3
		var fidqb, fidqf, fidqur, fidqul, fidqll, fidqlr int
		if d.GF.IsUpwardFace(fid) {
			fidqb = d.GF.CellFace[nidb][3]
			fidqf = d.GF.CellFace[nidf][1]
			fidqur = d.GF.CellFace[nidf][2]
			fidqul = d.GF.CellFace[nidb][2]
			fidqll = d.GF.CellFace[nidb][0]
			fidqlr = d.GF.CellFace[nidf][0]
		} else {
			fidqb = d.GF.CellFace[nidb][2]
			fidqf = d.GF.CellFace[nidf][0]
			fidqur = d.GF.CellFace[nidf][1]
			fidqul = d.GF.CellFace[nidb][1]
			fidqll = d.GF.CellFace[nidb][3]
			fidqlr = d.GF.CellFace[nidf][3]
		}
		nb := &d.ns[nidb]
		d.WKR[i] = worker{
			q:   &d.qs[fid],
			qb:  &d.qs[fidqb],
			qf:  &d.qs[fidqf],
			qur: &d.qs[fidqur],
			qul: &d.qs[fidqul],
			qll: &d.qs[fidqll],
			qlr: &d.qs[fidqlr],
			nb:  nb,
			nf:  &d.ns[nidf],
			zx:  math.Max(d.ns[nidb].z, d.ns[nidf].z),
			n2:  math.Pow(((d.ns[nidb].n + d.ns[nidf].n) / 2.), 2.),
		}
		// // check
		// fmt.Printf("%d %p %p -- ", i, &d.ns[i], d.WKR[i].nb)
		// d.WKR[i].nb.h += 10.
		// fmt.Printf("%f %f\n", d.ns[i].h, d.WKR[i].nb.h)
	}
}

// func (d *Domain) setCurrentState() {
// 	// set timestep
// 	dmax := dmin //-math.MaxFloat64 // the maximum water depth over the domain
// 	for _, n := range d.ns {
// 		// fmt.Println(i, n.h, n.z)
// 		if n.h-n.z > dmax {
// 			dmax = n.h - n.z
// 		}
// 	}
// 	d.dt = d.Alpha * d.dx / math.Sqrt(g*dmax) // eq.12
// 	d.tcum += d.dt

// 	// type kv struct {
// 	// 	i int
// 	// 	s state
// 	// }
// 	// chst := make(chan kv, d.nfi)
// 	var wg sync.WaitGroup
// 	for i := range d.st {
// 		if d.fxr[i] == nil { // boundary or inactive face
// 			continue
// 		}
// 		wg.Add(1)
// 		go func(i int, s *state) {
// 			defer wg.Done()
// 			s.n0h = d.ns[d.fxr[i][0]].h
// 			s.n1h = d.ns[d.fxr[i][1]].h
// 			s.bflux = d.fs[d.fxr[i][2]].q
// 			if len(d.fxr[i]) == 3 { // ghost node boundary condition
// 				s.avgOrthoFlux = 0.
// 				s.fflux = s.bflux
// 			} else {
// 				s.fflux = d.fs[d.fxr[i][3]].q
// 				qorth := 0.
// 				for j := 4; j < 8; j++ {
// 					qorth += d.fs[d.fxr[i][j]].q
// 				}
// 				s.avgOrthoFlux = qorth / 4. // eq. 9-10 average orthogonal flux
// 			}
// 			// chst <- kv{i, s}
// 		}(i, &d.st[i])
// 	}
// 	// for i := 0; i < d.nf; i++ {
// 	// 	if d.fxr[i] != nil { // internal face
// 	// 		kv := <-chst
// 	// 		d.st[kv.i] = kv.s
// 	// 	}
// 	// }
// 	// close(chst)
// 	wg.Wait()
// }

// // func (d *Domain) updateFluxes() {
// // 	type kv struct {
// // 		i int
// // 		f face
// // 	}
// // 	chf := make(chan kv, d.nfi)
// // 	dt, th := d.dt, d.Theta
// // 	for k, f := range d.fs {
// // 		if d.bf[k] {
// // 			continue
// // 		}
// // 		go func(i int, f face, s *state) {
// // 			f.updateFlux(s, dt, th)
// // 			chf <- kv{i, f}
// // 		}(k, f, &d.st[k])
// // 	}
// // 	for i := 0; i < d.nfi; i++ {
// // 		kv := <-chf
// // 		d.fs[kv.i] = kv.f
// // 	}
// // 	close(chf)
// // }

// func (d *Domain) updateFluxes() {
// 	var wg sync.WaitGroup
// 	for k := range d.fs {
// 		if d.bf[k] {
// 			continue
// 		}
// 		wg.Add(1)
// 		go func(i int) {
// 			defer wg.Done()
// 			d.fs[i].updateFlux(&d.st[i], d.dt, d.Theta)
// 		}(k)
// 	}
// 	wg.Wait()
// }

// func (d *Domain) updateHeads() float64 {
// 	d1 := d.dt / d.dx // math.Pow(d.dt/d.dx,2.) // error in equation 11, see eq. 20 in de Almeda etal 2012
// 	// type kv struct {
// 	// 	i  int
// 	// 	dh float64
// 	// }
// 	// ch := make(chan kv, d.nn)
// 	ch := make(chan float64, d.nn)
// 	for i := 0; i < d.nn; i++ {
// 		go func(i int, n *node) {
// 			dh := d1 * (d.fs[n.fid[2]].q - d.fs[n.fid[0]].q + d.fs[n.fid[3]].q - d.fs[n.fid[1]].q) // eq.11
// 			if d.r != nil {
// 				print("updateHeads TODO: nodal source/sinks (rain/infiltration")
// 				// if v, ok := d.r[k]; ok {
// 				// 	dh += v
// 				// }
// 			}
// 			// ch <- kv{i, dh}
// 			n.h += dh
// 			ch <- dh
// 		}(i, &d.ns[i])
// 	}
// 	// if len(d.ns) > d.nn {
// 	// 	log.Fatalf("updateHeads todo: ghostnodes")
// 	// 	// if k >= d.nc { // ghost node boundary condition   ////////////////////////////////////  assumes orderd map ///////////////////////
// 	// 	// 	break
// 	// 	// }
// 	// }

// 	// dhx, adhx := 0., 0.
// 	// for i := 0; i < d.nn; i++ {
// 	// 	kv := <-ch
// 	// 	adh := math.Abs(kv.dh)
// 	// 	if adh > adhx {
// 	// 		adhx = adh
// 	// 		dhx = kv.dh
// 	// 	}
// 	// 	d.ns[kv.i].h += kv.dh
// 	// }
// 	// close(ch)
// 	dhx, adhx := 0., 0.
// 	for i := 0; i < d.nn; i++ {
// 		dh := <-ch
// 		adh := math.Abs(dh)
// 		if adh > adhx {
// 			adhx = adh
// 			dhx = dh
// 		}
// 	}
// 	close(ch)

// 	return dhx
// }
