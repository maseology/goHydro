package pet

import "math"

const (
	ca  = 1.004 // J/g/K specific heat capacity of air
	mwr = .622  // ratio molecular weight of water vapour: dry air (Mw:Md)
)

///////////////
// see: 'slope saturation curve and psychrometric constant.xlsx' where equations have been confirmed
///////////////

// saturationVapourPressure [Pa]
func saturationVapourPressure(tempC float64) float64 { // [Pa]
	// August-Roche-Magnus approximation (from pg.38 of Lu, N. and J.W. Godt, 2013. Hillslope Hydrology and Stability. Cambridge Universtiy Press. 437pp.)
	// for -30°C =< T =< 50°C
	return 610.49 * math.Exp(17.625*tempC/(tempC+243.04)) // [Pa=N/m²] R²=1
}

// slopeSaturationCurve [Pa/°C]
func slopeSaturationCurve(TmeanC float64) float64 { // \Delta [Pa/°C]
	// based on Tetens formula (from pg.)
	// ref: Tetens, O., 1930. Uber einige meteorologische Begriffe. z. Geophys. 6:297-309.
	// also ref: Murray, F.W. 1967. On the computation of saturation vapor pressure. J. Appl. Meteor. 6: 203-204.
	return 4098. * saturationVapourPressure(TmeanC) / math.Pow(TmeanC+237.3, 2.) // Pa/°C
}

// latenHeatVapouration [MJ/kg]
func latenHeatVapouration(tempC float64) float64 { // \lambda [MJ/kg]
	// pg.17 DeWalle, D.R. and A. Rango, 2008. Principles of Snow Hydrology. Cambridge University Press, Cambridge. 410pp.
	// for -50°C =< T =< 40°C
	return 3.e-6*tempC*tempC - .0025*tempC + 2.4999 // MJ/kg {1 MJ/kg = 1000 J/g}
}

// psychrometricConstant [Pa/°C]
func psychrometricConstant(tempC, presPa float64) float64 { // \gamma [Pa/°C]
	// The psychrometric constant relates the partial pressure of water in air to the air temperature.
	return ca / 1000. * presPa / latenHeatVapouration(tempC) / mwr // Pa/°C
}

///////////////
// see: 'water.xlsx' where equations have been confirmed
///////////////

// densityLiquidWater [kg/m³]
func densityLiquidWater(tempC float64) float64 { // [kg/m³]
	// see 'Physical constants.xlsx' with data from the Handbook of Chemistry and Physics, 76th Edition
	if tempC <= 0. {
		return 999.84
	}
	// for 0°C =< T =< 60°C
	return 4.e-5*tempC*tempC*tempC - 0.008*tempC*tempC + 0.0613*tempC + 999.85 // kg/m³
}
