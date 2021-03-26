package glue

// Generalized Likelihood Uncertainty Estimator

import "sort"

// GLUEsortInterface interfaces sort.Sort
type GLUEsortInterface interface {
	sort.Interface
	// Partition returns slice[:i] and slice[i+1:]
	// These should references the original memory
	// since this does an in-place sort
	Partition(i int) (left GLUEsortInterface, right GLUEsortInterface)
}

type GLUEi struct {
	Likelihood, Value float64
}

type GLUE []GLUEi

func (g GLUE) Less(i, j int) bool {
	return g[i].Value < g[j].Value
}

func (g GLUE) Swap(i, j int) {
	g[i].Value, g[j].Value = g[j].Value, g[i].Value
}

func (g GLUE) Len() int {
	return len(g)
}

// Partition : splits index array around pivot
func (g GLUE) Partition(i int) (left GLUEsortInterface, right GLUEsortInterface) {
	return GLUE(g[:i]), GLUE(g[i+1:])
}

func (g GLUE) P5o95() (float64, float64) {
	// must be sorted
	c := 0.
	for _, v := range g {
		c += v.Likelihood
	}
	p, o5, o95 := 0., -9999., 0.
	for _, v := range g {
		p += v.Likelihood / c
		if o5 == -9999. && p > .05 {
			o5 = v.Value
			continue
		}
		if p > .95 {
			o95 = v.Value
			break
		}
	}
	return o5, o95
}
