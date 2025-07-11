package rainrun

import (
	"fmt"
	"time"

	mmplt "github.com/maseology/mmPlot"
	"github.com/maseology/objfunc"
)

// EvalPNG prints model output to a png
func EvalPNG(m Lumper, frc *Frc, prfx string) string {
	o := make([]float64, frc.Ndt)
	s := make([]float64, frc.Ndt)
	b := make([]float64, frc.Ndt)
	ys, es, as, rs, gs, qs := 0., 0., 0., 0., 0., 0.
	tt := time.Now()
	for i, v := range frc.D {
		y := v.Yield()
		a, r, g := m.Update(y, v.Ep)
		o[i] = v.Q
		s[i] = r
		b[i] = g
		ys += y
		es += v.Ep
		as += a
		rs += r
		gs += g
		qs += v.Q
	}
	f := 366. / float64(frc.Ndt)
	stOf := fmt.Sprintf(" KGE: %.3f\tNSE: %.3f\tRMSE: %.6f\tmon-wr2: %.3f\tBias: %.3f\n", objfunc.KGE(o[365:], s[365:]), objfunc.NSE(o[365:], s[365:]), objfunc.RMSE(o[365:], s[365:]), objfunc.Krause(o[365:], s[365:]), objfunc.Bias(o[365:], s[365:]))
	stSum := fmt.Sprintf(" y: %.3f\tpet: %.3f\taet: %.3f\trch: %.3f\tro: %.3f\tqobs: %.3f\n", ys*f, es*f, as*f, gs*f, rs*f, qs*f)
	stElapsed := fmt.Sprintf(" run-time for %d timesteps: %v\n", frc.Ndt, time.Since(tt))
	fmt.Print(stOf)
	fmt.Print(stSum)
	fmt.Print(stElapsed)
	mmplt.ObsSim(prfx+".hyd.png", o[365:], s[365:])
	mmplt.ObsSimFDC(prfx+".fdc.png", o[365:], s[365:])
	SumHydrograph(frc, o, s, b, prfx)
	SumMonthly(frc.DT, o, s, frc.Timestep, 1., prfx)
	return stOf + stSum + stElapsed
}
