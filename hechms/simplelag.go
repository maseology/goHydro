package hechms

type simplelag struct {
	trnfrm, lag []float64
	lag0        int
}

func (sl *simplelag) Update(i float64) (o float64) {
	if i < 0 {
		o = sl.lag[0] // pop reach (volumetric) flow [mm.km2]
		sl.lag = append(sl.lag[1:], 0.)
		return
	}
	for v, u := range sl.trnfrm { // add to storage
		// if jj+v < sl.lag0 {
		// 	sl.lag[v] = bsn[i].qbf
		// } else {
		sl.lag[v] += i * u // downstream reach transform transform [mm.km2]
		// }
	}
	return
}
