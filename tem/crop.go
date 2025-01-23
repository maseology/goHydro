package tem

import "github.com/maseology/goHydro/grid"

// // TEM topologic elevation model
// type TEM struct {
// 	TEC  map[int]TEC
// 	USlp map[int][]int
// }

// NumCells number of cells that make up the TEM
func (t *TEM) ClipToActives(gd *grid.Definition) TEM {
	o := TEM{make(map[int]TEC), make(map[int][]int)}
	for c, tec := range t.TEC {
		if gd.IsActive(c) {
			o.TEC[c] = tec
			u := make([]int, 0, len(t.USlp[c]))
			for _, uc := range t.USlp[c] {
				if gd.IsActive(uc) {
					u = append(u, uc)
				}
			}
			o.USlp[c] = u
		}
	}
	return o
}
