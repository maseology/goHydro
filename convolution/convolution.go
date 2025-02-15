package convolution

type Convolution struct{ w, q []float64 }

func (cv *Convolution) Update(qIn float64) float64 {
	for i := 1; i < len(cv.w); i++ {
		cv.q[i] = cv.q[i-1]
	}
	cv.q[0] = qIn

	qOut := 0.
	for i, w := range cv.w {
		qOut += w * cv.q[i]
	}
	return qOut
}
