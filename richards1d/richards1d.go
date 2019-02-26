package richards1d

// Bittelli, M., Campbell, G.S., and Tomei, F., 2015. Soil Physics with Python. Oxford University Press.
// Richards, L.A., 1931. Capillary conduction of liquids through porous media. Physics 1: 318-333.

import (
	"github.com/maseology/goHydro/porousmedia"
	. "github.com/maseology/goHydro/profile"
)

// Maxter: maximum iterations;
// NsubLay: number of profile sublayers
const (
	waterDensity = 1000.0
	area         = 1.0
	MaxIter      = 100
	tolerance    = 1e-6
	NsubLay      = 50
	g            = 9.8065
	geomSubLay   = true
)

// RPM is an alias for PorousMedium needed
// to add methods to the root struct for
// the Richards 1D solver.
type RPM struct {
	*porousmedia.PorousMedium
}

// ProfileState holds the dynamic state for a profile
type ProfileState struct {
	z, v, dz, cz, T, Tl, Psi, H, K map[int]float64
	b3, mfpHe                      map[int]float64
	PM                             map[int]RPM
}

// buildSubLayers discretizes the profile into many
// fintie volume cells.
func (t *ProfileState) buildSubLayers(depth float64, geom bool) {
	t.z = make(map[int]float64)
	t.z[0] = 0.0 // ghost cell
	t.z[1] = 0.0 // top of profile

	if geom { // geometric distribution
		sum := 0.0
		for i := 0; i <= NsubLay; i++ {
			sum += float64(i * i)
		}
		dz := depth / sum
		for i := 1; i <= NsubLay; i++ {
			t.z[i+1] = t.z[i] + dz*float64(i*i)
		}
	} else { // linear distribution
		dz := depth / float64(NsubLay)
		for i := 1; i <= NsubLay; i++ {
			t.z[i+1] = t.z[i] + dz
		}
	}
}

// InitializeWater used to initialize profile state
func (t *ProfileState) InitializeWater(p Profile, se float64, solver int) {
	t.buildSubLayers(p.D[len(p.D)], geomSubLay)
	t.v, t.dz, t.cz = make(map[int]float64), make(map[int]float64), make(map[int]float64)
	t.v[0] = 0.0
	for i := 0; i <= NsubLay; i++ {
		t.dz[i] = t.z[i+1] - t.z[i]
		if i > 0 {
			t.v[i] = area * t.dz[i]
		}
	}
	for i := 1; i <= NsubLay+1; i++ {
		t.cz[i] = t.z[i] + t.dz[i]*0.5 // cell center (as depth from top), adding ghost cell below model for boundary conditions
	}

	if solver == 1 {
		// adjust cell centered finite volume nodal distances at boundaries
		for i := 0; i <= NsubLay; i++ {
			t.dz[i] = t.cz[i+1] - t.cz[i]
		}
	}

	// inital conditions
	t.PM, t.T, t.Tl, t.Psi, t.H, t.K, t.b3, t.mfpHe = make(map[int]RPM), make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64)
	t.Psi[0] = 0.
	t.PM[0] = RPM{p.GetPorousMedium(0.)}
	t.b3[0] = (2.*t.PM[0].B + 3.) / (t.PM[0].B + 3.)
	t.mfpHe[0] = t.PM[0].mfpHe()
	for i := 1; i <= NsubLay+1; i++ {
		pm := RPM{p.GetPorousMedium(t.cz[i])}
		t.PM[i] = pm
		t.T[i] = pm.GetThetaSe(se)
		t.Tl[i] = t.T[i]
		if solver == 3 { // matric flux potential linearization
			t.Psi[i] = pm.MFPfromTheta(t.T[i])
			t.K[i] = pm.hydraulicConductivityFromMFP(t.Psi[i])
		} else {
			t.Psi[i] = pm.GetPsi(t.T[i])
			t.K[i] = pm.GetK(t.T[i])
		}
		t.H[i] = t.Psi[i] - t.cz[i]*g // could set t.H[NsubLay+1] for bottom constant head bc
		t.b3[i] = (2.*pm.B + 3.) / (pm.B + 3.)
		t.mfpHe[i] = pm.mfpHe()
	}
}
