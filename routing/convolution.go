package routing

import (
	"fmt"
	"math"
	"sort"

	"github.com/maseology/goHydro/convolution"
	"github.com/maseology/goHydro/tem"
	"github.com/maseology/mmaths/slice"
)

type Convroute struct {
	W, Z float64
	I    int
}

// func BuildFromTEM(dem *tem.TEM, thrsh, theta float64, maxrad int) [][]Convroute {
func BuildFromTEM(dem *tem.TEM, alpha, fmax float64, maxrad int) [][]Convroute {
	// maxrad is the "reach" i.e., the number of cells upstream and downstream to collect
	// thrsh and theta are model parameters relating slope to convolution transfer function shape
	cids, ds := dem.DownslopeContributingAreaIDs(-1)
	climb := func(cid int) map[int]int {
		c := make(map[int]int)
		var climbRecurs func(int, int)
		climbRecurs = func(cid, radius int) {
			if radius > maxrad {
				return
			}
			if _, ok := c[cid]; ok {
				return
			}
			c[cid] = -radius
			for _, i := range dem.USlp[cid] {
				climbRecurs(i, radius+1)
			}
		}
		climbRecurs(cid, 0)
		return c
	}
	fall := func(dcid int) map[int]int {
		c := make(map[int]int)
		var fallRecurs func(int, int)
		fallRecurs = func(cid, radius int) {
			if cid < 0 {
				return
			}
			if radius > maxrad {
				return
			}
			if _, ok := c[cid]; ok {
				return
			}
			c[cid] = radius
			if d, ok := ds[cid]; ok {
				fallRecurs(d, radius+1)
			}
		}
		fallRecurs(dcid, 1)
		return c
	}
	// skew := func(slp, thrsh, theta float64) float64 {
	skew := func(slp, alpha, fmax float64) float64 {
		// u := 1 - math.Exp(slp*slp/-alpha) // Gaussian decay
		u := 1 - math.Exp(math.Sqrt(slp)/-alpha) // Gaussian decay
		return (fmax-.5)*u + .5
		// return .8
		// s := math.Sqrt(slp)
		// f := 0.
		// if s > thrsh {
		// 	f := (s - thrsh) * math.Tan(theta)
		// 	if f > 1. {
		// 		f = 1.
		// 	}
		// }
		// return (f + 1.) / 2 // skew should range [.5,1], otherwise up-gradient flow will occur
	}

	xr := make(map[int]int, len(cids))
	for a, c := range cids {
		xr[c] = a
	}

	base := maxrad*2 + 1
	o := make([][]Convroute, len(cids))
	slpcoll := make([]float64, len(cids))
	cccc := 0
	for a, c := range cids {
		rads := climb(c)
		outlet := false
		if d, ok := ds[c]; ok {
			for k, v := range fall(d) {
				rads[k] = v
			}
		} else {
			outlet = true
		}

		base, skew, lag := float64(base), skew(dem.TEC[c].G, alpha, fmax), 0.
		slpcoll[a] = math.Sqrt(dem.TEC[c].G)
		if skew < .8 {
			cccc++
		}
		tri := convolution.Triangular(lag, base+lag, skew*base+lag)
		na := len(rads)
		if outlet {
			na++
		}
		o[a] = make([]Convroute, 0, na)
		s := 0.
		for cc, r := range rads {
			s += tri[maxrad+r]
			o[a] = append(o[a], Convroute{tri[maxrad+r], dem.TEC[cc].Z, xr[cc]})
		}
		if outlet {
			w := 0.
			for i := 1; i <= maxrad; i++ {
				w += tri[i+maxrad]
			}
			s += w
			o[a] = append(o[a], Convroute{w, dem.TEC[c].Z - 1., -1})
		}
		sort.Slice(o[a], func(i, j int) bool {
			return o[a][i].Z < o[a][j].Z
		})
		for i := range o[a] {
			o[a][i].W /= s
		}
	}
	m, sd := slice.MeanSD(slpcoll)
	fmt.Println(cccc, float64(cccc)/float64(len(o)), m, sd)
	return o
}
