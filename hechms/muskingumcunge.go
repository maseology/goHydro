package hechms

import (
	"fmt"
	"math"
)

/*
https://uon.sdsu.edu/variable_parameter_muskingum_cunge_method_revisited.html
https://ponce.sdsu.edu/muskingum_cunge_method_with_variable_parameters.html
https://ponce.sdsu.edu/
https://ponce.sdsu.edu/muskingum_cunge_method_explained.html
https://www.lmnoeng.com/Channels/trapezoid.php
*/

func MCvar() {
	Qp := 1000. // cms
	Ap := 400.  // m2
	L := 14.4   // km
	beta := 1.6
	ts := 1. // timestep [hr]

	Vp := Qp / Ap
	cp := beta * Vp                    // flood wave celerity
	dt := math.Min(ts*3600, L*1000/Vp) // timestep
	dx := cp * dt

	fmt.Println(dx, dt)

}

func MC() {

	Qp := 1000.    // cms
	Qb := 0.       // cms
	Ap := 400.     // m2
	Tp := 100.     // m
	So := 0.000868 // [-]
	beta := 1.6    // [-]
	L := 14.4      // km
	dt := 1.       // [hr]

	dx := L * 1000          // [m]
	Qo := Qb + (Qp-Qb)/2    // reference discharge
	V := Qp / Ap            // channel velocity
	qo := Qp / Tp           // unit discharge (at reference discharge)
	c := beta * V           // flood wave celerity
	C := c * dt * 3600 / dx // Courant number
	D := qo / So / c / dx

	_ = Qo

	c0 := (-1 + C + D) / (1 + C + D)
	c1 := (1 + C - D) / (1 + C + D)
	c2 := (1 - C + D) / (1 + C + D)

	fmt.Println(c0, c1, c2, c0+c1+c2)

	Qin := []float64{0, 200, 400, 600, 800, 1000, 800, 600, 400, 200, 0, 0, 0, 0}
	Qout := make([]float64, len(Qin))
	for i, I := range Qin {
		if i == 0 {
			Qout[i] = Qb
		} else {
			Qout[i] = c0*I + c1*Qin[i-1] + c2*Qout[i-1]
		}
	}
	fmt.Println(Qout)
}

// const ncell = 200

// type muskingumcunge struct {
// 	rchs                       []muskingumcungeseg
// 	dt, beta, s, Qo, L, So, co float64
// }

// // type muskingumcungeseg struct{ icur, ilast, olast, c0, c1, c2 float64 }
// type muskingumcungeseg struct{ q0, q1 float64 }

// // dx [m]; dt [hr]
// func NewMuskingumCunge(Qp, Qb, So, beta, L, dt float64) *muskingumcunge {
// 	// https://ponce.sdsu.edu/muskingum_cunge_method_with_variable_parameters.html

// 	V := Qp / Ap

// 	c := beta * V
// 	qo := Q0 / Tp

// 	// dx := L / ncell

// 	mcs := make([]muskingumcungeseg, ncell)
// 	for j := 0; j < ncell; j++ {
// 		mcs[j].q0 = qo
// 		mcs[j].q1 = qo
// 	}
// 	return &muskingumcunge{
// 		rchs: mcs,
// 		// dx:   dx,
// 		dt:   dt,
// 		beta: beta,
// 		s:    So,
// 		Qo:   Qb + (Qp-Qb)/2,
// 		L:    L, // reach length
// 	}

// }

// // Variable-parameter Muskingum-Cunge method revisited (1994)
// // https://uon.sdsu.edu/variable_parameter_muskingum_cunge_method_revisited.html
// func (mc *muskingumcunge) Update(i float64) (o float64) {

// 	dx := mc.co * mc.dt
// 	if dx >= (mc.co*mc.dt+mc.Qo/mc.T/mc.So/mc.co)/2 {

// 	}

// 	if i < 0 {
// 		// mc := *mc
// 		for j := 1; j < ncell; j++ {
// 			q := (mc.rchs[j-1].q0 + mc.rchs[j-1].q1 + mc.rchs[j].q0) / 3
// 			c := mc.beta * q / y

// 			C := c * 3600 * mc.dt / mc.dx
// 			D := q / mc.s / c / mc.dx

// 			c0 := (-1 + C + D) / (1 + C + D)
// 			c1 := (1 + C - D) / (1 + C + D)
// 			c2 := (1 - C + D) / (1 + C + D)

// 			if math.Abs(c0+c1+c2-1) > 1e-5 {
// 				println(c0 + c1 + c2)
// 				panic("NewMuskingumCunge error1")
// 			}
// 			// fmt.Println(c0, c1, c2Println(a)

// 			mc.rchs[j].q0 = mc.rchs[j].q1
// 			mc.rchs[j].q1 = q

// 		}

// 		// o = mc.c0*mc.icur + mc.c1*mc.ilast + mc.c2*mc.olast
// 		mc.olast = o
// 		mc.ilast = mc.icur
// 		mc.icur = 0
// 		return
// 	}
// 	mc.icur += i
// 	return
// }
