package convolution

import "github.com/maseology/mmaths"

// func buildOrdinatesFromPoints(pts []mmaths.Point, ts float64, nstep int) []float64 {
// 	const steps = 1000
// 	xs, ys := make([]float64, steps), make([]float64, steps)
// 	setLine := func(i int) mmaths.LineSegment {
// 		ln := mmaths.LineSegment{P0: pts[i], P1: pts[i+1]}
// 		ln.Build()
// 		return ln
// 	}

// 	ln, next := setLine(0), 1
// 	for i := 0; i < steps; i++ {
// 		x := float64(i) * pts[len(pts)-1].X / steps
// 		xs[i] = x
// 		for {
// 			if next < len(pts)-1 && x > pts[next].X {
// 				ln = setLine(next)
// 				next += 1
// 			} else {
// 				break
// 			}
// 		}
// 		ys[i] = ln.IntersectionX(x).Y
// 	}

// 	ords, s, stp := make([]float64, nstep-1), 0., 1
// 	for i := 0; i < steps; i++ {
// 		if float64(stp)*ts < xs[i] {
// 			ords[stp-1] = s
// 			stp += 1
// 			s = 0.
// 		}
// 		s += ys[i]
// 	}
// 	s = 0.
// 	for _, o := range ords {
// 		s += o
// 	}
// 	for i := range ords {
// 		ords[i] /= s // normalize
// 	}
// 	return ords
// }

func buildOrdinatesFromPoints(pts []mmaths.Point, ts float64, nstep int) []float64 {
	if nstep <= 1 {
		return []float64{1.}
	}

	ords := make([]float64, nstep)
	setLine := func(i int) mmaths.LineSegment {
		ln := mmaths.LineSegment{P0: pts[i], P1: pts[i+1]}
		ln.Build()
		return ln
	}

	s, cur, ln, next := 0., ts/2, setLine(0), 1
	chk := func() {
		for {
			if next < 6 && cur > pts[next].X {
				ln = setLine(next)
				next += 1
			} else {
				break
			}
		}
	}
	chk()
	for i := 0; i < nstep; i++ {
		ords[i] = ln.IntersectionX(cur).Y
		s += ords[i]
		cur += ts
		chk()
	}
	for i := 0; i < nstep; i++ {
		ords[i] /= s // normalize
	}
	return ords
}
