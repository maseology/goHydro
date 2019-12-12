package lia

import "math"

type face struct {
	forth                []int   // orthogonal faces
	q, dx, n2, zx        float64 // parameters and variables: q: flux;
	nfrom, nto, ffw, fbw int     // node and face identifiers
}

func (f *face) initialize(n0, n1 *node, cellsize float64) {
	f.dx = cellsize
	f.zx = math.Max(n0.z, n1.z)
	f.n2 = math.Pow(((n0.n + n1.n) / 2.), 2.)
}

func (f *face) nodeIDs() (int, int) {
	return f.nfrom, f.nto // from node id, to node id
}

func (f *face) isBoundary() bool {
	return len(f.forth) == 0
}

func (f *face) isInactive() bool {
	return f.nfrom == -1 && f.nto == -1
}

func (f *face) idColl() []int {
	in1 := make([]int, 8)
	in1[0] = f.nfrom
	in1[1] = f.nto
	in1[2] = f.fbw
	in1[3] = f.ffw
	for i := 0; i < 4; i++ {
		in1[4+i] = f.forth[i]
	}
	return in1
}

func (f *face) updateFlux(s *state, dt float64) {
	hf := math.Max(s.n0h, s.n1h) - f.zx // the depth at the interface between cells
	if hf <= 0.000001 {
		f.q = 0.
	} else {
		// qmag := math.Abs(f.q) // de Almeda etal 2012
		qmag := math.Sqrt(math.Pow(f.q, 2.) + math.Pow(s.avgOrthoFlux, 2.))                  // eq. 8
		f.q = theta*f.q + .5*(1.-theta)*(s.fflux+s.bflux) - 9.80665*hf*dt*(s.n1h-s.n0h)/f.dx // eq. 7 numer
		f.q /= 1. + 9.80665*dt*f.n2*qmag/math.Pow(hf, 2.33333)                               // eq. 7 denom
	}
}
