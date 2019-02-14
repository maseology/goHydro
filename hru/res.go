package hru

// res simple linear reservoir
type res struct {
	sto, cap float64
}

// overflow : update state. p is an net addition and function returns excess.
// If p<0 and |p|>sto, function returns remainder of sink
func (r *res) overflow(p float64) float64 {
	r.sto += p
	if r.sto < 0. {
		d := r.sto
		r.sto = 0.
		return d
	} else if r.sto > r.cap {
		d := r.sto - r.cap
		r.sto = r.cap
		return d
	} else {
		return 0.
	}
}

// Storage returns total current storage
func (r *res) Storage() float64 {
	return r.sto
}
