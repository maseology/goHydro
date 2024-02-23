package tem

import "github.com/maseology/goHydro/grid"

func (t *TEM) GetStreamSegments(gd *grid.Definition, cid0, ccmin int) [][][]float64 {
	cc := t.ContributingCellCounts()
	c := make(map[int]bool)
	var climbRecurs func(int)
	var segs [][][]float64
	climbRecurs = func(cid int) {
		if _, ok := c[cid]; ok {
			return
		}
		c[cid] = true
		xy0 := gd.CellCentroid(cid)
		for _, i := range t.USlp[cid] {
			if cc[i] <= ccmin {
				continue
			}
			xy1 := gd.CellCentroid(i)
			segs = append(segs, [][]float64{{xy0[0], xy0[1]}, {xy1[0], xy1[1]}})
			climbRecurs(i)
		}
	}
	climbRecurs(cid0)
	return segs
}
