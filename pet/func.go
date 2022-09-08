package pet

import (
	"fmt"
	"log"
	"math"
)

const (
	ca  = 1.004 // J/g/K specific heat capacity of air
	mwr = .622  // ratio molecular weight of water vapour: dry air (Mw:Md)
)

///////////////
// see: 'slope saturation curve and psychrometric constant.xlsx' where equations have been confirmed
///////////////

// saturationVapourPressure [Pa]
func saturationVapourPressure(tC float64) float64 { // [Pa]
	// August-Roche-Magnus approximation (from pg.38 of Lu, N. and J.W. Godt, 2013. Hillslope Hydrology and Stability. Cambridge University Press. 437pp.)
	// for -30°C =< T =< 50°C
	return 610.49 * math.Exp(17.625*tC/(tC+243.04)) // [Pa=N/m²] R²=1
}

// slopeSaturationCurve [Pa/°C]
func slopeSaturationCurve(tC float64) float64 { // \Delta [Pa/°C]
	// based on Tetens formula (from pg.)
	// ref: Tetens, O., 1930. Uber einige meteorologische Begriffe. z. Geophys. 6:297-309.
	// also ref: Murray, F.W. 1967. On the computation of saturation vapor pressure. J. Appl. Meteor. 6: 203-204.
	return 4098. * saturationVapourPressure(tC) / math.Pow(tC+237.3, 2.) // Pa/°C
}

// latenHeatVapouration [MJ/kg]
func latenHeatVapouration(tC float64) float64 { // \lambda [MJ/kg]
	// pg.17 DeWalle, D.R. and A. Rango, 2008. Principles of Snow Hydrology. Cambridge University Press, Cambridge. 410pp.
	// for -50°C =< T =< 40°C
	return 3.e-6*tC*tC - .0025*tC + 2.4999 // MJ/kg {1 MJ/kg = 1000 J/g}
}

// psychrometricConstant [Pa/°C]
func psychrometricConstant(tC, Pa float64) float64 { // \gamma [Pa/°C]
	// The psychrometric constant relates the partial pressure of water in air to the air temperature.
	if Pa <= 10000 {
		log.Fatalf("ERROR [psychrometricConstant] pressure value looks suspect: %.0f Pa", Pa)
	}
	return ca / 1000. * Pa / latenHeatVapouration(tC) / mwr // Pa/°C
}

///////////////
// see: 'water.xlsx' where equations have been confirmed
///////////////

// densityLiquidWater [kg/m³]
func densityLiquidWater(tC float64) float64 { // [kg/m³]
	// see 'Physical constants.xlsx' with data from the Handbook of Chemistry and Physics, 76th Edition
	if tC <= 0. {
		return 999.84
	}
	// for 0°C =< T =< 60°C
	return 4.e-5*tC*tC*tC - 0.008*tC*tC + 0.0613*tC + 999.85 // kg/m³
}

// densityDryAir [g/m³]
func densityDryAir(tC float64) float64 {
	// see 'Physical constants.xlsx' with data from Appendix A of DeWalle and Rango, 2008 'for -40°C =< T =< 30°C
	// if tC < -40. || tC > 30. {
	// 	print("blah")
	// }
	return 0.0143*tC*tC - 4.8357*tC + 1292.2 // g/m³ R²=0.9993
}

// densitySaturatedAir [g/m³]
func densitySaturatedAir(tC float64) float64 {
	// see 'Physical constants.xlsx' with data from Appendix A of DeWalle and Rango, 2008 'for -40°C =< T =< 30°C
	// if tC<-40. || tC > 30. {
	// 	print("blah")
	// }
	return 0.0135*tC*tC - 5.0047*tC + 1287.8 // g/m³ R²=0.9994
}

// densityMoistAir [g/m³]
func densityMoistAir(tC, rh float64) float64 {
	if tC < -40. || tC > 30. {
		fmt.Println("Warning [DensityMoistAir] temperature out of range: ", tC)
	}
	if rh > 1. || rh < 0. {
		log.Fatalln("ERROR [DensityMoistAir] relative humidity out of range [0,1]: ", rh)
	}
	return (1.-rh)*densityDryAir(tC) + rh*densitySaturatedAir(tC)
}

// vapourPressureDeficit [Pa]
func vapourPressureDeficit(tC, rh float64) float64 {
	if rh > 1. || rh < 0. {
		log.Fatalln("ERROR [vapourPressureDeficit] relative humidity out of range [0,1]: ", rh)
	}
	return (1. - rh) * saturationVapourPressure(tC)
}
