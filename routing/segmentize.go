package routing

import (
	"sort"

	"github.com/maseology/mmaths"
)

func Segmentize(plns [][][3]float64, nodethresh float64) [][][3]float64 {
	pts := [][2]float64{}
	for _, pln := range plns {
		for _, v := range pln {
			pts = append(pts, [...]float64{v[0], v[1]})
		}
	}
	xys := xySearch(pts)
	_ = xys
	panic("todo: segmantize")
}

func xySearch(pts [][2]float64) int {
	xs, ys := make([]float64, len(pts)), make([]float64, len(pts))
	var xi, yi mmaths.IndexedSlice
	xi.New(xs)
	yi.New(ys)
	sort.Sort(xi)
	sort.Sort(yi)

	return 0
}
