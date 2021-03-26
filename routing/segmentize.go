package routing

import (
	"github.com/maseology/mmaths"
)

func Segmentize(plns [][][3]float64, nodethresh float64) [][][3]float64 {
	// // print for testing
	// sca, ii := mmio.NewCSVwriter("M:/vertices.csv"), 0
	// sca.WriteHead("x,y,fid,vid")
	// for i, pln := range plns {
	// 	for _, v := range pln {
	// 		sca.WriteLine(v[0], v[1], i, ii)
	// 		ii++
	// 	}
	// }
	// sca.Close()

	pts, zs, eps, meps, nvert := [][2]float64{}, []float64{}, []int{}, make(map[int]bool, len(plns)), 0
	for _, pln := range plns {
		np := len(pln)
		if np == 0 {
			continue
		}
		for _, v := range pln {
			pts = append(pts, [...]float64{v[0], v[1]})
			zs = append(zs, v[2])
		}
		eps = append(eps, nvert+np-1) // endpoint IDs
		meps[nvert+np-1] = true
		meps[nvert+np] = true // start points
		nvert += np
	}
	meps[0] = true
	var xys mmaths.XYsearch
	xys.New(pts)

	i0 := 0
	brkpnt := make(map[int]bool, len(eps))
	for _, ep := range eps {
		f := func(p int) {
			cp, _ := xys.ClosestIDs(pts[p], nodethresh)
			for _, c := range cp {
				if c == p {
					continue
				}
				if _, ok := meps[c]; ok {
					return // near-by point already an start/end point
				}
				// must be an interior point
				if c < i0 || c > p { // not part of current segment
					brkpnt[c] = true
					return // cep[0] is the closest
				}
			}
		}
		f(i0)
		f(ep)
		i0 = ep + 1
	}

	outplns := [][][3]float64{}
	i0 = 0
	for _, c := range eps {
	breaksegment:
		xycoll := [][3]float64{[...]float64{pts[i0][0], pts[i0][1], zs[i0]}}
		for i := i0 + 1; i <= c; i++ {
			xycoll = append(xycoll, [...]float64{pts[i][0], pts[i][1], zs[i]})
			if i < c {
				if _, ok := brkpnt[i]; ok {
					outplns = append(outplns, xycoll)
					i0 = i
					goto breaksegment
				}
			}
		}
		outplns = append(outplns, xycoll)
		i0 = c + 1
	}
	return outplns
}
