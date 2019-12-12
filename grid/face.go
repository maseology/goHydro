package grid

import "github.com/maseology/mmaths"

// Face definition struct
// collection of cell IDs surrounding grid faces
// EXAMPLE: 4x5 cells = 2*(4x5)+4+5 = 49 faces
// gd1 := Definition("test", 4, 5)
//  o---o---o---o---o---o       o-0-o-1-o-2-o-3-o-4-o      o---o---o---o---o---o
//  | 0 | 1 | 2 | 3 | 4 |       |   |   |   |   |   |     25  26  27  28  29  30        North-east              |
//  o---o---o---o---o---o       o-5-o-6-o-7-o-8-o-9-o      o---o---o---o---o---o     ^   positive          from | to
//  | 5 | 6 | 7 | 8 | 9 |       |   |   |   |   |   |     31  32  33  34  35  36    /|\                         |
//  o---o---o---o---o---o  -->  o-10o-11o-12o-13o-14o  &   o---o---o---o---o---o     |
//  | 10| 11| 12| 13| 14|       |   |   |   |   |   |     37  38  39  40  41  42     |   +                     to
//  o---o---o---o---o---o       o-15o-16o-17o-18o-19o      o---o---o---o---o---o     |_________\             ------
//  | 15| 16| 17| 18| 19|       |   |   |   |   |   |     43  44  45  46  47  48               /              from
//  o---o---o---o---o---o       o-20o-21o-22o-23o-24o      o---o---o---o---o---o
type Face struct {
	gd                 *Definition
	cxy                map[int]mmaths.Point
	CellFace, FaceCell map[int][]int
	boundface          []int
	Nfaces, isplit     int
}

// NewFace creates a new Face struct
func NewFace(gd *Definition) *Face {
	var f Face
	f.gd = gd
	ncell := gd.Ncells()
	f.Nfaces = 2*gd.nr*gd.nc + gd.nr + gd.nc
	f.CellFace = make(map[int][]int, ncell)
	f.cxy = make(map[int]mmaths.Point, ncell)
	f.FaceCell = make(map[int][]int, f.Nfaces)
	f.isplit = (gd.nr + 1) * gd.nc
	for i := 0; i < gd.nr; i++ {
		for j := 0; j < gd.nc; j++ {
			//   1
			// 2   0
			//   3
			in1, cid := make([]int, 4), gd.CellID(i, j)
			in1[1] = cid                    // up
			in1[3] = cid + gd.nc            // down
			in1[2] = cid + f.isplit + i     // left
			in1[0] = cid + f.isplit + i + 1 // right
			f.CellFace[cid] = in1
			f.cxy[cid] = gd.Coord[cid]
		}
	}
	for i := 0; i < f.Nfaces; i++ {
		f.FaceCell[i] = []int{-1, -1}
	}
	for k, cf := range f.CellFace {
		if !gd.IsActive(k) {
			continue
		}
		f.FaceCell[cf[0]][0] = k
		f.FaceCell[cf[1]][0] = k
		f.FaceCell[cf[2]][1] = k
		f.FaceCell[cf[3]][1] = k
	}

	f.boundface = []int{}
	for k, fc := range f.FaceCell {
		for _, i := range fc {
			if i == -1 {
				f.boundface = append(f.boundface, k)
				break
			}
		}
	}
	return &f
}

// IsUpwardFace returns whether the orientation of the face is normal to the vertical
func (f *Face) IsUpwardFace(fid int) bool {
	return fid < f.isplit
}
