package subbasin

import (
	"github.com/maseology/goHydro/forcing"
	"github.com/maseology/mmio"
)

func big(i int) string { return mmio.Thousands(int64(i)) }

func LoadDomain(mdlprfx string) (*Structure, *Subwatershed, *Mapper, *forcing.Forcing) {
	chkerr := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	strc, err := LoadGobStructure(mdlprfx + "structure.gob")
	chkerr(err)
	sws, err := LoadGobSubwatershed(mdlprfx + "subwatershed.gob")
	chkerr(err)
	mp, err := LoadGobMapper(mdlprfx + "mapper.gob")
	chkerr(err)
	frc, err := forcing.LoadGobForcing(mdlprfx + "forcing.gob")
	chkerr(err)

	return strc, sws, mp, frc
}
