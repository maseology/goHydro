package hyetograph

func Unit(nstep int) []float64 {
	o := make([]float64, nstep)
	v := 1. / float64(nstep)
	for i := 0; i < nstep; i++ {
		o[i] = v
	}
	return o
}
