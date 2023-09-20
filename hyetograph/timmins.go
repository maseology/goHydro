package hyetograph

// Timmins design storm
// Environmental Water Resoureces Group, 2017. Technical Guidelines for Flood Hazard Mapping (March, 2017). 137pp.
func Timmins(arf float64) []float64 {
	timmins := []float64{15, 20, 10, 3, 5, 20, 43, 20, 23, 13, 13, 8}
	o := make([]float64, len(timmins))
	s := 0.
	for i, t := range timmins {
		s += t
		o[i] = t * arf
	}
	return o
}
