package subbasin

import (
	"github.com/maseology/goHydro/rainrun"
)

func (ev *Evaluator) buildBasin(nt int, collectGrids bool) []*basin {
	ns := len(ev.Sais)
	rel := make([]*basin, ns)

	for k := range ev.Sais {

		var bcol *basinCollect
		if collectGrids {
			bcol = &basinCollect{
				spr:  make([]float64, 12),
				sae:  make([]float64, 12),
				sro:  make([]float64, 12),
				srch: make([]float64, 12),
			}
		}

		x := make([]rainrun.Lumper, ev.Nsmdl[k])
		for i := range ev.Nsmdl[k] {
			x[i] = ev.Mdls[k][i]
		}

		rel[k] = &basin{
			x:    x,
			nc:   ev.Nsmdl[k],
			coll: bcol,
		}
	}

	return rel
}
