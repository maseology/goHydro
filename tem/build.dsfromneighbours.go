package tem

import "math"

func (t *TEM) buildDsFromNeighbours(bufs map[int][]int) map[int]int {
	ds := make(map[int]int, len(t.TEC))
	f := []float64{math.Sqrt2, 1, math.Sqrt2, 1, 1, math.Sqrt2, 1, math.Sqrt2}
	for c, tt := range t.TEC {
		ds[c] = func() int { // Single direction steepest decent (D8)
			ii := -1
			gradmax := 0.
			for i, bc := range bufs[c] {
				if bc < 0 {
					continue
				}
				if bt, ok := t.TEC[bc]; ok {
					grad := (tt.Z - bt.Z) / f[i]
					if grad < 0 {
						continue
					}
					if grad > gradmax {
						gradmax = grad
						ii = bc
					}
				} else {
					panic("buildDsFromNeighbours err1")
				}
			}
			return ii
		}()
	}
	return ds
}
