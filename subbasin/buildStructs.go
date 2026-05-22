package subbasin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/maseology/goHydro/forcing"
	"github.com/maseology/goHydro/grid"
	"github.com/maseology/mmio"
)

func BuildStructs(controlFP string,
	iksat func(*grid.Definition, []int, []int) ([]float64, []int),
	xlu func(*grid.Definition, string, []int) SurfaceSet,
	intvl float64,
	epsg int,
) (*Structure, *Mapper, *Subwatershed, *forcing.Forcing, string, float64, int, bool) {

	// Defaults
	strmkm2 := 1. // contibuting area (km2) for when cells are deem "stream cells"

	fmt.Printf("loading control file %s\n", controlFP)
	var mdlprfx, gdefFP, hdemFP, swsFP, luFP, sgFP, gwzFP, ncfp string
	cid0, lakfrac, gwids := -1, -1., []int{}
	crop := false

	getFilePaths := func(fp string) {
		var err error
		ins := mmio.NewInstruct(fp)
		mdlprfx = ins.Param["prfx"][0]
		if !(mmio.GetFileDir(mdlprfx) != "." && mmio.DirExists(mmio.GetFileDir(mdlprfx))) {
			mdlprfx = mmio.GetFileDir(fp) + "/" + mdlprfx
		}

		gdefFP = ins.Param["gdeffp"][0] // grid definition
		hdemFP = ins.Param["hdemfp"][0] // hydrologically-corrected DEM
		swsFP = ins.Param["swsfp"][0]   // subbasin raster
		luFP = ins.Param["lufp"][0]     // land use
		sgFP = ins.Param["sgfp"][0]     // surficial geology

		// optional files
		if gfp, ok := ins.Param["gwzfp"]; ok {
			gwzFP = gfp[0] // groundwater zone id
		}
		if mfp, ok := ins.Param["ncfp"]; ok {
			ncfp = mfp[0] // input climate data (netCDF)
		}

		if _, ok := ins.Param["intvl"]; ok { // set time step (seconds)
			if intvl, err = strconv.ParseFloat(ins.Param["intvl"][0], 64); err != nil {
				panic(err)
			}
		}
		if _, ok := ins.Param["strmkm2"]; ok { // set stream channel contributing area
			if strmkm2, err = strconv.ParseFloat(ins.Param["strmkm2"][0], 64); err != nil {
				panic(err)
			}
		}

		if _, ok := ins.Param["epsg"]; ok { // outlet cell ID, <0 keeps while model domain
			if epsg, err = strconv.Atoi(ins.Param["epsg"][0]); err != nil {
				panic(err)
			}
		}
		if _, ok := ins.Param["cid0"]; ok { // outlet cell ID, <0 keeps while model domain
			if cid0, err = strconv.Atoi(ins.Param["cid0"][0]); err != nil {
				panic(err)
			}
			crop = cid0 >= 0
		}
		if _, ok := ins.Param["gwid"]; ok {
			cid0 = -1
			if onegwz, err := strconv.Atoi(ins.Param["gwid"][0]); err == nil {
				gwids = []int{onegwz}
			} else {
				rn := []rune(ins.Param["gwid"][0])
				if string(rn[0]) == "[" && string(rn[len(rn)-1]) == "]" {
					trm := strings.Split(string(rn[1:len(rn)-1]), ",")
					gwids = make([]int, len(trm))
					for i, v := range trm {
						if gi, err := strconv.Atoi(v); err != nil {
							panic(fmt.Errorf("builder.go: gwid read error: %v", err))
						} else {
							gwids[i] = gi
						}
					}
				} else {
					panic(fmt.Errorf("builder.go: gwid read error: %s", ins.Param["gwid"][0]))
				}
			}
			crop = true
		}
		if _, ok := ins.Param["lakefrac"]; ok {
			if lakfrac, err = strconv.ParseFloat(ins.Param["lakefrac"][0], 64); err != nil {
				panic(err)
			}
		}

		gdefFP = mmio.RelativeFileCheck(fp, gdefFP)
		hdemFP = mmio.RelativeFileCheck(fp, hdemFP)
		swsFP = mmio.RelativeFileCheck(fp, swsFP)
		luFP = mmio.RelativeFileCheck(fp, luFP)
		sgFP = mmio.RelativeFileCheck(fp, sgFP)
		gwzFP = mmio.RelativeFileCheck(fp, gwzFP)
		if len(ncfp) > 0 {
			ncfp = mmio.RelativeFileCheck(fp, ncfp)
		}
	}

	getFilePaths(controlFP)
	chkdir := mmio.GetFileDir(mdlprfx) + "/check/"
	mmio.MakeDir(chkdir)
	chkdir += mmio.FileName(mdlprfx, true) // adding prefix

	////////////////////////////////////////
	// BUILD
	////////////////////////////////////////

	println("\nbuilding model structure..")
	strc := buildSTRC(gdefFP, hdemFP, cid0)

	println("\nsetting grid mappings..")
	mp := strc.buildMapper(luFP, sgFP, gwzFP, iksat, xlu, strmkm2)

	println("\nbuilding sub-watersheds (computational queuing)..")
	sws := strc.loadSWS(swsFP)
	sws.buildComputationalOrder1()
	println("   gathering sub-watershed routing info..")
	sws.BuildReaches(&strc, &mp, sinuosity*strc.GD.Cwidth)

	////////////////////////////////////////
	// ADJUST
	////////////////////////////////////////

	// re-project groundwater zones to sub-watersheds
	println(" > re-mapping unique groundwater zones to subwatersheds..")
	mp.Fngwc, mp.Igw = sws.remapGWzones(&mp)

	// set Lake HRUs
	if lakfrac > 0 {
		println(" > re-mapping lakes to subwatersheds..")
		sws.remapLakes(&mp, lakfrac)
	}

	////////////////////////////////////////
	// SUBSETTING
	////////////////////////////////////////
	if len(gwids) > 0 { // by select gwzones
		println(" > sub-setting domain to select subwatersheds..")
		subsetByGWzones(&strc, &sws, &mp, gwids)
	}

	////////////////////////////////////////
	// SUMMARIZE
	////////////////////////////////////////

	if len(chkdir) > 0 {
		println("\nBuilding summary rasters\n==================================")
		strc.Checkandprint(chkdir, crop)
		mp.Checkandprint(strc.GD, float64(strc.Nc), chkdir, crop)
		sws.Checkandprint(strc.GD, strc.Cids, float64(strc.Nc), chkdir, crop)
	}

	////////////////////////////////////////
	// CLIMATE FORCINGS
	////////////////////////////////////////

	frc := func(fp string) *forcing.Forcing {
		if len(fp) == 0 {
			return nil
		}
		if _, ok := mmio.FileExists(fp); ok {
			println("\nLoading forcings..")
			frc, err := forcing.LoadGobForcing(fp)
			if err != nil {
				panic(err)
			}
			return frc
		}
		var frc forcing.Forcing
		switch mmio.GetExtension(ncfp) {
		case ".nc":
			fmt.Printf("\n > load forcings from %s..\n", ncfp)
			vars := []string{
				"water_potential_evaporation_amount", // PE
				"rainfall_amount",
				"surface_snow_melt_amount",
			}
			frc = forcing.GetForcings(sws.Isws, intvl, 0, ncfp, "", vars) // sws id refers to the climate lists
		case "":
			return nil
		default:
			fmt.Printf(" Load forcing ERROR: unknown file type: %s.  File %s not created.", ncfp, fp)
			return nil
		}
		frc.ToBil(strc.GD, strc.Cids, sws.Sais, chkdir, crop)
		if err := frc.SaveGobForcing(fp); err != nil {
			panic(err)
		}
		return &frc
	}(mdlprfx + "forcing.gob")

	return &strc, &mp, &sws, frc, mdlprfx, intvl, epsg, crop
}
