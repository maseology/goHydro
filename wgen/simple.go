package wgen

import (
	"math/rand"
	"time"
)

type WGEN struct {
	WetDry [][]float64
	rng    *rand.Rand
}

// src = mrg63k3a.New() or src = rand.NewSource(seed)
func New(src rand.Source) *WGEN {
	var w WGEN
	w.rng = rand.New(src)
	w.rng.Seed(time.Now().UnixNano())
	return &w
}

func (w *WGEN) Generate(last float64) float64 {
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
