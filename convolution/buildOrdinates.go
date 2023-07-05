package convolution

import "github.com/maseology/mmaths"

func buildOrdinatesFromPoints(pts []mmaths.Point, ts float64, nstep int) []float64 {
	ords := make([]float64, nstep)
	setLine := func(i int) mmaths.LineSegment {
		ln := mmaths.LineSegment{P0: pts[i], P1: pts[i+1]}
		ln.Build()
		return ln
	}

	s, cur, ln, next := 0., ts/2, setLine(0), 1
	for i := 0; i < nstep; i++ {
		ords[i] = ln.IntersectionX(cur).Y
		s += ords[i]
		cur += ts
		for {
			if next < 6 && cur > pts[next].X {
				ln = setLine(next)
				next += 1
			} else {
				break
			}
		}
	}
	for i := 0; i < nstep; i++ {
		ords[i] /= s // normalize
	}
	return ords
}
