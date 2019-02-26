package tem

// TEC topologic elevation model cell
type TEC struct {
	Z, S, A float64
	Ds      int
}

// New constructor
func (t *TEC) New(z, s, a float64, ds int) {
	t.Z = z   // elevation
	t.S = s   // slope/gradient (rise/run)
	t.A = a   // aspect (counter-clockwise from east)
	t.Ds = ds // downslope id
}

// // TECXY topologic elevation model cell
// type TECXY struct {
// 	X, Y, Z, S, A float64
// 	Ds            int
// }

// // New constructor
// func (t *TECXY) New(x, y, z, s, a float64, ds int) {
// 	t.X = x   // x-coordinate
// 	t.Y = y   // y-coordinate
// 	t.Z = z   // elevation
// 	t.S = s   // slope/gradient (rise/run)
// 	t.A = a   // aspect (counter-clockwise from east)
// 	t.Ds = ds // downslope id
// }
