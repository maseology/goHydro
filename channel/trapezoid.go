package channel

import (
	"math"

	"github.com/maseology/glbopt"
	"github.com/maseology/mmaths"
)

type Trapezoid struct{ Z1, Z2, B, N, S float64 }

// https://www.lmnoeng.com/Channels/trapezoid.php
func (t *Trapezoid) DepthArea(Q float64) (float64, float64) {
	tpar := func(y float64) (T, P, A, R float64) {
		T = t.B + y*(t.Z1+t.Z2)                                     // top width
		P = t.B + y*(math.Sqrt(1+t.Z1*t.Z1)+math.Sqrt(1+t.Z2*t.Z2)) // wetted perimeter
		A = y * (T + t.B) / 2                                       // area
		R = A / P                                                   // hydraulic radius
		return
	}
	trnsfrm := func(u float64) float64 {
		return mmaths.LinearTransform(.01, 10., u)
	}
	lhs := func(u float64) float64 {
		y := trnsfrm(u)
		_, _, A, R := tpar(y)
		rhs := t.N * Q / math.Sqrt(t.S) // knowns
		return math.Abs(A*math.Pow(R, 2./3.) - rhs)
	}

	uFib, _ := glbopt.Fibonacci(lhs)
	y := trnsfrm(uFib)
	_, _, A, _ := tpar(y)

	return y, A
}
