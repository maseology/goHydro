package hru

// Res simple linear reservoir
// Sto returns total current storage
type Res struct {
	Sto, Cap float64
}

// Initialize Res
func (r *Res) Initialize(init, cap float64) {
	r.Sto = init
	r.Cap = cap
}

// Overflow : update state. p is an net addition and function returns excess (+).
// If p<0 and |p|>sto, function returns remainder of sink (-).
func (r *Res) Overflow(p float64) float64 {
	r.Sto += p
	if r.Sto < 0. {
		d := r.Sto
		r.Sto = 0.
		return d
	} else if r.Sto > r.Cap {
		d := r.Sto - r.Cap
		r.Sto = r.Cap
		return d
	} else {
		return 0.
	}
}

// Skim returns excess (sto-cap>0) and resets sto=cap.
// if negative, Skim returns the negative of Deficit.
func (r *Res) Skim() float64 {
	x := r.Sto - r.Cap
	if x > 0. {
		r.Sto = r.Cap
	}
	return x
}

// // Storage returns total current storage
// func (r *Res) Storage() float64 {
// 	return r.sto
// }

// Deficit returns current storage deficit
func (r *Res) Deficit() float64 {
	return r.Cap - r.Sto
}
