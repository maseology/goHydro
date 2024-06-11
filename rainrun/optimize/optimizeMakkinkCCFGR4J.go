package optimize

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/maseology/glbopt"
	rr "github.com/maseology/goHydro/rainrun"
	"github.com/maseology/goHydro/rainrun/sample"
	"github.com/maseology/goHydro/solirrad"
	mrg63k3a "github.com/maseology/goRNG/MRG63k3a"
	mmplt "github.com/maseology/mmPlot"
	"github.com/maseology/mmio"
	"github.com/maseology/objfunc"
)

// MakkinkCCFGR4J a single or set of rainrun models
func MakkinkCCFGR4J(logfp string) {
	logger := mmio.GetInstance(logfp)

	// lat, _, err := UTM.ToLatLon(rr.Loc[1], rr.Loc[2], 17, "", true)
	// if err != nil {
	// 	log.Fatalf("%v", err)
	// }
	// si := solirrad.New(lat, math.Tan(rr.Loc[4]), rr.Loc[5])
	si := solirrad.New(43.6, 0., 0.)

	obs := make([]float64, gfrc.Ndt)
	for i, v := range gfrc.D {
		obs[i] = v.Q // [m/d]
	}

	rng := rand.New(mrg63k3a.New())
	rng.Seed(time.Now().UnixNano())

	genMakkinkCCFGR4J := func(u []float64) float64 {
		var m rr.MakkinkCCFGR4J
		m.New(append(sample.MakkinkCCFGR4J(u), gfrc.D[0].Q)...)
		m.SI = &si

		f := func(obs []float64) float64 {
			sim := make([]float64, gfrc.Ndt)
			for i, v := range gfrc.D {
				_, _, r, _ := m.Update(&v)
				sim[i] = r
			}
			return minimizer(obs[365:], sim[365:])
		}(obs)
		if math.IsNaN(f) {
			log.Fatalf("Objective function error, u: %v\n", u)
		}
		return f
	}

	uFinal, _ := glbopt.SCE(ncmplx, 10, rng, genMakkinkCCFGR4J, true)
	// uFinal, _ := glbopt.SurrogateRBF(nrbf, 10, rng, genMakkinkCCFGR4J)

	func() {
		par := []string{"x1", "x2", "x3", "x4", "tindex", "ddfc", "baseT", "tsf", "alpha", "beta"}
		pFinal := sample.MakkinkCCFGR4J(uFinal)
		fmt.Println("Optimum:")
		for i, v := range par {
			fmt.Printf(" %10s: %10.4f\t[%.4e]\n", v, pFinal[i], uFinal[i])
		}

		var m rr.MakkinkCCFGR4J
		m.SI = &si
		m.New(append(pFinal, gfrc.D[0].Q)...)
		sim, aet, bf := make([]float64, gfrc.Ndt), make([]float64, gfrc.Ndt), make([]float64, gfrc.Ndt)
		y, ep := make([]float64, gfrc.Ndt), make([]float64, gfrc.Ndt)
		txx, tnn := -math.MaxFloat64, math.MaxFloat64
		for i, v := range gfrc.D {
			yy, a, r, g := m.Update(&v)
			y[i] = yy
			ep[i] = v.Ep
			txx = math.Max(txx, v.Tx)
			tnn = math.Min(tnn, v.Tn)
			aet[i] = a
			sim[i] = r
			bf[i] = g
		}
		kge, nse, mwr2, bias := objfunc.KGE(obs[365:], sim[365:]), objfunc.NSE(obs[365:], sim[365:]), objfunc.Krause(obs[365:], sim[365:]), objfunc.Bias(obs[365:], sim[365:])
		fmt.Printf(" KGE: %.3f\tNSE: %.3f\tmon-wr2: %.3f\tBias: %.3f\n", kge, nse, mwr2, bias)

		func() {
			idt, iy, ia, iob, is, ig := make([]interface{}, gfrc.Ndt), make([]interface{}, gfrc.Ndt), make([]interface{}, gfrc.Ndt), make([]interface{}, gfrc.Ndt), make([]interface{}, gfrc.Ndt), make([]interface{}, gfrc.Ndt)
			ys, es, as, rs, gs, qs := 0., 0., 0., 0., 0., 0.
			for i, o := range obs {
				idt[i] = gfrc.DT[i]
				iy[i] = y[i]
				ia[i] = aet[i]
				iob[i] = obs[i]
				is[i] = sim[i]
				ig[i] = bf[i]
				ys += y[i]
				es += ep[i]
				as += aet[i]
				rs += sim[i]
				gs += bf[i]
				qs += o
			}
			f := 366. / float64(len(obs))
			rr.SumHydrograph(gfrc, obs, sim, bf, "MakkinkCCFGR4J")
			mmplt.ObsSim("hyd.png", obs[365:], sim[365:])
			mmplt.ObsSimFDC("fdc.png", obs[365:], sim[365:])
			mmio.WriteCSV(mmio.RemoveExtension(gfrc.FilePath)+".hydrograph.csv", "date,y,aet,obs,sim,bf", idt, iy, ia, iob, is, ig)
			sum1 := fmt.Sprintf(" y: %.3f\tpet: %.3f\taet: %.3f\trch: %.3f\ttmax: %.3f\ttmin: %.3f\tro: %.3f\tqobs: %.3f", ys*f, es*f, as*f, gs*f, txx, tnn, rs*f, qs*f)
			logger.Println(fmt.Sprintf("\nsta\t%s\n%s\nnam\t%v\nU\t%v\nP\t%v\nKGE\t%f\nNSE\t%f\nmwr2\t%f\nbias\t%f\n", mmio.FileName(gfrc.FilePath, false), sum1, par, uFinal, pFinal, kge, nse, mwr2, bias))
			fmt.Println(sum1)
		}()
	}()
}
