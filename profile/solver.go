package profile

/*
	Vapour Transport

	assumes:
		- considers only vapour flux at the soil surface
*/

import (
	"fmt"
	"math"

	"github.com/maseology/mmaths"
)

const (
	maxTimeStep = 3600. // [s]
	maxIter     = 100
	tolerance   = 1e-9

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
	wa = 0.5          // relative humidity of the atmosphere [-]
	ta = 293.         // air temperature [K] (~20°C)
	ts = 293.         // soil temperature [K]
	qp = 1.43e-2      // saturated specific humidity (moisture content) of pore-space air [kg/kg] at temperature ts [°C], pressure pa [kPa]
	qa = 7.16e-3      // specific humidity (moisture content) of air [kg/kg] at temperature ta, pressure pa and relative humidity wa
	da = 2.12e-5      // coefficient of molecular diffusion of water vapour in air [m²/s] (pg.43 Bittelli)
	kv = 1 / 100.     // water vapour turbulent transport coefficient [m/s] =1/ra
	ep = 0.21 / 3600. // potential evaportation rate [mm/s]

	isFreeDrainage = true
	isInfiltrating = false // set to false to active evaporation
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
		ok, initer, flx := ps.newtonRaphson(dt) // tends to work only for single material profiles
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
			ps.K[i] = ps.PM[i].GetK(ps.t[i])              // liquid conductance
			u[i] = -g * ps.K[i]                           // gravitational flux
			ps.K[i] += ps.PM[i].GetKvap(ps.q[i], ps.t[i]) // vapour conductance
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
		mmaths.ThomasBoundaryCondition(a, b, c, d, dpsi, 1, nsl)

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

// finiteVolume is a a cell-centered finite volume solution to richards equation
func (ps *State) finiteVolume(dt float64) (bool, int, float64) {

	h0, cp, f := make(map[int]float64, nsl), make(map[int]float64), make(map[int]float64)
	a, b, c, d := make(map[int]float64), make(map[int]float64), make(map[int]float64), make(map[int]float64)

	sum0 := 0.0
	for i := 1; i <= nsl; i++ {
		h0[i] = ps.p[i] - ps.cz[i]*g
		ps.h[i] = h0[i]
		sum0 += rhow * ps.vol[i] * ps.t[i]
	}

	massBalance := sum0
	nIter := 0
	// for massBalance > tolerance && nIter < MaxIter {
	// 	for i := 1; i <= NsubLay; i++ {
	// 		t.K[i] = t.PM[i].GetK(t.T[i])
	// 		cap := t.PM[i].dThetadH(h0[i], t.H[i], t.cz[i])
	// 		cp[i] = (waterDensity * t.vol[i] * cap) / dt
	// 	}

	// 	f[0] = 0
	// 	for i := 1; i <= NsubLay; i++ {
	// 		f[i] = area * meanK(t.K[i], t.K[i+1]) / t.dz[i]
	// 	}

	// 	for i := 1; i <= NsubLay; i++ {
	// 		a[i] = -f[i-1]
	// 		if i == 1 {
	// 			b[i] = 1.0
	// 			c[i] = 0.0
	// 			d[i] = h0[i]
	// 		} else if i < NsubLay {
	// 			b[i] = cp[i] + f[i-1] + f[i]
	// 			c[i] = -f[i]
	// 			d[i] = cp[i] * h0[i]
	// 		} else {
	// 			b[NsubLay] = cp[NsubLay] + f[NsubLay-1]
	// 			c[NsubLay] = 0.0
	// 			if isFreeDrainage {
	// 				d[NsubLay] = cp[NsubLay]*h0[NsubLay] - area*t.K[NsubLay]*g
	// 			} else {
	// 				d[NsubLay] = cp[NsubLay]*h0[NsubLay] - f[NsubLay]*(t.H[NsubLay]-t.H[NsubLay+1])
	// 			}
	// 		}
	// 	}

	// 	mmaths.ThomasBoundaryCondition(a, b, c, d, t.H, 1, NsubLay)

	// 	newSum := 0.0
	// 	for i := 1; i <= NsubLay; i++ {
	// 		t.Psi[i] = t.H[i] + g*t.cz[i]
	// 		t.T[i] = t.PM[i].GetTheta(t.Psi[i])
	// 		newSum += waterDensity * t.vol[i] * t.T[i]
	// 	}

	// 	if isFreeDrainage {
	// 		t.Psi[NsubLay+1] = t.Psi[NsubLay]
	// 		t.T[NsubLay+1] = t.T[NsubLay]
	// 		t.K[NsubLay+1] = t.K[NsubLay]
	// 		massBalance = math.Abs(newSum - (sum0 + f[1]*(t.H[1]-t.H[2])*dt - area*t.K[NsubLay]*g*dt))
	// 	} else {
	// 		massBalance = math.Abs(newSum - (sum0 + f[1]*(t.H[1]-t.H[2])*dt - f[NsubLay]*(t.H[NsubLay]-t.H[NsubLay+1])*dt))
	// 	}
	// 	nIter++
	// }

	// if massBalance < tolerance {
	// 	return true, nIter, f[1] * (t.H[1] - t.H[2])
	// }
	return false, nIter, 0.0
}
