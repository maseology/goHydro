package porousmedia

import "math"

// PorousMedium contains a set of parameters that
// describe the transport properties of porour media
// Currently only supporting Campbell (1974) parameterization
// ref: Campbell, G.S., 1974. A simple method for determining unsaturated conductivity from moisture retention data. Soil Science, 117: 311-387.
// b: shape parameter, He: air-entry potential [J/kg]
type PorousMedium struct {
	Ts, Tr, Ks, He, B float64
}

// // New default constructor for testing
// // Note: residual soil moisture is not in the Campbell model
// func (pm *PorousMedium) New() {
// 	// *pm = PorousMedium{
// 	// 	Ts: 0.44,
// 	// 	Tr: 0.01,
// 	// 	Ks: 0.001,
// 	// 	He: -2.08,
// 	// 	B:  4.74,
// 	// }
// 	*pm = PorousMedium{ //silt loam
// 		Ts: 0.43,
// 		Tr: 0.05,
// 		Ks: 0.003,
// 		He: -2.08,
// 		B:  4.74,
// 	}
// }

// GetK returns the hydraulic conductivity for a given
// volumetric water content (Campbell, 1974).
func (pm *PorousMedium) GetK(theta float64) float64 {
	if theta >= pm.Ts {
		return pm.Ks // saturated hydraulic conductivity
	}
	return pm.Ks * math.Pow(theta/pm.Ts, 2.0*pm.B+3.0)
}

// GetKfromPsi returns the hydraulic conductivity for a given
// matric potential (Campbell, 1974).
func (pm *PorousMedium) GetKfromPsi(psi float64) float64 {
	if psi >= pm.He {
		return pm.Ks // saturated hydraulic conductivity
	}
	return pm.Ks * math.Pow(pm.He/psi, 2.0+3.0/pm.B)
}

// GetPsi returns the matric potential for a given
// volumetric water content (Campbell, 1974).
func (pm *PorousMedium) GetPsi(theta float64) float64 {
	if theta >= pm.Ts {
		return pm.He // air entry potential
	}
	return pm.He * math.Pow(theta/pm.Ts, -pm.B)
}

// GetTheta returns the volumetric water content for a
// given matric potential (Campbell, 1974).
func (pm *PorousMedium) GetTheta(psi float64) float64 {
	if psi >= pm.He {
		return pm.Ts // saturated volumetric water content
	}
	return pm.Ts * math.Pow(psi/pm.He, -1.0/pm.B)
}

// GetThetaSe returns the volumetric water content for a
// given degree of saturation Se=(t-tr)/(ts-tr)~t/ts
func (pm *PorousMedium) GetThetaSe(se float64) float64 {
	return se * pm.Ts //+ pm.tr*(1.0-se)
}

// GetSe returns the volumetric water content for a
// given volumetric water content. Se=t/ts
func (pm *PorousMedium) GetSe(theta float64) float64 {
	return theta / pm.Ts
}

// GetSePsi returns the volumetric water content for a
// given matric potential (Campbell, 1974). Se=t/ts
func (pm *PorousMedium) GetSePsi(psi float64) float64 {
	if psi >= 0.0 {
		return 1.0
	}
	if psi >= pm.He {
		return 1.0
	}
	return math.Pow(psi/pm.He, -1.0/pm.B)
}
