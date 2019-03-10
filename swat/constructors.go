package swat

import "math"

const (
	nsl     = 50  // number of soilzone layers
	lythick = 10. // layer thickness [mm]
)

/*
go implementaiton of the SWAT model
ref: Neitsch, S.L., J.G. Arnold, J.R., Kiniry, J.R. Williams, 2011. Soil and Water Assessment Tool: Theoretical Documentation Version 2009 (September 2011). 647pp.

specifications:
	1. designed for long-term daily simulation
	2. currently only applying water balance, i.e., no sediment or water quality
	3. SCS CN is adjusted according to soil moisture (ICN=0)
	4. snowpack is modelled externally and thus melt is added as an input, so snowpack (sublimation, etc.) is not modelled here
	5. assumes a uniform 0.5m soil zone depth, subdivided into 50 1cm layers (see const above)
	X. Not included:
		- Penman-Monteith is not used; therefore transpiration is not modelled (pg.135)
		- vertisols are not included, i.e., no bypass flow (pg.152)
		- perched water table (pg.158)
		- lateral flow (pg.160)
		- no evaporation or pumping from shallow gw reservoirs (pg.176)
		- no percolation to deep aquifers (i.e., no gw sink) (pg.178)
		- no shallow aquifer baseflow threshold (GWQMIN/aqt) (pg.174)
		- using (HYMO) variable storage routing method (pg.433)
		- no in-line transmission losses, evaporation, bank storage Sec7 Ch1

*/

// New SWAT SubBasin constructor
// SUBKM: area of subbasin [km2]
// SLSUBBSN: (L_slp) average slope length [m]
// CHL: (L) longest tributary channel length in subbasin [km]
// CHS: (slp_ch) average slope of tributary channels [m/m]
// CHN: (n) Manning's n value for tributary channels
// CHW: (W_bankfull) width of channel top at bank [m]
// CHD: (depth_bankfull) depth of water wht filled to bank [m]
// GWDELAY: (delta_gw) delay time for aquifer recharge [days]
// GWQMN: (aq_shthr) threshold water level in aquifer for baseflow [mm]
// ALPHABF: (alpha_bf) baseflow recession coeficient (1/k)
func (b *SubBasin) New(HRUs []HRU, Chn Channel, SUBKM, SLSUBBSN, CHL, CHS, CHN, GWDELAY, ALPHABF float64) {
	b.ca = SUBKM // subbasin contributing area [km²]
	b.dgw = GWDELAY
	b.aqt = 0. // GWQMN (no shallow aquifer baseflow threshold (GWQMIN/aqt))
	b.agw = ALPHABF

	// build HRUs and tconc, add channel element
	b.chn = &Chn
	ftot := 0.
	b.hru = make([]*HRU, len(HRUs))
	for i, u := range HRUs {
		ftot += u.f
		b.hru[i] = &u
	}
	if math.Abs(1.-ftot) > 0.001 {
		for _, u := range b.hru {
			u.f /= ftot // nomalize hru fractions
		}
	}
	for _, u := range b.hru {
		u.tconc = tconc(u.slp, SLSUBBSN, u.ovn, CHL, CHS, CHN, SUBKM*u.f)
	}
}

// tconc returns the time of concentration to subbasin outlet [hr]
// lengths [m]; slopes [m/m]
func tconc(slp, lslp, ovn, lch, sch, nch, carea float64) float64 {
	tov := math.Pow(lslp, 0.6) * math.Pow(ovn, 0.6) / math.Pow(slp, 0.3) / 18.              // pg.111
	tch := 0.62 * lch * math.Pow(nch, 0.75) / math.Pow(carea, 0.125) / math.Pow(sch, 0.375) // pg.113
	return tov + tch                                                                        // [hr]
}

// New SWAT HRU constructor
// HRUFR: fraction of subbasin area contained in HRU
// HRUSLP: (slp) average slope steepness [m/m]
// OVN: (n) Manning's n value for overland flow
// CN2: moisture condition II curve number
// SURLAG: surface runoff lag coefficient [0,15]
// CV: aboveground biomass and residue [kg/ha]
// IWATABLE: high water table code: set to true when seasonal high water table present
func (m *HRU) New(sz SoilLayer, HRUFR, HRUSLP, OVN, CN2, SURLAG, CV float64, IWATABLE bool) {
	m.f = HRUFR
	m.slp = HRUSLP
	m.ovn = OVN
	m.surlag = SURLAG
	m.cov = math.Exp(-5.0e-5 * CV) // soil cover index (pg.135)
	m.iwt = IWATABLE
	m.sz = make([]SoilLayer, nsl)
	for i := 0; i < nsl; i++ {
		m.sz[i] = sz
	}
	m.cn.New(CN2, sz.fc*float64(nsl), sz.sat*float64(nsl), HRUSLP) // fc, sat [mm]; SLP as fraction [m/m]
}

// New SWAT soil zone layer constructor
// CLAY: (m_c) percent clay content
// SOLBD: bulk density of soil (Mg/m³=g/cm³)
// SOLAWC: available water capacity as fraction of total soil volume [-]
// SOLK: (ksat) saturated hydraulic conductivity [mm/hr]
func (sl *SoilLayer) New(CLAY, SOLBD, SOLAWC, SOLK float64) {
	n := 1. - SOLBD/2.65              // porosity
	sl.wp = 0.4 * CLAY * SOLBD / 100. // water content at wilting point as fraction of total soil volume (pg.149)
	sl.fc = (sl.wp + SOLAWC)          // water content at field capacity as a fraction of total soil volume (pg.150)
	sl.frz = false
	// converting to [mm]
	sl.sat = n * lythick            // the amount of water in the soil profile when completely saturated [mm] using constant 1cm thickness
	sl.fc *= lythick                // the amount of water in the soil profile at field capacity [mm]
	sl.wp *= lythick                // [mm]
	sl.tt = (sl.sat - sl.fc) / SOLK // pg.151 percolation time of travel; Ksat saturated hydraulic conductivity [mm/hr]
}

// New variable storage routing method channel constructor
// CHW: (W_bankfull) width of channel top at bank [m]
// CHD: (depth_bankfull) depth of water filled to bank [m]
// CHL: (L_ch) length of main channel [km]
// CHS: (slp_ch) length of main channel [-]
// CHN: (n) Manning's n value for the main channel
func (c *Channel) New(CHW, CHD, CHL, CHS, CHN float64) {
	c.zch = zch
	c.len = CHL
	c.sqlp = math.Sqrt(CHS) // square root of channel slope fraction (rise/run)
	c.wbf = CHW             // channel width at bankfull [m]
	c.n = CHN               // channel Mannings roughness
	c.wfld = 5. * CHW
	c.dbf = CHD

	c.wbtm = CHW - 2.*zch*CHD
	if c.wbtm <= 0. {
		c.wbtm = CHW / 2.
		c.zch = (CHW - c.wbtm) / 2. / CHD
	}

	c.d = CHD                                           // initial conditions
	c.vstr = 1000. * c.len * (c.wbtm + c.zch*CHD) * CHD // initial conditions
	c.zch2 = math.Sqrt(1. + c.zch*c.zch)
	c.zfld2 = math.Sqrt(1. + zfld*zfld)
}
