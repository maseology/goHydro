package hechms

import (
	"math"
)

type muskingum struct{ icur, ilast, olast, c0, c1, c2 float64 }

// x weighting factor [-] ""
// k time constant or storage coefficient [hr] "travel time"
// x=0: linear reservour S=ko; k=dt x=.5 translation by k
func NewMuskingum(x, k, q0, dt float64) *muskingum {
	c0 := (dt - 2*k*x) / (2*k*(1-x) + dt)
	c1 := (dt + 2*k*x) / (2*k*(1-x) + dt)
	c2 := (2*k*(1-x) - dt) / (2*k*(1-x) + dt)
	if c0 < 0 {
		panic("muskingum error: c0 < 0 ")
	}
	if math.Abs(c0+c1+c2-1) > 1e-5 {
		println(c0 + c1 + c2)
		panic("muskingum error1")
	}
	// fmt.Println(c0, c1, c2)
	return &muskingum{
		c0:    c0,
		c1:    c1,
		c2:    c2,
		olast: q0,
		ilast: q0,
	}
}

func (mk *muskingum) Update(i float64) (o float64) {
	if i < 0 {
		o = mk.c0*mk.icur + mk.c1*mk.ilast + mk.c2*mk.olast
		mk.olast = o
		mk.ilast = mk.icur
		mk.icur = 0
		return
	}
	mk.icur += i
	return
}
