package rainrun

import (
	"fmt"

	mmplt "github.com/maseology/mmPlot"
	"github.com/maseology/objfunc"
)

// EvalPNG prints model output to a png
func EvalPNG(m Lumper) string {
	o := make([]float64, Ndt)
	s := make([]float64, Ndt)
	b := make([]float64, Ndt)
	ys, es, as, rs, gs, qs := 0., 0., 0., 0., 0., 0.
	for i, v := range FRC {
		a, r, g := m.Update(v[0], v[1])
		o[i] = v[2]
		s[i] = r
		b[i] = g
		ys += v[0]
		es += v[1]
		as += a
		rs += r
		gs += g
		qs += v[2]
	}
	f := 366. / float64(Ndt)
	stOf := fmt.Sprintf(" KGE: %.3f\tNSE: %.3f\tRMSE: %.6f\tmon-wr2: %.3f\tBias: %.3f\n", objfunc.KGE(o[365:], s[365:]), objfunc.NSE(o[365:], s[365:]), objfunc.RMSE(o[365:], s[365:]), objfunc.Krause(o[365:], s[365:]), objfunc.Bias(o[365:], s[365:]))
	stSum := fmt.Sprintf(" y: %.3f\tpet: %.3f\taet: %.3f\trch: %.3f\tro: %.3f\tqobs: %.3f\n", ys*f, es*f, as*f, gs*f, rs*f, qs*f)
	fmt.Print(stOf)
	fmt.Print(stSum)
	mmplt.ObsSim("hyd.png", o[365:], s[365:])
	mmplt.ObsSimFDC("fdc.png", o[365:], s[365:])
	SumHydrograph(o, s, b)
	SumMonthly(DT, o, s, Timestep, 1.)
	return stOf + stSum
}
