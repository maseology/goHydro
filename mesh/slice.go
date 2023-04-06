package mesh

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmaths/spatial"
	"github.com/maseology/mmaths/vector"
	"github.com/maseology/mmio"
)

// Slice struct of a uniform grid
type Slice struct {
	Nodes          [][]float64
	Elements, NExr [][]int
	Name           string
	xys            *spatial.XYsearch
	r              float64 // "radius" set to longest edge
}

// // NewSlice constructs a basic fem mesh
// func NewSlice(nam string, nr, nc int, UniformCellSize float64) *Slice {
// 	var gd Slice
// 	gd.Name = nam
// 	gd.Nrow, gd.Ncol, gd.Nact = nr, nc, nr*nc
// 	gd.Cwidth = UniformCellSize
// 	gd.Sactives = make([]int, gd.Nact)
// 	gd.act = make(map[int]bool, gd.Nact)
// 	for i := 0; i < gd.Nact; i++ {
// 		gd.Sactives[i] = i
// 		gd.act[i] = true
// 	}
// 	gd.Coord = make(map[int]mmaths.Point, gd.Nact)
// 	cid := 0
// 	for i := 0; i < gd.Nrow; i++ {
// 		for j := 0; j < gd.Ncol; j++ {
// 			p := mmaths.Point{X: gd.Eorig + UniformCellSize*(float64(j)+0.5), Y: gd.Norig - UniformCellSize*(float64(i)+0.5)}
// 			gd.Coord[cid] = p
// 			cid++
// 		}
// 	}
// 	return &gd
// }

// ReadAlgomesh imports a Algomesh grids in .ah2 or .ah3
func ReadAlgomesh(fp string, prnt bool) (*Slice, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("ReadGDEF: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	line, _, err := reader.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("ReadTextLines: %v", err)
	}
	nn, err := strconv.ParseInt(string(line), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ReadTextLines: %v", err)
	}

	f64 := func(s string) float64 {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic("ReadAlgomesh f64")
		}
		return f
	}
	i64 := func(s string) int {
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			panic("ReadAlgomesh i64")
		}
		if i < 0 {
			panic("wtf")
		}
		return int(i)
	}

	nds := make([][]float64, nn)
	for i := 0; i < int(nn); i++ {
		line, _, err := reader.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("ReadTextLines: %v", err)
		}
		sp := strings.Split(string(line), " ")
		switch len(sp) {
		case 3:
			nds[i] = []float64{f64(sp[0]), f64(sp[1]), f64(sp[2])}
		case 2:
			nds[i] = []float64{f64(sp[0]), f64(sp[1])}
		default:
			panic("ReadAlgomesh 1")
		}
	}

	line, _, err = reader.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("ReadTextLines: %v", err)
	}
	ne, err := strconv.ParseInt(string(line), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ReadTextLines: %v", err)
	}

	els := make([][]int, ne)
	ex := 0.
	checkLngest := func(ni []int) {
		chk := func(p0, p1 []float64) {
			d2 := math.Pow(p0[0]-p1[0], 2) + math.Pow(p0[1]-p1[1], 2)
			if d2 > ex {
				ex = d2
			}
		}
		chk(nds[ni[0]], nds[ni[1]])
		chk(nds[ni[0]], nds[ni[2]])
		chk(nds[ni[2]], nds[ni[1]])
	}
	for i := 0; i < int(ne); i++ {
		line, _, err := reader.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("ReadTextLines: %v", err)
		}
		sp := strings.Split(string(line), " ")
		if len(sp) != 3 {
			panic("ReadAlgomesh assumes triagular meshes for now")
		}
		els[i] = []int{i64(sp[0]) - 1, i64(sp[1]) - 1, i64(sp[2]) - 1}
		checkLngest(els[i])
	}

	if _, _, err := reader.ReadLine(); err != io.EOF {
		panic("ReadAlgomesh 3")
	}

	sl := Slice{Name: mmio.FileName(fp, false), Nodes: nds, Elements: els, r: math.Sqrt(ex)}
	sl.buildXR()

	if prnt {
		fmt.Printf("\n opened %s\n  nElements %d\n  nNodes %d\n  radius %.3f\n\n", fp, len(sl.Elements), len(sl.Nodes), sl.r)
	}
	return &sl, nil
}

func (sl *Slice) SaveAs(fp string) error {
	switch mmio.GetExtension(fp) {
	case ".ah2", ".ah3":
		t, err := mmio.NewTXTwriter(fp)
		if err != nil {
			return fmt.Errorf(" Slice.SaveAs: %v", err)
		}
		defer t.Close()

		t.WriteLine(fmt.Sprintf("%d", len(sl.Nodes)))
		for _, xy := range sl.Nodes {
			t.WriteLine(fmt.Sprintf("%f %f", xy[0], xy[1]))
		}

		t.WriteLine(fmt.Sprintf("%d", len(sl.Elements)))
		for _, nids := range sl.Elements {
			if len(nids) != 3 {
				panic(" Slice.SaveAs: TODO")
			}
			t.WriteLine(fmt.Sprintf("%d %d %d", nids[0]+1, nids[1]+1, nids[2]+1))
		}

		return nil
	default:
		return fmt.Errorf(" Unknown format: %s", fp)
	}
}

func (sl *Slice) buildXR() {
	sl.NExr = make([][]int, len(sl.Nodes))
	for eid, nids := range sl.Elements {
		for _, nid := range nids {
			sl.NExr[nid] = append(sl.NExr[nid], eid)
		}
	}
}

func (sl *Slice) Extent() *mmaths.Extent {
	var x mmaths.Extent
	x.New(sl.Nodes)
	return &x
}

func (sl *Slice) SurroundingElements(eid, levels int) []int {
	if levels == 0 {
		return []int{eid}
	}
	k := 0
	m := map[int]int{eid: 1}
	ly := []int{eid}
loop:
	for {
		lynew := []int{}
		for _, eid2 := range ly {
			for _, nid := range sl.Elements[eid2] {
				for _, beid := range sl.NExr[nid] {
					m[beid]++
					lynew = append(lynew, beid)
				}
			}
		}
		if k == levels-1 {
			break loop
		}
		copy(ly, lynew)
		k++
	}
	return func(m map[int]int) []int {
		i := make([]int, 0, len(m))
		for k := range m {
			i = append(i, k)
		}
		return i
	}(m)
}

func (sl *Slice) ElementToTriangle(eid int) *mmaths.Triangle {
	var tri mmaths.Triangle
	tri.New(sl.Nodes[sl.Elements[eid][0]], sl.Nodes[sl.Elements[eid][1]], sl.Nodes[sl.Elements[eid][2]])
	return &tri
}

func (sl *Slice) ElementPerimeter(eid int) [][]float64 {
	return [][]float64{sl.Nodes[sl.Elements[eid][0]], sl.Nodes[sl.Elements[eid][1]], sl.Nodes[sl.Elements[eid][2]], sl.Nodes[sl.Elements[eid][0]]}
}

func (sl *Slice) PointToElementID(x, y float64) int {
	if sl.xys == nil {
		var xys spatial.XYsearch
		xys.New(sl.Nodes)
		sl.xys = &xys
	}
	nids, _ := sl.xys.ClosestIDs([]float64{x, y}, sl.r)
	if len(nids) == 0 {
		return -1
	}
	// closest node
	for _, eid := range sl.NExr[nids[0]] {
		tri := sl.ElementToTriangle(eid)
		if tri.Contains(x, y) {
			return eid
		}
	}
	return -1
}

func (sl *Slice) LineToNodeIDs(x0, y0, x1, y1 float64) []int {
	ln := mmaths.LineSegment{P0: mmaths.Point{x0, y0, 0., 0.}, P1: mmaths.Point{x1, y1, 0., 0.}}
	ln.Build()
	ext := mmaths.Extent{math.Min(x0, x1) - sl.r, math.Max(x0, x1) + sl.r, math.Min(y0, y1) - sl.r, math.Max(y0, y1) + sl.r}
	nids := []int{}
	for i, xy := range sl.Nodes {
		p := &mmaths.Point{xy[0], xy[1], 0., 0.}
		if ext.Contains(p) {
			if ln.Intersects(p, sl.r) {
				nids = append(nids, i)
			}
		}
	}
	return nids
}

func (sl *Slice) LineToElementIDs(x0, y0, x1, y1 float64) []int {
	nids := sl.LineToNodeIDs(x0, y0, x1, y1)
	if len(nids) == 0 {
		return nil
	}
	intAbs := func(i int) int {
		if i < 0 {
			return -i
		}
		return i
	}
	m := make(map[int]int)
	p0, p1 := [3]float64{x0, y0, 0.}, [3]float64{x1, y1, 0.}
	for _, nid := range nids {
		for _, eid := range sl.NExr[nid] {
			x := sl.Elements[eid]
			c := 0
			for _, ni := range x {
				cp := ((x1-x0)*(sl.Nodes[ni][1]-y0) - (y1-y0)*(sl.Nodes[ni][0]-x0)) // cross product
				if cp > 0. {
					c++
				} else if cp < 0. {
					c--
				}
			}
			if intAbs(c) == 3 {
				continue // check to see if segment bisects triangle nodes
			}

			tri := sl.ElementToTriangle(eid)
			if tri.Contains(x0, y0) || tri.Contains(x1, y1) {
				m[eid]++ // vertex in triangle
			} else {
				for _, ni := range x {
					vp := [3]float64{sl.Nodes[ni][0], sl.Nodes[ni][1], 0.}
					_, t, _ := vector.PointToLine(vp, p0, p1)
					if t > 0. && t < 1. {
						m[eid]++ // point projecting onto segment
						break
					}
				}
			}
		}
	}
	eids := make([]int, 0, len(m))
	for k := range m {
		eids = append(eids, k)
	}

	return eids
}
