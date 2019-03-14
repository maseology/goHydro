package profile

import "math"

const (
	carea = 1.   // area of profile [m²]
	ztop  = 0.   // elevation of top of profile [m]
	nsl   = 50   // number of profile  sublayers
	geom  = true // use geometric layering
)

// State holds the dynamic state for a profile, that can be subdivided into multiple layers for numerical applications
type State struct {
	PM              map[int]*rPM    // material properties
	t, tl, q, ql, p map[int]float64 // state variables t (theta) soil moisture content; q specific humdity (gas-filled pore space moiture content); p (psi) matric potential
	dz, cz, vol, K  map[int]float64 // structure
}

// WaterContentProfile returns the State's water content with depth
func (ps *State) WaterContentProfile() (t, z []float64) {
	t, z = make([]float64, nsl), make([]float64, nsl)
	for i := 1; i <= nsl; i++ {
		z[i-1] = ps.cz[i]
		t[i-1] = ps.t[i]
	}
	return
}

// Initialize state
func (ps *State) Initialize(p Profile, initSe float64, cellCenter bool) {

	// set dimensions
	z := ps.buildSubLayers(p.D[len(p.D)], geom)
	ps.vol, ps.dz = make(map[int]float64, nsl+1), make(map[int]float64, nsl+1)
	ps.cz = make(map[int]float64, nsl+1)
	ps.vol[0] = 0.
	for i := 0; i <= nsl; i++ {
		ps.dz[i] = z[i+1] - z[i]
		if i > 0 {
			ps.vol[i] = carea * ps.dz[i]
		}
	}
	for i := 1; i <= nsl+1; i++ {
		ps.cz[i] = z[i] + ps.dz[i]/2. // cell center (as depth from top), adding ghost cell below model for boundary conditions
	}
	if cellCenter {
		// adjust cell centered finite volume nodal distances at boundaries
		for i := 0; i <= nsl; i++ {
			ps.dz[i] = ps.cz[i+1] - ps.cz[i]
		}
	}

	// inital conditions
	ps.PM, ps.p = make(map[int]*rPM, nsl+2), make(map[int]float64, nsl+2)
	ps.t, ps.tl, ps.q, ps.ql, ps.K = make(map[int]float64, nsl+1), make(map[int]float64, nsl+1), make(map[int]float64, nsl+1), make(map[int]float64, nsl+1), make(map[int]float64, nsl+1)
	ps.p[0] = 0.
	ps.PM[0] = newPM(p.GetPorousMedium(0.))
	for i := 1; i <= nsl+1; i++ {
		pm := newPM(p.GetPorousMedium(ps.cz[i]))
		ps.PM[i] = pm
		ps.t[i] = pm.GetThetaSe(initSe)
		ps.tl[i] = ps.t[i]
		ps.p[i] = pm.GetPsi(ps.t[i])
		ps.K[i] = pm.GetK(ps.t[i])
		ps.q[i] = qp * math.Exp(mw*ps.p[i]/r/ts) // ps.PM[i].GetSpecificHumidity(ps.p[i])
		ps.ql[i] = ps.q[i]
		// ps.h[i] = ps.p[i] - ps.cz[i]*g // could set ps.h[nsl+1] for bottom constant head bc (not used in Newton Raphson)
	}

	// reconfigure cz to elevation
	for i := 1; i <= nsl+1; i++ {
		ps.cz[i] = ztop - ps.cz[i] // cell center
	}
}

func (ps *State) buildSubLayers(depth float64, geom bool) map[int]float64 {
	z := make(map[int]float64, nsl+2)
	z[0] = 0.0 // ghost cell
	z[1] = 0.0 // top of profile

	if geom { // geometric distribution
		sum := 0.0
		for i := 0; i <= nsl; i++ {
			sum += float64(i * i)
		}
		dz := depth / sum
		for i := 1; i <= nsl; i++ {
			z[i+1] = z[i] + dz*float64(i*i)
		}
	} else { // linear distribution
		dz := depth / float64(nsl)
		for i := 1; i <= nsl; i++ {
			z[i+1] = z[i] + dz
		}
	}
	return z
}

func (ps *State) reset() {
	for i := range ps.t {
		ps.t[i] = ps.tl[i]
		ps.p[i] = ps.PM[i].GetPsi(ps.t[i])
		ps.q[i] = ps.ql[i]
	}
}

func (ps *State) save() {
	for i := range ps.t {
		ps.tl[i] = ps.t[i]
		ps.ql[i] = ps.q[i]
	}
}

func (ps *State) setToInfiltrating() {
	ps.p[0] = math.Min(ps.PM[1].He, ps.PM[0].He)
	ps.t[0] = ps.PM[0].GetTheta(ps.p[0])
	ps.p[1] = ps.p[0]
	ps.t[1] = ps.t[0]
	ps.tl[1] = ps.t[0]
	ps.q[1] = qp * math.Exp(mw*ps.p[1]/r/ts) // ps.PM[1].GetSpecificHumidity(ps.p[1])
}

func (ps *State) setToFreeDraining() {
	ps.p[nsl+1] = ps.p[nsl]
	// ps.h[nsl+1] = ps.p[nsl+1] - ps.cz[nsl+1]*g
	ps.t[nsl+1] = ps.t[nsl]
	ps.K[nsl+1] = ps.K[nsl]
	ps.q[nsl+1] = ps.q[nsl]
}
