package grid

import "math"

// https://homepages.inf.ed.ac.uk/rbf/HIPR2/gsmooth.htm
// center cell was modified from .15018 to .15020 such that the filter summed to 1.
var FilterGaussianSmoothing = [][]float64{
	{0.00366, 0.01465, 0.02564, 0.01465, 0.00366},
	{0.01465, 0.05861, 0.09524, 0.05861, 0.01465},
	{0.02564, 0.09524, 0.15020, 0.09524, 0.02564},
	{0.01465, 0.05861, 0.09524, 0.05861, 0.01465},
	{0.00366, 0.01465, 0.02564, 0.01465, 0.00366},
}

func (g Real) Min(buffer int) map[int]float64 {
	bc := SurroundingCells(buffer)
	Anew := make(map[int]float64, len(g.A))
	findmin := func(cid int) float64 {
		r, c := g.GD.RowCol(cid)
		vn := math.MaxFloat64
		for _, brc := range bc {
			bcid := g.GD.CellID(r+brc[0], c+brc[1])
			if bv, ok := g.A[bcid]; ok {
				if bv < vn && bv != -9999. {
					vn = bv
				}
			}
		}
		return vn
	}
	for cid := range g.A {
		Anew[cid] = findmin(cid)
	}
	return Anew
}
