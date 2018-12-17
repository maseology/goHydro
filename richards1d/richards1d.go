package richards1d

// Bittelli, M., Campbell, G.S., and Tomei, F., 2015. Soil Physics with Python. Oxford University Press.
// Richards, L.A., 1931. Capillary conduction of liquids through porous media. Physics 1: 318-333.

import (
	"math"

	"github.com/maseology/goHydro/porousmedia"
	. "github.com/maseology/goHydro/profile"
	"github.com/maseology/mmaths"
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
	t.Psi[0] = 0.0
	t.PM[0] = RPM{p.GetPorousMedium(0.0)}
	t.b3[0] = (2.0*t.PM[0].B + 3.0) / (t.PM[0].B + 3.0)
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
		t.b3[i] = (2.0*pm.B + 3.0) / (pm.B + 3.0)
		t.mfpHe[i] = pm.mfpHe()
	}
}

// CellCentFiniteVolWater solver
func (t *ProfileState) CellCentFiniteVolWater(dt, ubPotential float64, isFreeDrainage bool) (bool, int, float64) {
	// apply upper boundary condition
	t.Psi[0] = math.Min(ubPotential, t.PM[1].He)
	t.T[0] = t.PM[1].GetTheta(t.Psi[0])
	t.T[1] = t.T[0]
	t.Psi[1] = t.Psi[0]

	if isFreeDrainage {
		t.Psi[NsubLay+1] = t.Psi[NsubLay]
		t.H[NsubLay+1] = t.Psi[NsubLay+1] - t.cz[NsubLay+1]*g
		t.T[NsubLay+1] = t.T[NsubLay]
		t.K[NsubLay+1] = t.K[NsubLay]
	}

	h0, cp, f := make(map[int]float64), make(map[int]float64), make(map[int]float64)
	a, b, c, d := make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64)

	sum0 := 0.0
	for i := 1; i <= NsubLay; i++ {
		h0[i] = t.Psi[i] - t.cz[i]*g
		t.H[i] = h0[i]
		sum0 += waterDensity * t.v[i] * t.T[i]
	}

	massBalance := sum0
	nIter := 0
	for massBalance > tolerance && nIter < MaxIter {
		for i := 1; i <= NsubLay; i++ {
			t.K[i] = t.PM[i].GetK(t.T[i])
			cap := t.PM[i].dThetadH(h0[i], t.H[i], t.cz[i])
			cp[i] = (waterDensity * t.v[i] * cap) / dt
		}

		f[0] = 0
		for i := 1; i <= NsubLay; i++ {
			f[i] = area * meanK(t.K[i], t.K[i+1]) / t.dz[i]
		}

		for i := 1; i <= NsubLay; i++ {
			a[i] = -f[i-1]
			if i == 1 {
				b[i] = 1.0
				c[i] = 0.0
				d[i] = h0[i]
			} else if i < NsubLay {
				b[i] = cp[i] + f[i-1] + f[i]
				c[i] = -f[i]
				d[i] = cp[i] * h0[i]
			} else {
				b[NsubLay] = cp[NsubLay] + f[NsubLay-1]
				c[NsubLay] = 0.0
				if isFreeDrainage {
					d[NsubLay] = cp[NsubLay]*h0[NsubLay] - area*t.K[NsubLay]*g
				} else {
					d[NsubLay] = cp[NsubLay]*h0[NsubLay] - f[NsubLay]*(t.H[NsubLay]-t.H[NsubLay+1])
				}
			}
		}

		mmaths.ThomasBoundaryCondition(a, b, c, d, t.H, 1, NsubLay)

		newSum := 0.0
		for i := 1; i <= NsubLay; i++ {
			t.Psi[i] = t.H[i] + g*t.cz[i]
			t.T[i] = t.PM[i].GetTheta(t.Psi[i])
			newSum += waterDensity * t.v[i] * t.T[i]
		}

		if isFreeDrainage {
			t.Psi[NsubLay+1] = t.Psi[NsubLay]
			t.T[NsubLay+1] = t.T[NsubLay]
			t.K[NsubLay+1] = t.K[NsubLay]
			massBalance = math.Abs(newSum - (sum0 + f[1]*(t.H[1]-t.H[2])*dt - area*t.K[NsubLay]*g*dt))
		} else {
			massBalance = math.Abs(newSum - (sum0 + f[1]*(t.H[1]-t.H[2])*dt - f[NsubLay]*(t.H[NsubLay]-t.H[NsubLay+1])*dt))
		}
		nIter++
	}

	if massBalance < tolerance {
		return true, nIter, f[1] * (t.H[1] - t.H[2])
	}
	return false, nIter, 0.0
}

// NewtonRapsonMFP Matric Flux Potential method
func (t *ProfileState) NewtonRapsonMFP(dt, ubPotential float64, isFreeDrainage bool) (bool, int, float64) {
	// apply upper boundary condition
	t.Psi[1] = t.PM[0].mfpFromPsi(math.Min(ubPotential, t.PM[0].He))
	t.Tl[1] = t.PM[0].thetaFromMFP(t.Psi[1])
	t.T[1] = t.Tl[1]
	t.K[1] = t.PM[0].hydraulicConductivityFromMFP(t.Psi[1])
	t.Psi[0] = t.Psi[1]
	t.K[0] = 0.0

	if isFreeDrainage {
		t.Psi[NsubLay+1] = t.Psi[NsubLay]
		t.T[NsubLay+1] = t.T[NsubLay]
		t.K[NsubLay+1] = t.K[NsubLay]
	}

	u, cp, f := make(map[int]float64), make(map[int]float64), make(map[int]float64)
	a, b, c, d, dpsi := make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64)

	nIter := 0
	massBalance := 1.0
	for massBalance > tolerance && nIter < MaxIter {
		massBalance = 0.0
		for i := 1; i <= NsubLay; i++ {
			t.K[i] = t.PM[i].hydraulicConductivityFromMFP(t.Psi[i])
			cap := t.T[i] / ((t.PM[i].B + 3.0) * t.Psi[i])
			cp[i] = waterDensity * t.v[i] * cap / dt
			u[i] = g * t.K[i]
			f[i] = (t.Psi[i+1]-t.Psi[i])/t.dz[i] - u[i]
			b3 := (2.0*t.PM[i].B + 3.0) / (t.PM[i].B + 3.0)
			if i == 1 {
				a[i] = 0.0
				c[i] = 0.0
				b[i] = 1.0/t.dz[i] + cp[i] + g*b3*t.K[i]/t.Psi[i]
				d[i] = 0.0
			} else {
				b3m1 := (2.0*t.PM[i-1].B + 3.0) / (t.PM[i-1].B + 3.0)
				a[i] = -1.0/t.dz[i-1] - g*b3m1*t.K[i-1]/t.Psi[i-1]
				c[i] = -1.0 / t.dz[i]
				b[i] = 1.0/t.dz[i-1] + 1.0/t.dz[i] + cp[i] + g*b3*t.K[i]/t.Psi[i]
				d[i] = f[i-1] - f[i] + (waterDensity * t.v[i] * (t.T[i] - t.Tl[i]) / dt)
				massBalance += math.Abs(d[i])
			}
		}
		mmaths.ThomasBoundaryCondition(a, b, c, d, dpsi, 1, NsubLay)

		for i := 1; i <= NsubLay; i++ {
			t.Psi[i] -= dpsi[i]
			t.Psi[i] = math.Min(t.Psi[i], t.PM[i].mfpHe())
			t.T[i] = t.PM[i].thetaFromMFP(t.Psi[i])
		}

		if isFreeDrainage {
			t.Psi[NsubLay+1] = t.Psi[NsubLay]
			t.T[NsubLay+1] = t.T[NsubLay]
			t.K[NsubLay+1] = t.K[NsubLay]
		}
		nIter++
	}
	if massBalance < tolerance {
		return true, nIter, -f[1]
	}
	return false, nIter, 0.0
}

// returns specific moisture apacity pg.120
func (pm RPM) dThetadH(h0, h1, z float64) float64 {
	psi0 := h0 + g*z
	psi1 := h1 + g*z
	if math.Abs(psi1-psi0) < 1E-5 {
		return pm.dThetadPsi(psi0)
	}
	theta0 := pm.GetTheta(psi0)
	theta1 := pm.GetTheta(psi1)
	return (theta1 - theta0) / (psi1 - psi0)
}

func (pm RPM) dThetadPsi(psi float64) float64 {
	if psi > pm.He {
		return 0.0
	}
	return -pm.GetTheta(psi) / (pm.B * psi)
}

func (pm RPM) mfpHe() float64 {
	return pm.Ks * pm.He / (-3.0/pm.B - 1.0)
}

// MFPfromTheta is needed to determine the matrix
// flux potential from water content.
func (pm RPM) MFPfromTheta(theta float64) float64 {
	return pm.mfpHe() * math.Pow(theta/pm.Ts, pm.B+3.0)
}

func (pm RPM) mfpFromPsi(psi float64) float64 {
	return pm.mfpHe() * math.Pow(psi/pm.He, -3.0/pm.B-1.0)
}

func (pm RPM) thetaFromMFP(MFP float64) float64 {
	mfphe := pm.mfpHe()
	if MFP > mfphe {
		return pm.Ts
	}
	return pm.Ts * math.Pow(MFP/mfphe, 1.0/(pm.B+3.0))
}

func meanK(k1, k2 float64) float64 {
	if k1 != k2 {
		return (k1 - k2) / math.Log(k1/k2) // logarithmic mean
	}
	return k1
}

func (pm RPM) hydraulicConductivityFromMFP(MFP float64) float64 {
	return pm.Ks * math.Pow(MFP/pm.mfpHe(), (2.0*pm.B+3.0)/(pm.B+3.0))
}
