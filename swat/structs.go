package swat

/*
go implementaiton of the SWAT model
ref: Neitsch, S.L., J.G. Arnold, J.R., Kiniry, J.R. Williams, 2011. Soil and Water Assessment Tool: Theoretical Documentation Version 2009 (September 2011). 647pp.
*/

// SubBasin SWAT subbasin
type SubBasin struct {
	hru               []*HRU   // hydrologic response unit (state variable)
	chn               *Channel // channel unit (state variable)
	ca, dgw, aqt, agw float64  // parameters
	aq, wrch, qbf     float64  // state variables
}

// HRU SWAT hydrologic response unit
type HRU struct {
	sz                 []SoilLayer // soil zone layers (state variable)
	cn                 SCSCN
	f, ovn, slp        float64 // strucutral
	surlag, tconc, cov float64 // parameters
	qstr               float64 // state variables
	iwt                bool    // flags
}

// SoilLayer is a soil layer unit use in SWAT
type SoilLayer struct {
	sat, fc, wp, tt float64 // parameters
	sw              float64 // state variables
	frz             bool
}
