package sample

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/maseology/UTM"
	rr "github.com/maseology/goHydro/rainrun"
	"github.com/maseology/goHydro/solirrad"
	mrg63k3a "github.com/maseology/goRNG/MRG63k3a"
	"github.com/maseology/montecarlo"
)

// Sample samples a rainrun model
func Sample(frc rr.Frc, nsmpl int, fitness func(o, s []float64) float64) ([][]float64, []float64) {

	lat, _, err := UTM.ToLatLon(frc.Loc[1], frc.Loc[2], 17, "", true)
	if err != nil {
		log.Fatalf("%v", err)
	}
	si := solirrad.New(lat, math.Tan(frc.Loc[4]), frc.Loc[5])

	obs := make([]float64, frc.Ndt)
	for i, v := range frc.D {
		obs[i] = v.Q // [m/d]??
	}

	rng := rand.New(mrg63k3a.New())
	rng.Seed(time.Now().UnixNano())

	ndim := 10
	gen := func(u []float64, i int) float64 {
		var m rr.MakkinkCCFGR4J
		m.New(append(MakkinkCCFGR4J(u), frc.D[0].Q)...)
		m.SI = &si

		f := func(obs []float64) float64 {
			sim := make([]float64, frc.Ndt)
			for i, v := range frc.D {
				_, _, r, _ := m.Update(&v)
				sim[i] = r
			}
			return fitness(obs[365:], sim[365:])
		}(obs)
		if math.IsNaN(f) {
			// log.Fatalf("Objective function error, u: %v\n", u)
			return -9999.
		}
		return f
	}

	return montecarlo.GenerateSamples(gen, ndim, nsmpl, runtime.GOMAXPROCS(0))
}
