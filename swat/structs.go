package swat

/*
go implementaiton of the SWAT model
ref: Neitsch, S.L., J.G. Arnold, J.R., Kiniry, J.R. Williams, 2011. Soil and Water Assessment Tool: Theoretical Documentation Version 2009 (September 2011). 647pp.
*/

// WaterShed holds a collection of subbasins that make up a watershed
type WaterShed = map[int]*SubBasin

// SubBasin SWAT subbasin
type SubBasin struct {
	hru                              []*HRU   // hydrologic response unit (state variable)
	chn                              *Channel // channel unit (state variable)
	ca, dgw, aqt, agw, surlag, tconc float64  // parameters
	aq, wrch, qbf, qstr              float64  // state variables
	Outflow                          int      // SubBasin id outflow from this SubBasin (<0: farfield outflow)
}

// HRU SWAT hydrologic response unit
type HRU struct {
	sz          []SoilLayer // soil zone layers (state variable)
	cn          SCSCN
	f, ovn, slp float64 // strucutral
	cov         float64 // parameters
	iwt         bool    // flags
}

// SoilLayer is a soil layer unit use in SWAT
type SoilLayer struct {
	sat, fc, wp, tt float64 // parameters
	sw              float64 // state variables
	frz             bool
}

// Channel is a channel units in SWAT
// ref: Neitsch, S.L., J.G. Arnold, J.R., Kiniry, J.R. Williams, 2011. Soil and Water Assessment Tool: Theoretical Documentation Version 2009 (September 2011). 647pp.
type Channel struct {
	len, sqlp, dbf, wbf, wbtm, wfld, zch float64 // geometry
	n, zch2, zfld2                       float64 // parameter
	vstr, d                              float64 // state variable (vstr: is the change in volume of storage during the time step m³)
}
