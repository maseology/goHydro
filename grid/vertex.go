package grid

//  Indexing of verticies that surround grid cells. For example 4x5 cells (5x6 nodes):
//  0---1---2---3---4---5
//  | 0 | 1 | 2 | 3 | 4 |
//  6---7---8---9--10--11
//  | 5 | 6 | 7 | 8 | 9 |
// 12--13--14--15--16--17
//  | 10| 11| 12| 13| 14|
// 18--19--20--21--22--23
//  | 15| 16| 17| 18| 19|
// 24--25--26--27--28--29

type Vertex struct {
	gd                   *Definition
	Nodecoord            map[int][]float64
	Nodecells, Cellnodes map[int][]int
	dir                  int // 0:vtk (default) 1:clockwise  -1:counter-clockwise
}

func (gd *Definition) ToVertex() *Vertex {

	ncrd, nv := make(map[int][]float64), 0
	for cid := 0; cid < gd.Ncol*gd.Nrow; cid++ {
		r, c, xl, yu := gd.CellOriginUL(cid)
		ncrd[nv] = []float64{xl, yu}
		nv++
		if c == gd.Ncol-1 {
			ncrd[nv] = []float64{xl + gd.Cwidth, yu}
			nv++
		}
		if r == gd.Nrow-1 {
			ncrd[nv] = []float64{xl, yu - gd.Cwidth}
			nv++
			if c == gd.Ncol-1 {
				ncrd[nv] = []float64{xl + gd.Cwidth, yu - gd.Cwidth}
				nv++
			}
		}
	}

	cn := make(map[int][]int, gd.Nact) //references nodes connected to each cell
	for _, c := range gd.Sactives {
		// p2---p3   y
		//  | c |    |       //  default to VTK standard
		// p0---p1   0---x
		i, _ := gd.RowCol(c)
		cn[c] = make([]int, 4)
		cn[c][0] = c + i + gd.Ncol + 1 // p0
		cn[c][1] = c + i + gd.Ncol + 2 // p1
		cn[c][2] = c + i               // p2
		cn[c][3] = c + i + 1           // p3
	}

	nc := make(map[int][]int, len(ncrd)) // references cells associated with each node
	for i := range ncrd {
		nc[i] = []int{-1, -1, -1, -1} // initialize
	}
	for _, c := range gd.Sactives {
		// 2---1
		// | n |
		// 3---0
		nc[cn[c][0]][1] = c
		nc[cn[c][1]][2] = c
		nc[cn[c][2]][0] = c
		nc[cn[c][3]][3] = c
	}

	if gd.Nact != gd.Ncol*gd.Nrow {
		rm := []int{}
		for k, a := range nc {
			for _, v := range a {
				if v >= 0 {
					goto next
				}
			}
			rm = append(rm, k)
		next:
		}
		for _, r := range rm {
			delete(nc, r)
			delete(ncrd, r)
		}
	}

	return &Vertex{
		gd:        gd,
		Nodecoord: ncrd,
		Nodecells: nc,
		Cellnodes: cn,
		dir:       0,
	}
}

func (v *Vertex) Nvert() int {
	return len(v.Nodecoord)
}

// Sub GridNode_clockwise()
// If _dir = 1 Then Exit Sub
// If _dir <> 0 Then Stop ' todo
// _dir = 1
// Dim lst1 As New Dictionary(Of Integer, Integer())
// For Each cn In _cellnodes
// 	'p1---p2   y
// 	' | c |    |
// 	'p0---p3   0---x
// 	Dim inA(3) As Integer
// 	inA(0) = cn.Value(0)
// 	inA(1) = cn.Value(2)
// 	inA(2) = cn.Value(3)
// 	inA(3) = cn.Value(1)
// 	lst1.Add(cn.Key, inA)
// Next
// _cellnodes = lst1
// End Sub

// Sub GridNode_counterclockwise()
// If _dir = -1 Then Exit Sub
// If _dir <> 0 Then Stop ' todo
// _dir = -1
// Dim lst1 As New Dictionary(Of Integer, Integer())
// For Each cn In _cellnodes
// 	'p3---p2   y
// 	' | c |    |
// 	'p0---p1   0---x
// 	Dim inA(3) As Integer
// 	inA(0) = cn.Value(0)
// 	inA(1) = cn.Value(1)
// 	inA(2) = cn.Value(3)
// 	inA(3) = cn.Value(2)
// 	lst1.Add(cn.Key, inA)
// Next
// _cellnodes = lst1
// End Sub
