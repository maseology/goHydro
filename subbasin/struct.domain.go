package subbasin

import "github.com/maseology/goHydro/forcing"

type Domain struct {
	Strc                 *Structure
	Sws                  *Subwatershed
	Mpr                  *Mapper
	Frc                  *forcing.Forcing
	Obs                  *Observations
	Nc, Nt, Nm, Ng, Cid0 int
	Prfx, Root           string
}
