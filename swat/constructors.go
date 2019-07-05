package swat

import (
	"log"
	"math"
)

const (
	nsl         = 50     // number of soilzone layers
	lythick     = 10.    // layer thickness [mm]
	satini      = 1.     // initial degree of soil saturation relative to fc
	minslp      = 0.0001 // min CHS: channel slope
	secperday   = 86400.
	hoursperday = 24.
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
	6. initial conditions starting at 0.25 bankfull
	X. Notes:
		- Penman-Monteith is not used; therefore transpiration is not modelled (pg.135)
		- vertisols are not included, i.e., no bypass flow (pg.152)
		- no bypass flow; no partioning to deep aquifer (pg.173)
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
// SURLAG: surface runoff lag coefficient [0,15]
// GWDELAY: (delta_gw) delay time for aquifer recharge [days]
// GWQMN: (aq_shthr) threshold water level in aquifer for baseflow [mm]
// ALPHABF: (alpha_bf) baseflow recession coeficient (1/k)
func (b *SubBasin) New(HRUs []*HRU, Chn *Channel, SUBKM, SLSUBBSN, CHL, CHS, CHN, SURLAG, GWDELAY, ALPHABF float64) {
	b.Ca = SUBKM // subbasin contributing area [km²]
	b.surlag = SURLAG
	b.dgw = GWDELAY // the delay of soil zone percolation to aquifer [days]
	b.aqt = 0.      // GWQMN (no shallow aquifer baseflow threshold (GWQMIN/aqt))
	b.agw = ALPHABF // baseflow recession coefficient
	b.Outflow = -1  // SubBasin outflow ID
	b.slplen = SLSUBBSN
	b.tribl = CHL
	b.tribs = CHS
	b.tribn = CHN

	// build HRUs and tconc, add channel element
	b.chn = *Chn
	ftot, wslp, wovn := 0., 0., 0.
	b.hru = make([]HRU, len(HRUs))
	for i, u := range HRUs {
		ftot += u.f
		b.hru[i] = *u
		wslp += u.f * u.slp // subsasin weighted average slope
		wovn += u.f * u.ovn // subsasin weighted average overland roughness
	}
	if math.Abs(1.-ftot) > 0.001 {
		for _, u := range b.hru {
			u.f /= ftot // nomalize hru fractions
		}
	}
	b.tconc = tconc(wslp, SLSUBBSN, wovn, CHL, CHS, CHN, SUBKM)
	if math.IsNaN(b.tconc) {
		log.Fatalf("SubBasin.New error: tconc is NaN\n")
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
// CV: aboveground biomass and residue [kg/ha]
// ESCO: soil evporation compensation coefficient [0,1] (pg.138)
// IWATABLE: high water table code: set to true when seasonal high water table present
func (m *HRU) New(sz SoilLayer, HRUFR, HRUSLP, OVN, CN2, CV, ESCO float64, IWATABLE bool) {
	m.f = HRUFR
	m.slp = HRUSLP
	m.ovn = OVN
	m.cov = math.Exp(-5.0e-5 * CV) // soil cover index (pg.135)
	m.esco = ESCO
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
	sl.sat = n * lythick            // the amount of water in the soil profile when completely saturated [mm] based on layer thickness
	sl.fc *= lythick                // the amount of water in the soil profile at field capacity [mm]
	sl.sw = sl.fc * satini          // initial saturation
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
	c.d = CHD           // initial flow depth [m]
	c.len = CHL * 1000. // converting to [m]
	c.zch = zch
	c.sqslp = math.Sqrt(math.Max(minslp, CHS)) // square root of channel slope fraction (rise/run)
	c.wbf = CHW                                // channel width at bankfull [m]
	c.n = CHN                                  // channel Mannings roughness
	c.wfld = 5. * CHW
	c.dbf = CHD // bankful depth [m]
	c.zch2 = math.Sqrt(1. + c.zch*c.zch)
	c.zfld2 = math.Sqrt(1. + zfld*zfld)

	c.wbtm = CHW - 2.*zch*CHD // pg.429
	if c.wbtm <= 0. {
		c.wbtm = CHW / zch
		c.zch = (CHW - c.wbtm) / 2. / CHD
	}
	c.vstr = c.len * (c.wbtm + c.zch*c.d) * c.d // initial volume [m³]

	ach := c.vstr / c.len         // pg.432
	pch := c.wbtm + 2.*c.d*c.zch2 // pg.430
	rch := ach / pch              // hydraulic radius

	q := ach * math.Pow(rch, twothird) * c.sqslp / c.n // pg.431
	tt := c.vstr / q
	if tt < secperday/2. {
		tt = secperday / 2.
	}
	c.sc = 2. * secperday / (2.*tt + secperday) // pg.434
	if c.sc < 0. || c.sc > 1. || (math.IsNaN(c.sc) && c.len > 0.) {
		log.Fatalf("Channel.New error: SC = %f\n", c.sc)
	}
}
