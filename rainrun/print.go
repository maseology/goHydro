package rainrun

import (
	"math"
	"time"

	"github.com/maseology/mmio"
)

func SumHydrograph(o, s, g []float64) {
	// C:/Users/mason/OneDrive/R/dygraph/obssim_csv_viewer.R
	idt, io, is, ig := make([]interface{}, Ndt), make([]interface{}, Ndt), make([]interface{}, Ndt), make([]interface{}, Ndt)
	for i, t := range DT {
		idt[i] = t
		io[i] = o[i]
		is[i] = s[i]
		ig[i] = g[i]
	}
	mmio.WriteCSV("hydrograph.csv", "date,obs,sim,bf", idt, io, is, ig)
}

func SumMonthly(dt []time.Time, o, s []float64, ts, ca float64) {
	tso, tss := make(mmio.TimeSeries, len(dt)), make(mmio.TimeSeries, len(dt))
	for i, d := range dt {
		if math.IsNaN(o[i]) || math.IsNaN(s[i]) {
			continue
		}
		tso[d] = o[i]
		tss[d] = s[i]
	}
	os, _ := mmio.MonthlySumCount(tso)
	ss, _ := mmio.MonthlySumCount(tss)
	dn, dx := mmio.MinMaxTimeseries(tso)
	dti, i := make([]interface{}, len(os)*12), 0
	osi, ssi := make([]interface{}, len(os)*12), make([]interface{}, len(ss)*12)
	for y := mmio.Yr(dn.Year()); y <= mmio.Yr(dx.Year()); y++ {
		for m := mmio.Mo(1); m <= 12; m++ {
			if v, ok := os[y][m]; ok {
				if math.IsNaN(v) || math.IsNaN(ss[y][m]) {
					continue
				}
				dti[i] = time.Date(int(y), m, 15, 0, 0, 0, 0, time.UTC)
				cf := ts * 1000. / ca // sum(cms) to mm/mo
				osi[i] = v * cf
				ssi[i] = ss[y][m] * cf
				i++
			}
		}
	}
	mmio.WriteCSV("monthlysum.csv", "date,obs,sim", dti, osi, ssi)
}
