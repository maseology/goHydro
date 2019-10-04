package snowpack

import (
	"log"
	"math"
)

// ref: DeWalle, D.R. and A. Rango, 2008. Principles of Snow Hydrology. Cambridge University Press, Cambridge. 410pp.

const (
	// physical constants
	pw = 999.84   // [kg/m³] Density of liquid water at 0°C
	pi = 917.     // [kg/m³] Density of ice at 0°C = 917  (pg. 54)
	lf = 334000.  // [J/kg] latent heat of fusion at freezing point
	lv = 2496000. // [J/kg] at 0°C latent heat of vaporation = 2496.0 J/g
	ci = 2100.    // [J/kg/K] specific heat capacity of ice
	cw = 4187.6   // [J/kg/K] specific heat of liquid water = 4.1876E3

	// parameters
	denmin = 25.  // [kg/m³] minimum snowfall density
	den0   = 350. // [kg/m³] density of falling ripe snow (at or above temperatures of 0°C)
	densf  = 100. // [kg/m³] (average) density of falling snow; can range from 50-350 kg/m³ (see pg. 55)
	swi    = 0.05 // irreducible liquid saturation, volume of liquid per volume of pore-space

	// other
	df = 1. // [ts/day] day factor (adjust when daily timesteps are not used)
)

type snowpack struct {
	swe, den, ts, tb, lwc float64
}

func (s *snowpack) properties() (porosity, depth float64) {
	// rearranging eq. 3.1 pg. 54 in DeWalle, D.R. and A. Rango, 2008. Principles of Snow Hydrology. Cambridge University Press, Cambridge. 410pp.
	if s.den == 0. || s.swe == 0. {
		return
	}
	depth = s.swe * pw / s.den
	tw := s.lwc / depth
	porosity = 1. - (s.den - pw*tw/pi)
	return
}

func (s *snowpack) addToPack(sweFall, denFall float64) {
	if sweFall > 0. {
		s.den = (s.swe*s.den + sweFall*denFall) / (s.swe + sweFall)
		s.swe += sweFall
		if s.den < denmin || s.den > pw {
			log.Fatalf("snowpack.addToPack error: snowpack density out of physical range")
		}
	} else {
		log.Fatalf("snowpack.addToPack error: negative swe being added")
	}
}

func (s *snowpack) drainFromPack() (drainage float64) {
	if s.lwc > 0. {
		if s.lwc == s.swe {
			drainage = s.swe
			s.swe = 0.
			s.lwc = 0.
			s.ts = 0.
			s.den = 0.
		} else {
			por, depth := s.properties()
			lwrc := por * swi * depth // snowpack liquid water retention capacity
			def := lwrc - s.lwc       // deficit
			if def < 0. {             // excess water
				drainage = -def
				pfroz := (s.swe*s.den - s.lwc*pw) / (s.swe - s.lwc)
				s.den = ((s.lwc-drainage)*pw + (s.swe-s.lwc)*pfroz) / (s.swe - drainage)
				s.lwc = lwrc
				s.swe += def
			}
		}
	}
	return
}

func (s *snowpack) updateSurfaceTemperature(t float64) { // pg.279
	const tsf = .5 // TSF (surface temperature factor), 0.1-0.5 have been used
	if s.swe > 0. {
		s.ts += tsf * df * (t - s.ts)
		if s.ts > 0. {
			s.ts = 0.
		}
	} else {
		s.ts = 0.
	}
}

func (s *snowpack) internalFreeze(sweAffected float64) {
	// internal state change (set sweAffected < 0.0 for internal melting)
	if sweAffected > 0. { // internal freezing
		s.den += sweAffected / s.swe * (pi - pw)
		if s.den < 0. {
			log.Fatalf("snowpack.internalFreeze error: density less than zero")
		}
	} else if sweAffected > 0. { // internal melting
		if s.swe == s.lwc-sweAffected {
			s.den = pw
		} else {
			pfroz := (s.swe*s.den - s.lwc*pw) / (s.swe - s.lwc)
			s.den = (s.lwc-sweAffected)/s.swe*(pw-pfroz) + pfroz
			if s.den <= 0. {
				log.Fatalf("snowpack.internalFreeze error: density less than zero")
			}
		}
	}
}

func (s *snowpack) densify() {
	const denscoef = 1. // coefficient to the densification factor
	if s.den > 0. {
		if s.den < pi {
			f := math.Pow(pi/s.frozenPackDensity(), df*denscoef)
			if f > 1. {
				if s.den*f > pi {
					s.den = pi
				} else {
					s.den *= f
				}
			}
		}
	}
}

func (s *snowpack) frozenPackDensity() float64 {
	return (s.swe*s.den - s.lwc*pw) / (s.swe - s.lwc) // density of frozen snowpack
}
