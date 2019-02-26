package richards1d

// Bittelli, M., Campbell, G.S., and Tomei, F., 2015. Soil Physics with Python. Oxford University Press.
// Richards, L.A., 1931. Capillary conduction of liquids through porous media. Physics 1: 318-333.

import (
	"math"

	"github.com/maseology/mmaths"
)

// CellCentFiniteVolWater is a cell-centered finite volume solution
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

// NewtonRapsonMP is the Matrix Potential solution with a Newton Rapson solver
func (t *ProfileState) NewtonRapsonMP(dt, ubPotential float64, isFreeDrainage bool) (bool, int, float64) {
	// apply upper boundary condition
	t.Psi[1] = math.Min(ubPotential, t.PM[0].He)
	t.Tl[1] = t.PM[0].GetTheta(t.Psi[1])
	t.T[1] = t.Tl[1]

	if isFreeDrainage {
		t.Psi[NsubLay+1] = t.Psi[NsubLay]
		t.T[NsubLay+1] = t.T[NsubLay]
		t.K[NsubLay+1] = t.K[NsubLay]
	}

	u, cp, f, du := make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64)
	a, b, c, d, dpsi, cn := make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64)
	for i := 1; i <= NsubLay; i++ {
		cn[i] = 2. + 3./t.PM[i].B // Campbell (1974) shape parameter
	}

	nIter := 0
	massBalance := 1.
	for massBalance > tolerance && nIter < MaxIter {
		massBalance = 0.
		for i := 1; i <= NsubLay; i++ {
			t.K[i] = t.PM[i].GetK(t.T[i])
			u[i] = g * t.K[i]
			du[i] = -u[i] * cn[i] / t.Psi[i]
			cap := t.PM[i].dThetadPsi(t.Psi[i])
			cp[i] = (waterDensity * t.v[i] * cap) / dt
		}

		for i := 1; i <= NsubLay; i++ {
			f[i] = ((t.Psi[i+1]*t.K[i+1] - t.Psi[i]*t.K[i]) / (t.dz[i] * (1. - cn[i]))) - u[i]
			if i == 1 {
				a[i] = 0.
				b[i] = 0.
				c[i] = t.K[i]/t.dz[i] + cp[i] + du[i]
				d[i] = 0.
			} else {
				a[i] = -t.K[i-1]/t.dz[i-1] - du[i-1]
				c[i] = -t.K[i+1] / t.dz[i]
				b[i] = t.K[i]/t.dz[i-1] + t.K[i]/t.dz[i] + cp[i] + du[i]
				d[i] = f[i-1] - f[i] + (waterDensity*t.v[i]*(t.T[i]-t.Tl[i]))/dt
				massBalance += math.Abs(d[i])
			}
		}
		mmaths.ThomasBoundaryCondition(a, b, c, d, dpsi, 1, NsubLay)

		for i := 1; i <= NsubLay; i++ {
			t.Psi[i] -= dpsi[i]
			t.Psi[i] = math.Min(t.Psi[i], t.PM[0].He)
			t.T[i] = t.PM[i].GetTheta(t.Psi[i])
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
	return false, nIter, 0.
}

// NewtonRapsonMFP is the Matric Flux Potential solution with a Newton Rapson solver
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
	massBalance := 1.
	for massBalance > tolerance && nIter < MaxIter {
		massBalance = 0.
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
