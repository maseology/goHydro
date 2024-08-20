package rainrun

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/maseology/goHydro/pet"
	"github.com/maseology/goHydro/solirrad"
	"github.com/maseology/mmio"
)

type Frc struct {
	// Dat      [][]float64 // holds forcing data // tx, tn, r, s := v[0], v[1], v[2], v[3]
	D   []Dset    // holds forcing data
	Loc []float64 // contains location info (coordinates, catchment properties, etc.)
	// DOY      []int       // holds the day of year
	DT       []time.Time // dates
	Timestep float64     // timestep in seconds
	Ndt      int         // Ndt number of timesteps
	FilePath string
}

type Dset struct{ Q, Tx, Tn, rf, sf, sm, pa, Ep float64 }

func (d *Dset) Yield() float64 { return d.rf + d.sm }

func (d *Dset) Runoff() float64 { return d.Q }

// func (d *Dset) DatArray() (tx, tn, r, s float64) { return d.Tx, d.Tn, d.rf, d.sf }

func ReadOWRC(csvfp string, cakm2, latitude float64) ([]time.Time, []Dset) {
	cms2mmd := 86.4 / cakm2

	f, err := os.Open(csvfp)
	if err != nil {
		log.Fatalf("readOWRC failed: %v\n", err)
	}
	defer f.Close()

	recs := mmio.LoadCSV(io.Reader(f), 1) // "Date","Flow","Flag","Tx","Tn","Rf","Sf","Sm","Pa"
	o, ts := make([]Dset, 0), make([]time.Time, 0)
	si := solirrad.New(latitude, 0., 0.)
	for rec := range recs {
		// fmt.Println(rec)
		t, err := time.Parse("2006-01-02", rec[0])
		doy := t.YearDay()
		if err != nil {
			log.Fatalf("readOWRC date read fail: %v\n", err)
		}
		// fmt.Println(t)
		g := func(i int) float64 {
			v, err := strconv.ParseFloat(rec[i], 64)
			if err != nil {
				if rec[i] == "NA" {
					return 0. //math.NaN()
				}
				log.Fatalf("readOWRC date read fail: value parse error: %v (%d)", err, i)
			}
			return v
		}

		ep := func() float64 {
			const (
				a = 0.75
				b = 0.0025
				c = 2.5
			)
			// tx, tn, pa := g(3), g(4), g(8)*1000.
			tx, tn, pa := g(3), g(4), 101300.
			if tx < tn {
				fmt.Printf(" tx<tn %.1f !< %.1f\n", tx, tn)
				tx, tn = tn, tx
			}
			return func(Kg float64) float64 {
				const (
					alpha = 0.61
					beta  = .001 //-1.2e-4
				)
				tm := (tx + tn) / 2.
				return pet.Makkink(Kg, tm, pa, alpha, beta)
			}(si.GlobalFromPotential(tx, tn, a, b, c, doy)) * 1000. // mm/d
		}()
		if ep < 0 {
			panic("ep less than 0")
		}

		ts = append(ts, t)
		o = append(o, Dset{ // "Date","Flow","Flag","Tx","Tn","Rf","Sf","Sm","Pa"
			Q:  g(1) * cms2mmd,
			Tx: g(3),
			Tn: g(4),
			rf: g(5),
			sf: g(6),
			sm: g(7),
			// pa: g(8),
			Ep: ep,
		})
	}

	return ts, o
}
