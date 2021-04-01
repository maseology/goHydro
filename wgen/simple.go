package wgen

import (
	"math/rand"
	"time"

	mrg63k3a "github.com/maseology/pnrg/MRG63k3a"
)

type WGEN struct {
	WetDry [][]float64
	rng    *rand.Rand
}

func (w *WGEN) Generate(last float64) float64 {
	if w.rng == nil {
		w.rng = rand.New(mrg63k3a.New())
		w.rng.Seed(time.Now().UnixNano())
	}
	//              |  dry  |  wet  |
	// ------------ | ----- | ----- |
	// dry previous |       |       |
	// ------------ | ----- | ----- |
	// wet previous |       |       |
	// ------------------------------
	f := w.rng.Float64()
	if last > 0. { // wet
		if f > w.WetDry[1][0] { // continual rain
			return w.rng.Float64() * .008 ////////////////// need for a better distribution function
		}
	} else {
		if f > w.WetDry[0][0] { // new rain
			return w.rng.Float64() * .01 ////////////////// need for a better distribution function
		}
	}
	return 0.
}
