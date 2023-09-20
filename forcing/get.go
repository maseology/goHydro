package forcing

import (
	"fmt"
	"math"
	"time"

	"github.com/maseology/goHydro/gmet"
	"github.com/maseology/mmio"
)

func GetForcings(mids []int, intvl float64, offset int, ncfp, prfx string) Forcing {
	tt := time.Now()

	fmt.Println(" loading: " + ncfp)
	g := func(fp string) *gmet.GMET {
		var g *gmet.GMET
		var err error
		switch mmio.GetExtension(fp) {
		case ".nc":
			vars := []string{"precipitation_amount"}
			// vars := []string{
			// 	// "air_temperature",
			// 	// "air_pressure",
			// 	// "relative_humidity",
			// 	// "wind_speed",
			// 	"water_potential_evaporation_amount",
			// 	"rainfall_amount",
			// 	// "snowfall_amount",
			// 	"surface_snow_melt_amount",
			// }
			g, err = gmet.LoadNC(fp, prfx, vars)
		case ".csv":
			g, err = gmet.LoadCsv(fp, prfx, "precipitation_amount")
		default:
			panic("unknown frc type")
		}
		if err != nil {
			panic(err)
		}
		return g
	}(ncfp)

	// fmt.Printf("  Dates available (may not be in order): %v to %v\n", g.Ts[0], g.Ts[g.Nts-1])
	// collect sequential dates
	ts, xt, nts := func() ([]time.Time, []int, int) {
		d := make(map[int64]int, g.Nts)
		for j, t := range g.Ts {
			d[t.Unix()] = j
		}

		func() {
			dt, cdt := g.Ts[0], 0
			for {
				if _, ok := d[dt.Unix()]; !ok {
					// fmt.Printf("   > missing date %v\n", dt)
					d[dt.Unix()] = -1
					cdt++
				}
				dt = dt.Add(time.Second * time.Duration(intvl))
				if dt.After(g.Ts[g.Nts-1]) {
					break
				}
			}
			if cdt > 0 {
				fmt.Printf("     Total missing dates = %d\n", cdt)
			}
		}()

		o, x := make([]time.Time, 0, len(d)), make([]int, 0, len(d))
		dt := g.Ts[0]
		for {
			if xx, ok := d[dt.Unix()]; ok {
				x = append(x, xx)
			} else {
				panic("FORC sequential dates error")
			}
			o = append(o, dt)
			dt = dt.Add(time.Second * time.Duration(intvl))
			if dt.After(g.Ts[g.Nts-1]) {
				break
			}
		}
		fmt.Printf("  Dates available: %v to %v in %d steps\n", o[0], o[len(o)-1], len(o))
		return o, x, len(o)
	}()

	// collect subset of met IDs
	mmid := func() map[int]int {
		if len(g.Sids) == 1 {
			mmid := make(map[int]int, len(mids))
			for _, s := range mids {
				mmid[s] = 0
			}
			return mmid
		} else {
			mmid := make(map[int]int, len(g.Sids))
			for i, s := range g.Sids {
				mmid[s] = i
			}
			return mmid
		}
	}()

	// collect data
	pre := g.GetAllData("precipitation_amount")
	// rf := g.GetAllData("rainfall_amount")
	// sm := g.GetAllData("surface_snow_melt_amount")
	// eao := g.GetAllData("water_potential_evaporation_amount")

	ys := make([][]float64, len(mids))
	// es, el :=  make([][]float64, len(mids)), .001
	min0 := func(x float64, s string, m int, t time.Time) float64 {
		if math.IsNaN(x) {
			fmt.Printf("   > %s Nan (%d): %v -- set to zero\n", s, m, t)
			return 0.
		}
		if x < 0. {
			return 0.
		}
		return x / 1000. // to [m]
	}

	for ii, s := range mids {
		if i, ok := mmid[s]; ok {
			ya := make([]float64, nts)
			// ea := make([]float64, nts)
			for j, t := range ts {
				jj := xt[j] + offset // offset to end of timestep
				if jj >= 0 && jj < len(pre[i]) {
					ya[j] = min0(pre[i][jj], "pre", s, t)
					// ya[j] = min0(rf[i][jj], "rf", s, t) + min0(sm[i][jj], "sm", s, t)
				}
				// if jj >= 0 && eao[i][jj] >= 0. {
				// 	ea[j] = eao[i][jj] / 1000. // to [m]
				// 	el = ea[j]
				// } else {
				// 	ea[j] = el // infilling with last
				// }
			}
			ys[ii] = ya
			// es[ii] = ea
		} else {
			panic("loadForcings error: met index does not contain sws index")
		}
	}

	frc := Forcing{
		T:  ts,
		Ya: ys,
		// Ea: es,
		// XR:          cixr,
		IntervalSec: intvl,
	}

	fmt.Printf(" Forcing loaded - %v\n", time.Since(tt))
	return frc
}