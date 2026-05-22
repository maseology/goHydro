package subbasin

import (
	"github.com/maseology/goHydro/convolution"
	"github.com/maseology/goHydro/rainrun"
)

type Evaluator struct {
	Schannel          []*convolution.Convolution
	Mdls              [][]rainrun.Lumper
	Outer, Sais, Smon [][]int
	Nsmdl, Dsws       []int // number of submodels per basin
	Nc, Nm            int   // number of raster cells, number of monitors
}
