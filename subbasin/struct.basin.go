package subbasin

import (
	"github.com/maseology/goHydro/rainrun"
)

type basin struct {
	x    []rainrun.Lumper
	nc   int
	coll *basinCollect
}

type basinCollect struct{ spr, sae, sro, srch []float64 }

func (r *basin) update(ya, ea float64, mnt int) (qout float64) {

	ae, ro, rch := 0., 0., 0.

	for i := range r.nc {
		a, q, g := r.x[i].Update(ya, ea)
		ae += a
		ro += q
		rch += g

	}

	if r.coll != nil {
		r.coll.spr[mnt] += ya
		r.coll.sae[mnt] += ae
		r.coll.sro[mnt] += ro
		r.coll.srch[mnt] += rch
	}

	return ro
}
