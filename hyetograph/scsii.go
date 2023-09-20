package hyetograph

import (
	"strconv"
	"strings"
)

func SCSII(nstep, dur int) []float64 {
	var t0 []float64
	switch dur {
	case 6:
		t0 = scsiihr(scsii6)
	case 12:
		t0 = scsiihr(scsii12)
	case 24:
		t0 = scsiihr(scsii24)
	case 48:
		t0 = scsiihr(scsii48)
	default:
		panic("bad diration called")
	}

	lt0 := len(t0)
	if nstep < 1 {
		return t0
	}
	t1 := make([]float64, nstep*lt0)
	for j, v := range t0 {
		for i := 0; i < nstep; i++ {
			t1[j*nstep+i] = v
		}
	}

	t2 := make([]float64, nstep)
	ss := 0.
	for i := 0; i < nstep; i++ {
		for j := 0; j < lt0; j++ {
			t2[i] += t1[i*lt0+j]
		}
		t2[i] /= float64(nstep)
		ss += t2[i]
	}

	return t2
}

func scsiihr(tbl string) []float64 {
	rpl := strings.Replace(tbl, "depth=", " ", -1)
	stop := "smoothing=false"
	pos := strings.Index(rpl, stop) + len(stop)

	o := []float64{}
	for _, s := range strings.Split(rpl[pos:], " ") {
		t, err := strconv.ParseFloat(strings.Replace(s, "\n", "", -1), 64)
		if err != nil {
			continue
		}
		o = append(o, t)
	}
	// return o

	oo := make([]float64, len(o))
	for i, v := range o {
		if i == 0 {
			oo[i] = v
		} else {
			oo[i] = v - o[i-1]
		}
	}
	return oo
}
