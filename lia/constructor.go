package lia

import "github.com/maseology/goHydro/grid"

func newFace(gf *grid.Face, fid int) *face {
	var f face
	f.nfrom = gf.FaceCell[fid][0]
	f.nto = gf.FaceCell[fid][1]
	if f.nfrom == -1 || f.nto == -1 {
		f.q = 0. // (default) no flow boundary
	} else {
		f.forth = make([]int, 4)  // orthogonal faces
		if gf.IsUpwardFace(fid) { // upward meaning direction normal to face
			f.ffw = gf.CellFace[f.nto][1]
			f.fbw = gf.CellFace[f.nfrom][3]
			f.forth[0] = gf.CellFace[f.nfrom][2]
			f.forth[1] = gf.CellFace[f.nfrom][0]
			f.forth[2] = gf.CellFace[f.nto][2]
			f.forth[3] = gf.CellFace[f.nto][0]
		} else {
			f.ffw = gf.CellFace[f.nto][0]
			f.fbw = gf.CellFace[f.nfrom][2]
			f.forth[0] = gf.CellFace[f.nfrom][3]
			f.forth[1] = gf.CellFace[f.nfrom][1]
			f.forth[2] = gf.CellFace[f.nto][3]
			f.forth[3] = gf.CellFace[f.nto][1]
		}
	}
	return &f
}
