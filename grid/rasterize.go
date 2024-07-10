package grid

import (
	"math"
	"sort"

	"github.com/maseology/mmaths"
)

type PolygonRasterizer struct {
	gd       *Definition
	i        [][]int
	in       [][]bool
	v        [][]float64
	pln      *mmaths.Polyline
	rcn, rcx []int
	ext      *mmaths.Extent
}

func (pr *PolygonRasterizer) InteriorCellIDs(GD *Definition, Polygon [][]float64) ([]int, int) {
	pr.gd = GD
	if GD.Rotation != 0. {
		panic("TODO: PolygonRasterizer.New  rotated grid") // ' may need some code changes, best to work this on an rotated coordinate system, ie. rotate polygon vertices. need changes to getExtents() and other functions such as: CellCentroid, etc.
	}

	sqdist := func(p1, p2 []float64) float64 {
		return math.Pow(p1[0]-p2[0], 2.0) + math.Pow(p1[1]-p2[1], 2.0)
	}

	if sqdist(Polygon[0], Polygon[len(Polygon)-1]) < 1e-5 {
		pr.v = Polygon
		pr.pln = &mmaths.Polyline{S: Polygon}
	} else {
		pr.v = make([][]float64, len(Polygon)+1)
		pr.pln = &mmaths.Polyline{S: make([][]float64, len(Polygon)+1)}
		for i, v := range Polygon {
			pr.v[i] = v
			pr.pln.S[i] = v
		}
		pr.v[len(Polygon)] = Polygon[0]
		pr.pln.S[len(Polygon)] = Polygon[0]
	}
	pr.getExtents()
	pr.i, pr.in = make([][]int, GD.Nrow), make([][]bool, GD.Nrow)
	for i := range GD.Nrow {
		pr.i[i] = make([]int, GD.Ncol)
		pr.in[i] = make([]bool, GD.Ncol)
	}

	pr.raterize()

	var cids []int
	for i := pr.rcn[0]; i <= pr.rcx[0]; i++ {
		for j := pr.rcn[1]; j <= pr.rcx[1]; j++ {
			if pr.in[i][j] {
				cids = append(cids, pr.gd.CellID(i, j))
			}
		}
	}
	return cids, len(cids)
}

func (pr *PolygonRasterizer) getExtents() {
	pr.ext = &mmaths.Extent{}
	pr.ext.New(pr.v)
	rul, cul := pr.gd.PointToRowCol(pr.ext.Xn, pr.ext.Yx) // UL
	rlr, clr := pr.gd.PointToRowCol(pr.ext.Xx, pr.ext.Yn) // LR
	keepwithinbounds := func(r, c int) []int {
		if r < 0 {
			r = 0
		}
		if c < 0 {
			c = 0
		}
		if r >= pr.gd.Nrow {
			r = pr.gd.Nrow - 1
		}
		if c >= pr.gd.Ncol {
			c = pr.gd.Ncol - 1
		}
		return []int{r, c}
	}
	pr.rcn = keepwithinbounds(rul, cul)
	pr.rcx = keepwithinbounds(rlr, clr)
}

func (pr *PolygonRasterizer) raterize() {
	// based on PNPoly by W. Randolph Franklin: http://www.ecse.rpi.edu/~wrf/Research/Short_Notes/pnpoly.html
	for range pr.v {
		for i := pr.rcn[0]; i <= pr.rcx[0]; i++ {
			for j := pr.rcn[1]; j <= pr.rcx[1]; j++ {
				if pr.i[i][j] == 3 {
					continue // evaluated or captured by laser in both directions
				}
				if pr.i[i][j] == 0 {
					pr.i[i][j] = 3 // 0: unevaluated; 1: lasered horizontally; 2: lasered vertically; 3: evaluated/lasered in both directions
					blIn := false
					cid := pr.gd.CellID(i, j)
					ccent := pr.gd.CellCentroid(cid)
					if !pr.ext.Contains(mmaths.Point{X: ccent[0], Y: ccent[1]}) {
						goto skipPnP
					}

					blIn = mmaths.PnPoly(pr.v, ccent)

				skipPnP:
					if blIn {
						pr.in[i][j] = true
					}
					pr.laserH(mmaths.Polyline{S: [][]float64{{pr.ext.Xn, ccent[1]}, {pr.ext.Xx, ccent[1]}}}, ccent[0], i, j)
					pr.laserV(mmaths.Polyline{S: [][]float64{{ccent[0], pr.ext.Yn}, {ccent[0], pr.ext.Yx}}}, ccent[1], i, j)
				}
			}
		}
	}
}

// laserH horizontal laser from cell centroid to polygon boundary
func (pr *PolygonRasterizer) laserH(laser mmaths.Polyline, cx float64, ir, jc int) {
	intXYh := pr.pln.Intersections(&laser)
	switch len(intXYh) {
	case 0:
		for j := pr.rcn[1]; j <= pr.rcx[1]; j++ {
			if pr.i[ir][j] > 0 || j == jc {
				continue
			}
			pr.i[ir][j] += 1
			pr.in[ir][j] = pr.in[ir][jc]
		}
	case 1:
		if cx == intXYh[0][0] {
			return
		}
		blL := cx < intXYh[0][0]
		if !pr.in[ir][jc] {
			blL = !blL
		}
		for j := pr.rcn[1]; j <= pr.rcx[1]; j++ {
			if pr.i[ir][j] > 0 || j == jc {
				continue
			}
			pr.i[ir][j]++
			cid := pr.gd.CellID(ir, j)
			if pr.gd.Coord[cid].X < intXYh[0][0] {
				pr.in[ir][j] = blL
			}
		}
	default:
		jsv := 0
		xs := make([]float64, len(intXYh))
		for i, v := range intXYh {
			xs[i] = v[0]
		}
		sort.Float64s(xs)
		for _, x := range xs {
			if cx == x {
				return
			}
			if cx > x {
				jsv++
			}
		}
		switch jsv {
		case 0:
			_, ctmp := pr.gd.PointToRowCol(xs[0], 0.)
			if ctmp == -1 {
				ctmp++
			}
			if ctmp >= pr.gd.Ncol {
				ctmp = pr.gd.Ncol - 1
			}
			cid := pr.gd.CellID(ir, ctmp)
			if pr.gd.Coord[cid].X > xs[0] {
				ctmp--
			}
			for j := pr.rcn[1]; j < ctmp; j++ {
				if pr.i[ir][j] > 0 || j == jc {
					continue
				}
				pr.i[ir][j]++
				pr.in[ir][j] = pr.in[ir][jc]
			}
		case len(xs):
			_, ctmp := pr.gd.PointToRowCol(xs[len(xs)-1], 0.)
			if ctmp == -1 {
				ctmp++
			}
			if ctmp >= pr.gd.Ncol {
				ctmp = pr.gd.Ncol - 1
			}
			cid := pr.gd.CellID(ir, ctmp)
			if pr.gd.Coord[cid].X < xs[len(xs)-1] {
				ctmp++
			}
			for j := ctmp; j < pr.rcx[1]; j++ {
				if pr.i[ir][j] > 0 || j == jc {
					continue
				}
				pr.i[ir][j]++
				pr.in[ir][j] = pr.in[ir][jc]
			}
		default:
			_, cL := pr.gd.PointToRowCol(xs[jsv-1], 0.)
			if cL == -1 {
				cL++
			}
			cidL := pr.gd.CellID(ir, cL)
			if pr.gd.Coord[cidL].X < xs[jsv-1] {
				cL++
			}
			_, cR := pr.gd.PointToRowCol(xs[jsv], 0.)
			if cR == pr.gd.Ncol {
				cR--
			}
			cidR := pr.gd.CellID(ir, cR)
			if pr.gd.Coord[cidR].X > xs[jsv] {
				cR--
			}
			for j := cL; j < cR; j++ {
				if pr.i[ir][j] > 0 || j == jc {
					continue
				}
				pr.i[ir][j]++
				pr.in[ir][j] = pr.in[ir][jc]
			}
		}
	}
}

// vertical laser from cell centroid to polygon boundary
func (pr *PolygonRasterizer) laserV(laser mmaths.Polyline, cy float64, ir, jc int) {
	intXYv := pr.pln.Intersections(&laser)
	switch len(intXYv) {
	case 0:
		for i := pr.rcn[0]; i <= pr.rcx[0]; i++ {
			if pr.i[i][jc] > 0 || i == ir {
				continue
			}
			pr.i[i][jc] += 2
			pr.in[i][jc] = pr.in[ir][jc]
		}
	case 1:
		if cy == intXYv[0][1] {
			return
		}
		blU := cy > intXYv[0][1]
		if !pr.in[ir][jc] {
			blU = !blU
		}
		for i := pr.rcn[0]; i <= pr.rcx[0]; i++ {
			if pr.i[i][jc] > 0 || i == ir {
				continue
			}
			pr.i[i][jc] += 2
			cid := pr.gd.CellID(i, jc)
			if pr.gd.Coord[cid].Y > intXYv[0][1] {
				pr.in[i][jc] = blU
			}
		}
	default:
		isv := 0
		ys := make([]float64, len(intXYv))
		for i, v := range intXYv {
			ys[i] = v[0]
		}
		sort.Float64s(ys)
		for _, y := range ys {
			if cy == y {
				return
			}
			if cy < y {
				isv++
			}
		}
		switch isv {
		case 0:
			rtmp, _ := pr.gd.PointToRowCol(0., ys[0])
			if rtmp == -1 {
				rtmp++
			}
			if rtmp >= pr.gd.Nrow {
				rtmp = pr.gd.Nrow - 1
			}
			cid := pr.gd.CellID(rtmp, jc)
			if pr.gd.Coord[cid].Y < ys[0] {
				rtmp--
			}
			for i := pr.rcn[0]; i < rtmp; i++ {
				if pr.i[i][jc] > 0 || i == ir {
					continue
				}
				pr.i[i][jc] += 2
				pr.in[i][jc] = pr.in[ir][jc]
			}
		case len(ys):
			rtmp, _ := pr.gd.PointToRowCol(0., ys[len(ys)-1])
			if rtmp == -1 {
				rtmp++
			}
			if rtmp >= pr.gd.Nrow {
				rtmp = pr.gd.Nrow - 1
			}
			cid := pr.gd.CellID(rtmp, jc)
			if pr.gd.Coord[cid].Y > ys[len(ys)-1] {
				rtmp++
			}
			for i := rtmp; i < pr.rcx[0]; i++ {
				if pr.i[i][jc] > 0 || i == ir {
					continue
				}
				pr.i[i][jc] += 2
				pr.in[i][jc] = pr.in[ir][jc]
			}
		default:
			rT, _ := pr.gd.PointToRowCol(0., ys[isv-1])
			if rT == -1 {
				rT++
			}
			cidT := pr.gd.CellID(rT, jc)
			if pr.gd.Coord[cidT].X < ys[isv-1] {
				rT++
			}
			rB, _ := pr.gd.PointToRowCol(0., ys[isv])
			if rB == pr.gd.Ncol {
				rB--
			}
			cidB := pr.gd.CellID(rB, jc)
			if pr.gd.Coord[cidB].Y < ys[isv] {
				rB--
			}
			for i := rT; i < rB; i++ {
				if pr.i[i][jc] > 0 || i == ir {
					continue
				}
				pr.i[i][jc] += 2
				pr.in[i][jc] = pr.in[ir][jc]
			}
		}
	}
}
