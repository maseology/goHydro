package solirrad

import (
	"fmt"
	. "math"
	"strconv"
	"time"
)

// The solar irradition package uses solar irradition to compute irradiance integrated over time
// From appendix B.1 of DeWalle, D.R. and A. Rango, 2008. Principles of Snow Hydrology. Cambridge University Press, Cambridge. 410pp.
// applicable for determining total potential solar radiation on a horizontal surface from sunrise to sunset (MJ/m²). Description begins on pg.397

const (
	solcnst           = 4.896             // 1360 W/m² = 4.896 MJ/hr/m² is the solar constant: "the average flux density of radiation received outside the Earth's atmosphere perpendicular to the solar beam" (DeWalle and Rango, 2008)
	angvel            = 0.261799387799149 // rad/hr: angular velocity = 2*PI/24hr
	earthAxialTilt    = 23.43689 * Pi / 180.
	earthEccentricity = 0.0167
)

// SolIrad a stuct used to compute potential solar irradiation given latitude, slope and aspect.
type SolIrad struct {
	Lat, Slope, Aspect               float64      // representative/average latitude of interested area in radians; slope in radians; aspect CW from north
	solardec, radivectsq             [366]float64 // Solar irradtion constants: solar declination (degrees); radius vector squared (DeWalle and Rango, 2006, pg.395).
	psih                             [366]float64 // Daily Potential Solar Irradiation on a horizontal surface in MJ/m² from sunrise to sunset
	zeffh                            [366]float64 // Average effective zenith angle per Julian day on a horizontal surface
	dayhoursh                        [366]float64 // time of sunrise to sunset in hours for a horizontal surface, per Julian day
	psi, dayhours, zeff, srh, ssh, f [366]float64
}

// New constructor
func New(LatitudeDeg, SlopeRad, AspectCwnRad float64) SolIrad {
	// southern Ontario Latitude_deg = 43.6
	var si SolIrad
	si.Lat = LatitudeDeg * Pi / 180.
	si.Slope = SlopeRad
	si.Aspect = AspectCwnRad
	si.buildSolarDeclination()
	si.buildRadiusVectorSquared()
	si.horizontalSurfaceCompute()
	si.slopingSurfaceCompute()
	return si
}

// PSIfactor returns the ratio of potential solar irradiation of a sloped surface to the horitontal
func (si *SolIrad) PSIfactor() [366]float64 {
	return si.f
}

func (si *SolIrad) buildSolarDeclination() {
	// based on Wikipedia page: Position of the Sun
	for i := 0; i <= 365; i++ {
		si.solardec[i] = Asin(Sin(-earthAxialTilt) * Cos(2.*Pi*float64(i+11)/365.24+2.*earthEccentricity*Sin(2.*Pi*float64(i-1)/365.24)))
	}
	//si.solardec[37] = Pi * -15.583 / 180. ///////////////////////////HARD-CODED FOR TESTING///////////////////////////////////////////////////
}

func (si *SolIrad) buildRadiusVectorSquared() {
	si.radivectsq = [...]float64{0.96853, 0.96863, 0.96872, 0.96881, 0.96891, 0.969, 0.9691, 0.96919, 0.96929, 0.96938, 0.96963, 0.96988, 0.97012, 0.97037, 0.97062, 0.97087, 0.97112, 0.97136, 0.97161, 0.97186, 0.97211, 0.97235, 0.9726, 0.97285, 0.97321, 0.97357, 0.97393, 0.97429, 0.97465, 0.97501, 0.97538, 0.97574, 0.9761, 0.97646, 0.97682, 0.97718, 0.97754, 0.9779, 0.97837, 0.97884, 0.97932, 0.97979, 0.98026, 0.98073, 0.98121, 0.98168, 0.98215, 0.98262, 0.9831, 0.98357, 0.98404, 0.98451, 0.98498, 0.98545, 0.98592, 0.98638, 0.98685, 0.98732, 0.98779, 0.98826, 0.98873, 0.9892, 0.98967, 0.99013, 0.9906, 0.99107, 0.99154, 0.99212, 0.99269, 0.99327, 0.99384, 0.99442, 0.99499, 0.99557, 0.99615, 0.99672, 0.9973, 0.99787, 0.99845, 0.99902, 0.9996, 1.00016, 1.00071, 1.00127, 1.00183, 1.00238, 1.00294, 1.0035, 1.00405, 1.00461, 1.00516, 1.00572, 1.00628, 1.00683, 1.00739, 1.00792, 1.00844, 1.00897, 1.00949, 1.01002, 1.01055, 1.01107, 1.0116, 1.01212, 1.01265, 1.01318, 1.0137, 1.01423, 1.01475, 1.01528, 1.01575, 1.01623, 1.0167, 1.01717, 1.01764, 1.01812, 1.01859, 1.01906, 1.01954, 1.02001, 1.02048, 1.02095, 1.02143, 1.0219, 1.02226, 1.02262, 1.02298, 1.02333, 1.02369, 1.02405, 1.02441, 1.02477, 1.02513, 1.02549, 1.02585, 1.0262, 1.02656, 1.02692, 1.02728, 1.02754, 1.0278, 1.02806, 1.02831, 1.02857, 1.02883, 1.02909, 1.02935, 1.02961, 1.02987, 1.03012, 1.03038, 1.03064, 1.0309, 1.031, 1.0311, 1.0312, 1.03129, 1.03139, 1.03149, 1.03159, 1.03169, 1.03179, 1.03189, 1.03198, 1.03208, 1.03218, 1.03228, 1.03238, 1.03248, 1.03258, 1.03267, 1.03277, 1.03287, 1.03297, 1.03287, 1.03276, 1.03266, 1.03256, 1.03245, 1.03235, 1.03225, 1.03214, 1.03204, 1.03194, 1.03183, 1.03173, 1.03162, 1.03152, 1.03142, 1.03131, 1.03121, 1.03111, 1.031, 1.0309, 1.03066, 1.03042, 1.03018, 1.02993, 1.02969, 1.02945, 1.02921, 1.02897, 1.02873, 1.02849, 1.02825, 1.028, 1.02776, 1.02752, 1.02728, 1.0269, 1.02651, 1.02613, 1.02574, 1.02536, 1.02497, 1.02459, 1.02421, 1.02382, 1.02344, 1.02305, 1.02267, 1.02228, 1.0219, 1.02146, 1.02102, 1.02058, 1.02013, 1.01969, 1.01925, 1.01881, 1.01837, 1.01793, 1.01749, 1.01705, 1.0166, 1.01616, 1.01572, 1.01528, 1.01475, 1.01423, 1.0137, 1.01318, 1.01265, 1.01212, 1.0116, 1.01107, 1.01055, 1.01002, 1.00949, 1.00897, 1.00844, 1.00792, 1.00739, 1.00683, 1.00628, 1.00572, 1.00516, 1.00461, 1.00405, 1.0035, 1.00294, 1.00238, 1.00183, 1.00127, 1.00071, 1.00016, 0.9996, 0.99906, 0.99853, 0.99799, 0.99745, 0.99691, 0.99638, 0.99584, 0.9953, 0.99476, 0.99423, 0.99369, 0.99315, 0.99261, 0.99208, 0.99154, 0.991, 0.99047, 0.98993, 0.9894, 0.98886, 0.98833, 0.98779, 0.98725, 0.98672, 0.98618, 0.98565, 0.98511, 0.98458, 0.98404, 0.9836, 0.98316, 0.98272, 0.98229, 0.98185, 0.98141, 0.98097, 0.98053, 0.98009, 0.97965, 0.97922, 0.97878, 0.97834, 0.9779, 0.97754, 0.97718, 0.97682, 0.97646, 0.9761, 0.97574, 0.97538, 0.97501, 0.97465, 0.97429, 0.97393, 0.97357, 0.97321, 0.97285, 0.9726, 0.97235, 0.97211, 0.97186, 0.97161, 0.97136, 0.97112, 0.97087, 0.97062, 0.97037, 0.97012, 0.96988, 0.96963, 0.96938, 0.96929, 0.96919, 0.9691, 0.969, 0.96891, 0.96881, 0.96872, 0.96863, 0.96853, 0.96844, 0.96834, 0.96825, 0.96816, 0.96806, 0.96797, 0.96787, 0.96778, 0.96768, 0.96759, 0.96768, 0.96778, 0.96787, 0.96797, 0.96806, 0.96816, 0.96825, 0.96834, 0.96844}
}

func (si *SolIrad) horizontalSurfaceCompute() {
	// for horizontal surfaces
	for i := 0; i <= 365; i++ {
		d := -Tan(si.Lat) * Tan(si.solardec[i]) // Eq B.3: cosine of the hour angle
		if d >= 1. {                            // no sun exposure
			si.dayhoursh[i] = 0.0
			si.psih[i] = 0.0
			si.zeffh[i] = -9999.0
		} else {
			if d < -1. { // when <= -1.0, sun exposure all day (pg. 397)
				d = -1.
			}
			d = Acos(d)
			t1 := d / angvel               // ±{time sunrise/time sunset}
			si.dayhoursh[i] = 2. * Abs(t1) // number of hours per julian day on a horizontal surface
			if si.dayhoursh[i] > 24. {
				si.dayhoursh[i] = 24.0
			}
			if si.dayhoursh[i] < 0. {
				si.dayhoursh[i] = 0.0
			}
			d = si.dayhoursh[i]*Sin(si.Lat)*Sin(si.solardec[i]) + Cos(si.Lat)*Cos(si.solardec[i])*(Sin(d)-Sin(-d))/angvel // = CosZ = SinA; Eq B.3
			si.psih[i] = solcnst / si.radivectsq[i] * d                                                                   // in MJ/m² from sunrise to sunset. Eq B.4
			si.zeffh[i] = Acos(d / si.dayhoursh[i])                                                                       // Average effective zenith angle on a horizontal surface for the Julian day
		}
	}
}

func (si *SolIrad) slopingSurfaceCompute() {
	// Solar Irradiation theory for sloping surfaces (see B.2.2 of DeWalle and Rango, 2008); AspectCWN_rad in Radians clockwise from north
	d := Sin(si.Slope)*Cos(si.Aspect)*Cos(si.Lat) + Cos(si.Slope)*Sin(si.Lat)
	lateq := Asin(d) // Eq B.6
	d = Cos(si.Slope)*Cos(si.Lat) - Cos(si.Aspect)*Sin(si.Slope)*Sin(si.Lat)
	a := Atan2(Sin(si.Aspect)*Sin(si.Slope), d) // Eq B.8 difference in longitude between equivalent horizontal surface and slope
	for i := 0; i <= 365; i++ {
		d = -Tan(lateq) * Tan(si.solardec[i]) // Eq B.10
		if d >= 1.0 {                         // no sun exposure
			si.dayhours[i] = 0.0
			si.psi[i] = 0.0
			si.zeff[i] = -9999.0
			si.srh[i] = -9999.0
			si.ssh[i] = -9999.0
			si.f[i] = -9999.0
		} else {
			if d < -1. { // when <= -1.0, sun exposure all day (pg. 397)
				d = -1.
			}
			d = Acos(d) // Hour angle at equivalent latitude
			t1 := (-d - a) / angvel
			t2 := (d - a) / angvel
			if t1 < -si.dayhoursh[i]/2. {
				t1 = -si.dayhoursh[i] / 2.
			}
			if t2 > si.dayhoursh[i]/2.0 {
				t2 = si.dayhoursh[i] / 2.
			}
			si.dayhours[i] = t2 - t1
			if si.dayhours[i] <= 0.0 { // no sun exposure
				si.dayhours[i] = 0.0
				si.psi[i] = 0.0
				si.zeff[i] = -9999.0
				si.srh[i] = -9999.0
				si.ssh[i] = -9999.0
				si.f[i] = -9999.0
			} else {
				d = si.dayhours[i]*Sin(lateq)*Sin(si.solardec[i]) + Cos(lateq)*Cos(si.solardec[i])*(Sin(angvel*t2+a)-Sin(angvel*t1+a))/angvel // Eq B.11
				si.psi[i] = solcnst / si.radivectsq[i] * d                                                                                    // MJ/m²
				d /= si.dayhours[i]
				if Abs(d) > 1. {
					panic("SolIrad error slopingSurfaceCompute")
				}
				si.zeff[i] = Acos(d)
				si.srh[i] = 12. + t1 // assuming solar noon is 12PM
				si.ssh[i] = 12. + t2
				si.f[i] = si.psi[i] / si.psih[i]
			}
		}
	}
}

// PSI potential solar irradiation for any instant in time [W/m2]
func (si *SolIrad) PSI(HourRelativeToNoon float64, DayOfYear int) float64 {
	d := Sin(si.Slope)*Cos(si.Aspect)*Cos(si.Lat) + Cos(si.Slope)*Sin(si.Lat)
	lateq := Asin(d) // Eq B.6
	a := Atan(Sin(si.Aspect) * Sin(si.Slope) / (Cos(si.Slope)*Cos(si.Lat) - Cos(si.Aspect)*Sin(si.Slope)*Sin(si.Lat)))
	wt := HourRelativeToNoon*angvel + a // Eq B.7
	return solcnst / .0036 / si.radivectsq[DayOfYear-1] * (Sin(lateq)*Sin(si.solardec[DayOfYear-1]) + Cos(lateq)*Cos(si.solardec[DayOfYear-1])*Cos(wt))
}

// PSIdaily potential solar irradiation for a given day of year [MJ/m2]
func (si *SolIrad) PSIdaily(DayOfYear int) float64 {
	return si.psi[DayOfYear-1]
}

// DaylightHours time of daylight (hr)
func (si *SolIrad) DaylightHours(DayOfYear int) float64 {
	return si.dayhours[DayOfYear-1]
}

// SunRiseSunSet time of sunrise and sunset assuming solar noon is 12PM
func (si *SolIrad) SunRiseSunSet(DayOfYear int) (float64, float64) {
	return si.srh[DayOfYear-1], si.ssh[DayOfYear-1]
}

// PrintSunRiseSunSet prints sunrise and sunset
func (si *SolIrad) PrintSunRiseSunSet(dt time.Time) {
	sr, ss := si.SunRiseSunSet(dt.YearDay())
	if tz, tzo := dt.Zone(); tz == "EDT" && tzo == -4*3600 { // adjust for DST in eastern timezone only
		sr++
		ss++
	}
	sr2, err := time.ParseDuration(strconv.FormatFloat(sr, 'f', 6, 64) + "h")
	if err != nil {
		fmt.Println(err)
	}
	ss2, err := time.ParseDuration(strconv.FormatFloat(ss, 'f', 6, 64) + "h")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(sr2, ss2)
}

// ApplyAtmosphericTransmissionCoefficient applies a correction factor to account for losses due to atmospheric scattering
func (si *SolIrad) ApplyAtmosphericTransmissionCoefficient(coef float64) {
	// see: Buffo, Fritschen, Murphy, 1972. USDA: Direct Solar Radiation on Various Slopes From 0 to 60 Degrees North Latitude
	// See also pg. 7 in Budyko: Climate and Life, eq. 1.16; pg. 345 in Oke: Boundary Layer Climates
	// coef = 0.6 smoggy air to 0.9 clean and clear, typical values 0.84 (see Oke and references therein)
	// Note: altitude angle (A) is complementary to zenith (i.e., A = 90-Z; or sin(A) = cos(Z))
	if si.Lat > 60.*Pi { // upper limit acording to Budyko - offers other methods (see below)
		panic("upper latitude limit for applying the Atmospheric Transmission Coefficient")
	}
	for i := 0; i <= 365; i++ {
		si.psi[i] *= Pow(coef, 1./Cos(si.zeff[i]))
		si.psih[i] *= Pow(coef, 1./Cos(si.zeffh[i]))
	}
}

// NetSWfromPotential is an empirical conversion from incomming potential radiantion to incomming SW radiation arriving at the surface
func (si *SolIrad) NetSWfromPotential(CloudCoverFraction float64, DayOfYear int) float64 {
	// see pg 151 in DeWalle & Rango; Also attributed to Linacre (1992), but the above form is preferred as sunshine hours are easiest to determine
	return (0.85 - 0.47*CloudCoverFraction) * si.psi[DayOfYear-1]
}
