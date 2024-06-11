package optimize

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/maseology/glbopt"
	rr "github.com/maseology/goHydro/rainrun"
	"github.com/maseology/goHydro/rainrun/sample"
	mrg63k3a "github.com/maseology/goRNG/MRG63k3a"
	"github.com/maseology/mmio"
	"github.com/maseology/objfunc"
)

const (
	nrbf   = 100
	ncmplx = 200
)

var gfrc *rr.Frc // global forcing data

var minimizer = func(o, s []float64) float64 { return 1. - objfunc.NSE(o, s) }

// Optimize a single or set of rainrun models
func Optimize(ifrc *rr.Frc, mdl, logfp string) {
	logger := mmio.GetInstance(logfp)

	gfrc = ifrc
	rng := rand.New(mrg63k3a.New())
	rng.Seed(time.Now().UnixNano())

	switch mdl {
	case "Atkinson":
		func() {
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 7, rng, genAtkinson, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 7, rng, genAtkinson)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.Atkinson{}
			pFinal := sample.Atkinson(uFinal)
			sp := fmt.Sprintf("\nfinal parameters: %v\n", pFinal)
			su := fmt.Sprintf("sample space:\t%f\n", uFinal)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
		}()
	case "DawdyODonnell":
		func() {
			if gfrc.Timestep <= 0. {
				log.Fatalf("need to set timestep length for Dawdy O'Donnell simulations")
			}
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 6, rng, genDawdyODonnell, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 6, rng, genDawdyODonnell)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.DawdyODonnell{}
			pFinal := sample.DawdyODonnell(uFinal, gfrc.Timestep)
			sp := fmt.Sprintf("\nfinal parameters: %v\n", pFinal)
			su := fmt.Sprintf("sample space:\t%f\n", uFinal)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
		}()
	case "GR4J":
		func() {
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 4, rng, genGR4J, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 4, rng, genGR4J)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.GR4J{}
			pFinal := sample.GR4J(uFinal)
			sp := fmt.Sprintf("\nfinal parameters:\t%.3e\n", pFinal)
			su := fmt.Sprintf("sample space:\t\t%f\n", uFinal)
			fmt.Print(sp + su)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
			// var m rr.Lumper = &rr.GR4J{}
			// ss := sampler.NewSet(sample.GR4J()) //////////////////////////////////  TO FIX
			// pFinal := ss.Sample(uFinal)
			// fmt.Printf("\nfinal parameters: %v\n", pFinal)
			// m.New(pFinal...)
			// rr.EvalPNG(m)
		}()
	case "HBV":
		func() {
			if gfrc.Timestep <= 0. {
				// log.Fatalf("need to set timestep length for HBV simulations")
				gfrc.Timestep = 86400.
			}
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 9, rng, genHBV, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 9, rng, genHBV)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.HBV{}
			pFinal := sample.HBV(uFinal, gfrc.Timestep)
			sp := fmt.Sprintf("\nfinal parameters:\t%.3e\n", pFinal)
			su := fmt.Sprintf("sample space:\t\t%f\n", uFinal)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
		}()
	case "ManabeGW":
		func() { // check
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 5, rng, genManabeGW, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 5, rng, genManabeGW)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.ManabeGW{}
			pFinal := sample.ManabeGW(uFinal)
			sp := fmt.Sprintf("\nfinal parameters: %v\n", pFinal)
			su := fmt.Sprintf("sample space:\t\t%f\n", uFinal)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
		}()
	case "MultiLayerCapacitance":
		func() { // check
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 9, rng, genMultiLayerCapacitance, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 9, rng, genMultiLayerCapacitance)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.MultiLayerCapacitance{}
			pFinal := sample.MultiLayerCapacitance(uFinal)
			sp := fmt.Sprintf("\nfinal parameters: %v\n", pFinal)
			su := fmt.Sprintf("sample space:\t\t%f\n", uFinal)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
		}()
	case "Quinn":
		func() { // check
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 11, rng, genQuinn, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 11, rng, genQuinn)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.Quinn{}
			pFinal := sample.Quinn(uFinal)
			sp := fmt.Sprintf("\nfinal parameters: %v\n", pFinal)
			su := fmt.Sprintf("sample space:\t\t%f\n", uFinal)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
		}()
	case "SIXPAR":
		func() { // check
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 6, rng, genSIXPAR, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 6, rng, genSIXPAR)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.SIXPAR{}
			pFinal := sample.SIXPAR(uFinal)
			sp := fmt.Sprintf("\nfinal parameters: %v\n", pFinal)
			su := fmt.Sprintf("sample space:\t\t%f\n", uFinal)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
		}()
	case "SPLR":
		func() { // check (negative AET??)
			tt := time.Now()
			uFinal, _ := glbopt.SCE(ncmplx, 6, rng, genSPLR, true)
			// uFinal, _ := glbopt.SurrogateRBF(nrbf, 6, rng, genSPLR)
			elpsd := time.Since(tt)

			var m rr.Lumper = &rr.SPLR{}
			pFinal := sample.SPLR(uFinal)
			sp := fmt.Sprintf("\nfinal parameters: %v\n", pFinal)
			su := fmt.Sprintf("sample space:\t\t%f\n", uFinal)
			m.New(pFinal...)
			logger.Println(mmio.FileName(gfrc.FilePath, false))
			logger.Printf(" optimization time elapsed %v\n", elpsd)
			logger.Print(sp + su)
			logger.Println("\n" + rr.EvalPNG(m, gfrc, mdl))
		}()
	default:
		fmt.Println("unrecognized model:" + mdl)
	}
}

// // permute used to create a complete sample set of
// // every possible permutation of p dimensions and w discrete
// // values.
// func permute(fp string) {
// 	rr.LoadMET(fp, true)
// 	var m rr.Lumper = &rr.DawdyODonnell{}
// 	for i, u := range smpln.Permutations(6, 3) {
// 		fmt.Println(i, u)
// 		m.New(sample.DawdyODonnell(u, gfrc.Timestep)...)
// 		if math.IsNaN(eval(m)) {
// 			panic("NaN")
// 		}
// 	}
// }
