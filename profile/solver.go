package profile

/*
	Vapour Transport

	assumes:
		- considers only vapour flux at the soil surface
		- isothermal and isobaric throughout the timestep
		- uniform soil temperature (model meant for soil surface)
*/

import (
	"fmt"
	"math"
)

const (
	isFreeDrainage = true
	isInfiltrating = true // set to false to active evaporation

	maxTimeStep = 3600. // [s]
	maxIter     = 100
	tolerance   = 1e-6

	// physical constants
	rhow = 1000.   // density of water [kg/m³]
	rhoa = 1.225   // density of air [kg/m³]
	mw   = 0.01802 // molecular weigh of water [kg/mol]
	eta  = 0.66    // Penman (1940)
	kap  = 0.4     // von-Karman
	r    = 8.3143  // gas constant [J/mol/K]
	g    = 9.8065  // gravitational acceleration [m²/s]

	// variables kept costant
	// pa = 101325.0     // 1 atm: standard pressure [Pa = N/m² = J/m³ = kg/m/s²]
	wa = 0.5      // relative humidity of the atmosphere [-]
	ta = 293.     // air temperature [K] (~20°C)
	ts = 293.     // soil temperature [K]
	qp = 1.43e-2  // saturated specific humidity (moisture content) of pore-space air [kg/kg] at temperature ts [°C], pressure pa [kPa]
	qa = 7.16e-3  // specific humidity (moisture content) of air [kg/kg] at temperature ta, pressure pa and relative humidity wa
	da = 2.12e-5  // coefficient of molecular diffusion of water vapour in air [m²/s] (pg.43 Bittelli)
	kv = 1 / 100. // water vapour turbulent transport coefficient [m/s] =1/ra
	// ep = 0.21 / 3600. // potential evaportation rate [mm/s]
)

// Solve vertical variable-timestep Newton-Raphson solution to richards equation with vapour flux
func (ps *State) Solve(simLenHr float64) (t, f []float64, ok bool) {
	// control
	endTime := simLenHr * 3600. // sec
	dt := maxTimeStep / 10.

	// counters
	time := 0.
	sumFlx := 0.
	totiter := 0

	// boundary conditions
	if isInfiltrating {
		ps.setToInfiltrating()
	}
	if isFreeDrainage {
		ps.setToFreeDraining()
	}

	// solve
	t, f = []float64{}, []float64{}
	for time < endTime {
		dt = math.Min(dt, endTime-time)
		ok, initer, flx := ps.cellCenteredFiniteVolume(dt)
		// ok, initer, flx := ps.newtonRaphson(dt) // tends to work only for single material profiles
		totiter += initer
		if ok {
			for i := 0; i <= nsl+1; i++ {
				ps.save()
			}
			sumFlx += flx * dt
			time += dt
			fmt.Printf(" time = %d\tdt = %.2f\titer = %d\tinfil = %.3f\n", int(time), dt, initer, sumFlx)
			if isInfiltrating {
				f = append(f, flx) // infiltration
				t = append(t, time)
			} else if time > dt {
				f = append(f, -flx*3600.) // evaporation
				t = append(t, time/3600.)
			}

			if float64(initer)/float64(maxIter) < 0.1 {
				dt = math.Min(dt*2., maxTimeStep) // increase timestep
			}
		} else {
			fmt.Printf(" dt = %.3f\tNo convergence\n", dt)
			if dt == 0.001 {
				fmt.Println(" solution would not converge")
				return t, f, false
			}
			dt = math.Max(dt/2., 0.001) // reduce timestep
			for i := 0; i <= nsl+1; i++ {
				ps.reset()
			}
		}
	}
	fmt.Printf("number of iterations per hour: %.1f\n\n", float64(totiter)/simLenHr)
	return t, f, true
}

// newtonRaphson is the Newton Rapson solution to the matrix potential form with vapour transfer
// tends to be unstable for multi-layered profiles
func (ps *State) newtonRaphson(dt float64) (bool, int, float64) {

	u, dudp := make(map[int]float64, nsl), make(map[int]float64, nsl) // source/sink term and in differential form
	cp, f := make(map[int]float64, nsl), make(map[int]float64, nsl)   // cp=rho C dz/dt; flux

	// tri-diagonal matrix coefficients
	a, b, c, d, dpsi := make(map[int]float64, nsl), make(map[int]float64, nsl), make(map[int]float64, nsl), make(map[int]float64, nsl), make(map[int]float64, nsl)

	nIter := 0
	massBalance := 1.
	for massBalance > tolerance && nIter < maxIter {
		massBalance = 0.
		for i := 1; i <= nsl; i++ {
			ps.K[i] = ps.PM[i].GetK(ps.t[i])                                             // liquid conductance
			u[i] = -g * ps.K[i]                                                          // gravitational flux
			ps.K[i] += (ps.PM[i].Ts - ps.t[i]) * mw * eta * rhoa * da * ps.q[i] / r / ts //ps.PM[i].GetKvap(ps.q[i], ps.t[i]) // vapour conductance
			dudp[i] = -u[i] * ps.PM[i].cn / ps.p[i]
			dtdp := ps.PM[i].dtdp(ps.p[i])
			dqdp := ps.q[i] * mw / r / ts                    // ps.PM[i].dqdp(ps.q[i])
			cp[i] = ps.vol[i] * (rhow*dtdp + rhoa*dqdp) / dt // storage pg.174
		}

		for i := 1; i <= nsl; i++ {
			f[i] = -(ps.K[i+1]*ps.p[i+1]-ps.K[i]*ps.p[i])/ps.dz[i]/ps.PM[i].b2 + u[i] // eq.8.52 pg.178
			if i == 1 {
				if isInfiltrating {
					a[i] = 0.
					c[i] = 0. // set to zero when constant potential at top
					b[i] = ps.K[i]/ps.dz[i] + cp[i] + dudp[i]
					d[i] = 0. // set to zero when constant potential at top
				} else {
					a[i] = 0.
					c[i] = -ps.K[i+1] / ps.dz[i]
					b[i] = ps.K[i]/ps.dz[i] + cp[i] + dudp[i]
					// wp := math.Exp(mw * ps.p[i] / (r * ts))
					// fe := ep * (wp - wa) / (1. - wa) // evaporation flux
					// fe := rhoa * kv * (ps.q[i] - qa) * (ps.PM[i].Ts - ps.t[i])
					fe := rhoa * kv * (ps.PM[i].Ts*(ps.q[i]-qa) + ps.t[i]*(qp-ps.q[i]))
					d[i] = f[i] - fe - ps.vol[i]*(rhow*(ps.t[i]-ps.tl[i])+rhoa*(ps.q[i]-ps.ql[i])*(ps.PM[i].Ts-ps.t[i]))/dt
				}
			} else {
				a[i] = -ps.K[i-1]/ps.dz[i-1] - dudp[i-1]
				c[i] = -ps.K[i+1] / ps.dz[i]
				b[i] = ps.K[i]/ps.dz[i-1] + ps.K[i]/ps.dz[i] + cp[i] + dudp[i]
				d[i] = f[i] - f[i-1] - ps.vol[i]*(rhow*(ps.t[i]-ps.tl[i])+rhoa*(ps.q[i]-ps.ql[i])*(ps.PM[i].Ts-ps.t[i]))/dt
				massBalance += math.Abs(d[i])
			}
		}
		thomasBoundaryCondition(a, b, c, d, dpsi, 1, nsl)

		for i := 1; i <= nsl; i++ {
			ps.p[i] += dpsi[i]
			ps.p[i] = math.Min(ps.p[i], ps.PM[i].He)
			ps.t[i] = ps.PM[i].GetTheta(ps.p[i])
			ps.q[i] = qp * math.Exp(mw*ps.p[i]/r/ts) // ps.PM[i].GetSpecificHumidity(ps.p[i])
		}

		if isFreeDrainage {
			ps.p[nsl+1] = ps.p[nsl]
			ps.t[nsl+1] = ps.t[nsl]
			ps.K[nsl+1] = ps.K[nsl]
			ps.q[nsl+1] = ps.q[nsl]
		}
		nIter++
	}
	if massBalance < tolerance {
		return true, nIter, -f[1]
	}
	return false, nIter, 0.
}

// cellCenteredFiniteVolume is a a cell-centered finite volume solution to richards equation
func (ps *State) cellCenteredFiniteVolume(dt float64) (bool, int, float64) {

	h, h0, cp, f := make(map[int]float64, nsl), make(map[int]float64, nsl), make(map[int]float64), make(map[int]float64)
	a, b, c, d := make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64)

	kavg := func(k1, k2 float64) float64 { // logarithmic mean
		if k1 == k2 {
			return k1
		}
		return (k1 - k2) / math.Log(k1/k2)
	}

	sum0 := 0.0
	for i := 1; i <= nsl; i++ {
		h0[i] = ps.p[i] + ps.cz[i]*g //ps.p[i]/g + ps.cz[i] // ps.p[i] - ps.cz[i]*g
		h[i] = h0[i]
		sum0 += ps.vol[i] * (rhow*ps.t[i] + rhoa*ps.q[i]*(ps.PM[i].Ts-ps.t[i])) // vapour mass
	}

	massBalance := sum0
	nIter := 0
	for massBalance > tolerance && nIter < maxIter {
		for i := 1; i <= nsl; i++ {
			ps.K[i] = ps.PM[i].GetK(ps.t[i])
			ps.K[i] += (ps.PM[i].Ts - ps.t[i]) * mw * eta * rhoa * da * ps.q[i] / r / ts // vapour conductance
			dtdh := ps.PM[i].dtdh(h0[i], h[i], ps.cz[i])
			dqdh := ps.q[i] * mw / r / ts                    // vapour (dqdh=dqdp)
			cp[i] = ps.vol[i] * (rhow*dtdh + rhoa*dqdh) / dt // vapour
		}

		f[0] = 0
		for i := 1; i <= nsl; i++ {
			f[i] = carea * kavg(ps.K[i], ps.K[i+1]) / ps.dz[i] // [kg s/m²]
		}

		for i := 1; i <= nsl; i++ {
			a[i] = -f[i-1]
			if i == 1 {
				if isInfiltrating {
					b[i] = 1.
					c[i] = 0.
					d[i] = cp[i] * h0[i]
				} else {
					b[i] = cp[i] + f[i]
					c[i] = -f[i]
					fe := -rhoa * kv * (ps.PM[i].Ts*(ps.q[i]-qa) + ps.t[i]*(qp-ps.q[i]))
					d[i] = cp[i]*h0[i] + carea*fe
				}
			} else if i < nsl {
				b[i] = cp[i] + f[i-1] + f[i]
				c[i] = -f[i]
				d[i] = cp[i] * h0[i]
			} else {
				b[nsl] = cp[nsl] + f[nsl-1]
				c[nsl] = 0.
				if isFreeDrainage {
					d[nsl] = cp[nsl]*h0[nsl] - carea*ps.K[nsl]*g
				} else {
					d[nsl] = cp[nsl]*h0[nsl] - f[nsl]*(h[nsl]-h[nsl+1])
				}
			}
		}

		thomasBoundaryCondition(a, b, c, d, h, 1, nsl)

		new1 := 0.
		for i := 1; i <= nsl; i++ {
			ps.p[i] = h[i] - g*ps.cz[i]
			ps.t[i] = ps.PM[i].GetTheta(ps.p[i])
			ps.q[i] = qp * math.Exp(mw*ps.p[i]/r/ts)
			new1 += ps.vol[i] * (rhow*ps.t[i] + rhoa*ps.q[i]*(ps.PM[i].Ts-ps.t[i])) // vapour mass
		}

		if isFreeDrainage {
			ps.p[nsl+1] = ps.p[nsl]
			ps.t[nsl+1] = ps.t[nsl]
			ps.K[nsl+1] = ps.K[nsl]
			ps.q[nsl+1] = ps.q[nsl]
			massBalance = math.Abs(new1 - (sum0 + f[1]*(h[1]-h[2])*dt - carea*ps.K[nsl]*g*dt))
		} else {
			massBalance = math.Abs(new1 - (sum0 + f[1]*(h[1]-h[2])*dt - f[nsl]*(h[nsl]-h[nsl+1])*dt)) // h[nsl+1] can be a specified head
		}
		nIter++
	}

	if massBalance < tolerance {
		return true, nIter, f[1] * (h[1] - h[2])
	}
	return false, nIter, 0.
}

func thomasBoundaryCondition(a, b, c, d, x map[int]float64, first, last int) {
	for i := first; i < last; i++ {
		c[i] /= b[i]
		d[i] /= b[i]
		b[i+1] -= a[i+1] * c[i]
		d[i+1] -= a[i+1] * d[i]
	}
	// back substitution
	x[last] = d[last] / b[last]
	for i := last - 1; i > first-1; i-- {
		x[i] = d[i] - c[i]*x[i+1]
	}
}
