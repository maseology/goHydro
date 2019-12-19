package lia // VERSION 2

import "math"

type face struct{ q float64 }

type node struct{ z, h, n float64 }

//             	direction of +ive flow
//          ------------------------------>
//            qul                      qur
// O ---------------------- O ---------------------- O
// |                        |                        |
// |                        |                        |
// | qb        nb           q           nf        qf |
// |                        |                        |
// |                        |                        |
// O ---------------------- O ---------------------- O
//            qll                      qlr
type worker struct {
	q, qb, qf, qur, qul, qll, qlr *face
	nb, nf                        *node
	zx, n2                        float64
}

func (w *worker) getFlux(theta, dt, dx float64) float64 {
	hf := math.Max(w.nb.h, w.nf.h) - w.zx // the depth at the interface between cells
	if hf <= 0.000001 {
		return 0.
	}
	if w.qb == nil { // ghost node
		q, nbh, nfh := w.q.q, w.nb.h, w.nf.h
		qmag := math.Abs(q)
		q = q - g*hf*dt*(nfh-nbh)/dx                   // eq. 7 numer
		q /= 1. + g*dt*w.n2*qmag/math.Pow(hf, 2.33333) // eq. 7 denom
		return q
	}
	q, qb, qf, nbh, nfh := w.q.q, w.qb.q, w.qf.q, w.nb.h, w.nf.h
	qorth := float64(w.qur.q+w.qul.q+w.qll.q+w.qlr.q) / 4.
	// qmag := math.Abs(q) // de Almeda etal 2012
	qmag := math.Sqrt(math.Pow(q, 2.) + math.Pow(qorth, 2.))   // eq. 8
	q = theta*q + (1.-theta)*(qf+qb)/2. - g*hf*dt*(nfh-nbh)/dx // eq. 7 numer
	q /= 1. + g*dt*w.n2*qmag/math.Pow(hf, 2.33333)             // eq. 7 denom
	return q
}
