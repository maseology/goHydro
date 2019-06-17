package transfunc

// TF is a general transfer function implimenter
type TF struct {
	SQ, QT []float64
}

// NewTF creates a new TF struct
func NewTF(base, skew, lag float64) TF {
	qt := Triangular(base, skew, lag) // MAXBAS: triangular weighted transfer function
	return TF{
		QT: qt,
		SQ: make([]float64, len(qt)+1), // delayed runoff
	}
}
