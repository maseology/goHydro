package lia // VERSION 2

import "math"

// GetVelocities returns the current cell velocities
func (d *Domain) GetVelocities() map[int]float64 {
	mout := make(map[int]float64, d.GF.GD.Nact)
	for i := range d.GF.GD.Sactives {
		dpth := d.ns[i].h - d.ns[i].z // depth
		nfs := d.GF.CellFace[i]       // node faces
		if dpth > 0.002*d.dx {
			mout[i] = math.Sqrt(math.Pow((d.qs[nfs[2]].q+d.qs[nfs[0]].q)/2., 2.)+math.Pow((d.qs[nfs[3]].q+d.qs[nfs[1]].q)/2., 2.)) / dpth
			// mout[i] = math.Sqrt(math.Pow(float64()+math.Pow(float64(d.qs[nfs[3]]+d.qs[nfs[1]])/2., 2.)) / dpth
		} else {
			mout[i] = 0.
		}
	}
	return mout
}
