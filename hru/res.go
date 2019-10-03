package hru

// Res simple linear reservoir
type Res struct {
	sto, cap float64
}

// Initialize Res
func (r *Res) Initialize(init, cap float64) {
	r.sto = init
	r.cap = cap
}

// Overflow : update state. p is an net addition and function returns excess (+).
// If p<0 and |p|>sto, function returns remainder of sink (-).
func (r *Res) Overflow(p float64) float64 {
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
func (r *Res) Storage() float64 {
	return r.sto
}

// Deficit returns current storage deficit
func (r *Res) Deficit() float64 {
	return r.cap - r.sto
}

// Skim returns excess (sto-cap>0) and resets sto=cap.
// if negative, Skim returns the negative of Deficit.
func (r *Res) Skim() float64 {
	x := r.sto - r.cap
	if x > 0. {
		r.sto = r.cap
	}
	return x
}
