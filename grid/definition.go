package grid

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

// Definition struct
type Definition struct {
	Coord                          map[int]mmaths.Point
	act                            map[int]bool
	Sactives                       []int   // a sorted slice of active cell IDs
	Eorig, Norig, Cwidth, Rotation float64 // Xul; Yul; cell width, grid rotation about ULorigin
	Nrow, Ncol, Nact               int
	Name                           string
}

// NewDefinition constructs a basic grid definition
func NewDefinition(nam string, nr, nc int, UniformCellSize float64) *Definition {
	var gd Definition
	gd.Name = nam
	gd.Nrow, gd.Ncol, gd.Nact = nr, nc, nr*nc
	gd.Cwidth = UniformCellSize
	gd.Sactives = make([]int, gd.Nact)
	gd.act = make(map[int]bool, gd.Nact)
	for i := 0; i < gd.Nact; i++ {
		gd.Sactives[i] = i
		gd.act[i] = true
	}
	gd.Coord = make(map[int]mmaths.Point, gd.Nact)
	cid := 0
	for i := 0; i < gd.Nrow; i++ {
		for j := 0; j < gd.Ncol; j++ {
			p := mmaths.Point{X: gd.Eorig + gd.Cwidth*(float64(j)+0.5), Y: gd.Norig - gd.Cwidth*(float64(i)+0.5)}
			gd.Coord[cid] = p
			cid++
		}
	}
	return &gd
}

// ReadGDEF imports a grid definition file
func ReadGDEF(fp string, print bool) (*Definition, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("ReadGDEF: %v", err)
	}
	defer file.Close()

	reader, a, l := bufio.NewReader(file), make([]string, 6), 0
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("ReadTextLines: %v", err)
		}
		a[l] = string(line)
		l++
		if l == 6 {
			break
		}
	}

	parseHeader := func(a []string, print bool) (Definition, error) {
		stErr, uni := make([]string, 0), false
		errfunc := func(v string, err error) {
			stErr = append(stErr, fmt.Sprintf("   failed to read '%v': %v", v, err))
		}

		oe, err := strconv.ParseFloat(a[0], 64)
		if err != nil {
			errfunc("OE", err)
		}
		on, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			errfunc("ON", err)
		}
		rot, err := strconv.ParseFloat(a[2], 64)
		if err != nil {
			errfunc("ROT", err)
		}
		nr, err := strconv.ParseInt(a[3], 10, 32)
		if err != nil {
			errfunc("NR", err)
		}
		nc, err := strconv.ParseInt(a[4], 10, 32)
		if err != nil {
			errfunc("NC", err)
		}
		cs, err := strconv.ParseFloat(a[5], 64)
		if err != nil {
			if a[5][0] == 85 { // 85 = acsii code for 'U'
				uni = true
			} else {
				errfunc("CS", err)
			}
			cs, err = strconv.ParseFloat(a[5][1:len(a[5])], 64)
			if err != nil {
				errfunc("CS", err)
			}
		} else {
			stErr = append(stErr, " *** Fatal error: ReadGDEF.parseHeader: non-uniform grids currently not supported ***")
		}

		// error handling
		if len(stErr) > 0 {
			return Definition{}, fmt.Errorf("fatal error(s): ReadGDEF.parseHeader:\n%s", strings.Join(stErr, "\n"))
		}

		gd := Definition{Eorig: oe, Norig: on, Rotation: rot, Cwidth: cs, Nrow: int(nr), Ncol: int(nc)}
		if print {
			fmt.Printf(" xul\t\t%.1f\n", oe)
			fmt.Printf(" yul\t\t%.1f\n", on)
			fmt.Printf(" rotation\t%f\n", rot)
			fmt.Printf(" nrows\t\t%d\n", nr)
			fmt.Printf(" ncols\t\t%d\n", nc)
			fmt.Printf(" cell size\t%.3f\n", cs)
			fmt.Printf(" is uniform:\t%t\n", uni)
		}

		return gd, nil
	}
	gd, err := parseHeader(a, print)
	if err != nil {
		return nil, err
	}
	if gd.Rotation != 0. {
		return nil, fmt.Errorf("ReadGDEF error: rotation no yet supported")
	}

	nc := gd.Nrow * gd.Ncol
	cn, cx := 0, nc
	if nc%8 != 0 {
		nc += 8 - nc%8 // add padding
	}
	nc /= 8

	b1 := make([]byte, nc)
	if err := binary.Read(reader, binary.LittleEndian, b1); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("fatal error: read actives failed: %v", err)
		}
		if print {
			fmt.Printf(" (no active cells)\n")
		}
		gd.Sactives = make([]int, cx)
		gd.act = make(map[int]bool, cx)
		gd.Nact = cx
		for i := 0; i < cx; i++ {
			gd.Sactives[i] = i
			gd.act[i] = true
		}
		gd.Coord = make(map[int]mmaths.Point, cx)
		cid := 0
		for i := 0; i < gd.Nrow; i++ {
			for j := 0; j < gd.Ncol; j++ {
				p := mmaths.Point{X: gd.Eorig + gd.Cwidth*(float64(j)+0.5), Y: gd.Norig - gd.Cwidth*(float64(i)+0.5)}
				gd.Coord[cid] = p
				cid++
			}
		}
	} else { // active cells
		t := make([]byte, 1)
		if v, _ := reader.Read(t); v != 0 {
			return nil, fmt.Errorf("fatal error: EOF not reached when expected")
		}
		gd.Sactives = []int{}
		gd.act = make(map[int]bool, cx)
		for _, b := range b1 {
			for i := uint(0); i < 8; i++ {
				if b&(1<<i)>>i == 1 {
					gd.Sactives = append(gd.Sactives, cn)
					gd.act[cn] = true
					gd.Nact++
				}
				cn++
				if cn >= cx {
					break
				}
			}
		}
		if cn != cx {
			return nil, fmt.Errorf("fatal error(s): ReadGDEF:\n   number of cells found (%d) not equal to total (%d): %v", cn, cx, err)
		}
		if gd.Nact > 0 && print {
			fmt.Printf(" %s active cells\n", mmio.Thousands(int64(gd.Nact))) //11,118,568
		}
		gd.Coord = make(map[int]mmaths.Point, gd.Nact)
		cid := 0
		for i := 0; i < gd.Nrow; i++ {
			for j := 0; j < gd.Ncol; j++ {
				if _, ok := gd.act[cid]; ok {
					p := mmaths.Point{X: gd.Eorig + gd.Cwidth*(float64(j)+0.5), Y: gd.Norig - gd.Cwidth*(float64(i)+0.5)}
					gd.Coord[cid] = p
				}
				cid++
			}
		}
	}
	if print {
		fmt.Println()
	}
	return &gd, nil
}

func ReadHdr(fp string) (*Definition, float64, error) {
	sa, err := mmio.ReadTextLines(fp)
	if err != nil {
		return nil, 0, err
	}
	var nr, nc int
	var xorig, yorig, cs, nd float64
	var cntrd, ll bool

	for _, s := range sa {
		sp := strings.Split(mmio.RemoveWhiteSpaces(s), " ")
		if len(sp) != 2 {
			break
		}
		lwr := strings.ToLower(sp[0])
		switch lwr {
		case "ncols":
			nc, _ = strconv.Atoi(sp[1])
		case "nrows":
			nr, _ = strconv.Atoi(sp[1])
		case "xllcorner":
			xorig, _ = strconv.ParseFloat(sp[1], 64)
			ll = true
		case "yllcorner":
			yorig, _ = strconv.ParseFloat(sp[1], 64)
		case "ulxmap":
			xorig, _ = strconv.ParseFloat(sp[1], 64)
			cntrd = true
		case "ulymap":
			yorig, _ = strconv.ParseFloat(sp[1], 64)
		case "cellsize", "xdim":
			cs, _ = strconv.ParseFloat(sp[1], 64)
		case "nodata_value", "nodata":
			nd, _ = strconv.ParseFloat(sp[1], 64)
		}
	}

	if nr == 0 {
		return nil, 0, fmt.Errorf("grid definition read error")
	}

	if ll { // lower-left to upper-left origin
		yorig += float64(nr) * cs
	}
	if cntrd { // centroidal
		xorig -= cs / 2
		yorig += cs / 2
	}

	return &Definition{
		Name:   mmio.FileName(fp, false),
		Nrow:   nr,
		Ncol:   nc,
		Nact:   nr * nc,
		Eorig:  xorig,
		Norig:  yorig,
		Cwidth: cs,
	}, nd, nil
}

// IsActive returns whether a cell ID is of an active cell
func (gd *Definition) IsActive(cid int) bool {
	return gd.act[cid]
}

// // Actives returns a slice of active cell IDs
// func (gd *Definition) Actives() []int {
// 	out, i := make([]int, gd.Nact), 0
// 	for k, v := range gd.act {
// 		if v {
// 			out[i] = k
// 			i++
// 		}
// 	}
// 	return out
// }

// // Sactives returns a SORTED slice of active cell IDs
// func (gd *Definition) Sactives() []int {
// 	a := gd.Actives()
// 	sort.Ints(a)
// 	return a
// }

// // Origin returns the upper-left corner of the grid's extent
// func (gd *Definition) Origin() (float64, float64) {
// 	return gd.Eorig, gd.Norig
// }

// RowCol returns row and column index for a given cell ID
func (gd *Definition) RowCol(cid int) (row, col int) {
	if cid < 0 || cid > gd.Nrow*gd.Ncol {
		log.Fatalf("Definition.RowCol error: invalid cell ID: %d", cid)
	}
	row = int(float64(cid) / float64(gd.Ncol))
	col = cid - row*gd.Ncol
	return
}

// CellID returns cell ID for a given row and column index
func (gd *Definition) CellID(row, col int) int {
	if row < 0 || row >= gd.Nrow || col < 0 || col >= gd.Ncol {
		// log.Fatalf("Definition.CellID error: invalid [row,col]: [%d,%d]", row, col)
		return -1
	}
	return row*gd.Ncol + col
}

// Ncells returns the count of grid cells
func (gd *Definition) Ncells() int {
	return gd.Ncol * gd.Nrow
}

// CellArea returns the (uniform) area of the grid cells
func (gd *Definition) CellArea() float64 {
	return gd.Cwidth * gd.Cwidth
}

func (gd *Definition) CellCentroid(cid int) []float64 {
	r, c := gd.RowCol(cid)
	return []float64{gd.Eorig + (float64(c)+.5)*gd.Cwidth, gd.Norig - (float64(r)+.5)*gd.Cwidth}
}

func (gd *Definition) CellPerimeter(cid int) [][]float64 {
	cw2 := gd.Cwidth / 2
	ctrd := gd.CellCentroid(cid)
	return [][]float64{{ctrd[0] + cw2, ctrd[1] + cw2}, {ctrd[0] + cw2, ctrd[1] - cw2}, {ctrd[0] - cw2, ctrd[1] - cw2}, {ctrd[0] - cw2, ctrd[1] + cw2}, {ctrd[0] + cw2, ctrd[1] + cw2}}
}

// CellIndexXR returns a mapping of cell id to an array index
func (gd *Definition) CellIndexXR() map[int]int {
	m := make(map[int]int, len(gd.Sactives))
	for i, c := range gd.Sactives {
		m[c] = i
	}
	return m
}

// PointToCellID returns the cell id that contains the xy coordinates
func (gd *Definition) PointToCellID(x, y float64) int {
	return gd.CellID(gd.PointToRowCol(x, y))
}

// PointToRowCol returns the row and column grid cell that contains the xy coordinates
func (gd *Definition) PointToRowCol(x, y float64) (row, col int) {
	row = -1
	col = -1
	if gd.Rotation != 0. {
		log.Fatalf(" Definition.PointToRowCol todo")
	}
	for {
		row++
		if gd.Norig-float64(row+1)*gd.Cwidth <= y {
			break
		}
	}
	for {
		col++
		if gd.Eorig+float64(col+1)*gd.Cwidth >= x {
			break
		}
	}
	return
}

// ConatainsPoint returns whether a point exists within a grid definition, with a specified buffer
func (gd *Definition) ConatainsPoint(x, y, buf float64) bool {
	if x < gd.Eorig-buf {
		return false
	}
	if x > gd.Eorig+float64(gd.Ncol)*gd.Cwidth+buf {
		return false
	}
	if y > gd.Norig+buf {
		return false
	}
	if y < gd.Norig-float64(gd.Nrow)*gd.Cwidth-buf {
		return false
	}
	return true
}

func (gd *Definition) LineToCellIDs(polyline [][]float64, thin bool) []int {
	kv := make(map[int]bool)

	if thin {
		panic("GD.LineToCellIDs thin: TODO")
		//       ' see (for example) Lindsay 2016 The practice of DEM stream burning revisited
		//       'Dim dbnc = _nc.Clone, dbec = _ec.Clone
		//       'Array.Sort(dbnc)
		//       'Dim ct As New Stats.BinaryTree(_ec), rt As New Stats.BinaryTree(dbnc)
		//       For k = 0 To Features.Count - 2
		//           Dim p0 = Features(k).Clone, p1 = Features(k + 1).Clone
		//           If _rotation <> 0.0 Then
		//               p0.Rotate(_rotation, True)
		//               p1.Rotate(_rotation, True)
		//           End If
		//           Dim ln1 As New Shapes.Line(p0, p1), lstIntersect As New List(Of Cartesian.XY)
		//           'Dim cid0 = ct.IndexOf(p0.X), rid0 = dbnc.length - rt.IndexOf(p0.Y) - 1, cid1 = ct.IndexOf(p1.X), rid1 = dbnc.length - rt.IndexOf(p1.Y) - 1
		//           Dim rc0 = Me.PointToRowCol(p0), rc1 = Me.PointToRowCol(p1)
		//           If IsNothing(rc0) Or IsNothing(rc1) Then Continue For
		//           Dim cid0 = rc0.Col, rid0 = rc0.Row, cid1 = rc1.Col, rid1 = rc1.Row
		//           If cid0 < 0 Then cid0 = 0
		//           If cid0 > _ec.Count - 1 Then cid0 = _ec.Count - 1
		//           If cid1 < 0 Then cid1 = 0
		//           If cid1 > _ec.Count - 1 Then cid1 = _ec.Count - 1
		//           If rid0 < 0 Then rid0 = 0
		//           If rid0 > _nc.Count - 1 Then rid0 = _nc.Count - 1
		//           If rid1 < 0 Then rid1 = 0
		//           If rid1 > _nc.Count - 1 Then rid1 = _nc.Count - 1
		//
		//           If cid0 <> cid1 Then
		//               Dim j0 As Integer, j1 As Integer
		//               If cid1 > cid0 Then
		//                   j0 = cid0 + 1
		//                   j1 = cid1
		//               Else
		//                   j0 = cid1 + 1
		//                   j1 = cid0
		//               End If
		//               For j = j0 To j1
		//                   Dim lni = ln1.IntersectionX(_ec(j))
		//                   If Not IsNothing(lni) Then lstIntersect.Add(lni)
		//               Next
		//           End If
		//           If rid0 <> rid1 Then
		//               Dim i0 As Integer, i1 As Integer
		//               If rid1 > rid0 Then
		//                   i0 = rid0 + 1
		//                   i1 = rid1
		//               Else
		//                   i0 = rid1 + 1
		//                   i1 = rid0
		//               End If
		//               For i = i0 To i1
		//                   'lstInt.Add(ln1.IntersectionY(dbnc(i)))
		//                   Dim lni = ln1.IntersectionY(_nc(i))
		//                   If Not IsNothing(lni) Then lstIntersect.Add(lni)
		//               Next
		//           End If
		//
		//           ' sort intersections along polyline
		//           Dim dicX As New Dictionary(Of Integer, Double), dicY As New Dictionary(Of Integer, Double)
		//           For Each xy1 In lstIntersect
		//               dicX.Add(dicX.Count, xy1.X)
		//               dicY.Add(dicY.Count, xy1.Y)
		//           Next
		//           Dim dy = Math.Abs(p0.Y - p1.Y), dx = Math.Abs(p0.X - p1.X)
		//           If p0.Y = p1.Y Or dx > dy Then
		//               If dicX.Count > 1 Then dicX = Stats.SortArray(dicX, p0.X > p1.X)
		//               lstIntersect = New List(Of Cartesian.XY)
		//               For Each v In dicX
		//                   lstIntersect.Add(New Cartesian.XY(v.Value, dicY(v.Key)))
		//               Next
		//           Else 'If p0.X = p1.X Then
		//               If dicY.Count > 1 Then dicY = Stats.SortArray(dicY, p0.Y > p1.Y)
		//               lstIntersect = New List(Of Cartesian.XY)
		//               For Each v In dicY
		//                   lstIntersect.Add(New Cartesian.XY(dicX(v.Key), v.Value))
		//               Next
		//           End If
		//
		//           'Using sw As New StreamWriter("M:\OWRC-RDRR\build\dem\observations\test\asdf.csv")
		//           '    sw.WriteLine("i,e,n")
		//           '    Dim cnt = 0
		//           '    For Each xy1 In lstIntersect
		//           '        sw.WriteLine("{0},{1},{2}", cnt, xy1.X, xy1.Y)
		//           '        cnt += 1
		//           '    Next
		//           'End Using
		//
		//           With Me
		//               For Each xy1 In lstIntersect
		//                   If _rotation <> 0.0 Then xy1.Rotate(_rotation, False)
		//                   lstOUT.Add(.PointToCellID(xy1))
		//               Next
		//           End With
		// 199:  Next
		//       'lstOUT.Insert(0, PointToCellID(Features.Last))
		//       'lstOUT.Add(PointToCellID(Features.Last))
	} else {
		for i := 0; i < len(polyline)-1; i++ {
			p0, p1 := polyline[i], polyline[i+1]
			for {
				c0, c1 := gd.PointToCellID(p0[0], p0[1]), gd.PointToCellID(p1[0], p1[1])
				if c0 == c1 {
					if i >= len(polyline)-2 {
						kv[c0] = true // last vertex
					}
					break
				} else {
					mx, my := gd.lineIntersection(p0, p1, 0.000001)
					mcid := gd.PointToCellID(mx, my)
					if mcid < 0 {
						break
					}
					kv[c0] = true
					p0 = []float64{mx, my}
				}
			}
		}
	}

	ks := make([]int, 0, len(kv))
	for k := range kv {
		ks = append(ks, k)
	}
	return ks
}

func (gd *Definition) lineIntersection(p0, p1 []float64, push float64) (x, y float64) {
	if gd.Rotation != 0 {
		panic("GD.lineIntersection TODO")
		// Dim oxy As New Cartesian.XY(_origin.X, _origin.Y)
		// pIN.Rotate(_rotation, oxy, True)
		// pOut.Rotate(_rotation, oxy, True)
	}

	var gl, ymid, xmid float64
	// vertical check
	if p0[1] > p1[1] { // downward
		gl := gd.Norig
		for i := 0; i < gd.Nrow; i++ {
			if p0[1] > gl && p1[1] < gl {
				goto exitVertical
			}
			gl -= gd.Cwidth
		}
		if p0[1] > gl && p1[1] < gl {
			goto exitVertical
		}
	} else { // upward
		gl = gd.Norig - float64(gd.Nrow)*gd.Cwidth
		for i := gd.Nrow - 1; i >= 0; i-- {
			if p0[1] < gl && p1[1] > gl {
				goto exitVertical
			}
			gl += gd.Cwidth
		}
		if p0[1] < gl && p1[1] > gl {
			goto exitVertical
		}
	}
	gl = -9999.
exitVertical:
	ymid = gl

	// horizontal
	if p0[0] > p1[0] {
		gl = gd.Eorig + float64(gd.Ncol)*gd.Cwidth
		for j := gd.Ncol - 1; j >= 0; j-- {
			if p0[0] > gl && p1[0] < gl {
				goto exitHorizontal
			}
			gl -= gd.Cwidth
		}
		if p0[0] > gl && p1[0] < gl {
			goto exitHorizontal
		}
	} else {
		gl = gd.Eorig
		for j := 0; j < gd.Ncol; j++ {
			if p0[0] < gl && p1[0] > gl {
				goto exitHorizontal
			}
			gl += gd.Cwidth
		}
		if p0[0] < gl && p1[0] > gl {
			goto exitHorizontal
		}
	}
	gl = -9999.
exitHorizontal:
	xmid = gl

	pNEW := []float64{-9999., -9999.}
	if xmid != -9999. && ymid != -9999. {
		// line crosses both vertical and horizontal gridlines
		dy0 := ymid - p0[1]
		dx0 := xmid - p0[0]
		if math.Abs(dx0) == math.Abs(dy0) { // must be uniform
			// line crosses corner
			pNEW[0] = xmid
			pNEW[1] = ymid
			goto exitIntersct
		} else {
			tang1 := math.Atan2(p0[1]-p1[1], p0[0]-p1[0])
			dy1 := dx0 * tang1
			dx1 := dy0 / tang1
			lv := math.Sqrt(dx0*dx0 + dy1*dy1)
			lh := math.Sqrt(dx1*dx1 + dy0*dy0)
			if lv > lh {
				xmid = -9999.
			} else {
				ymid = -9999.
			}
		}
	}

	if xmid == -9999. && ymid == -9999. {
		// line outside of grid extents or entirely within a cell
		return -9999., -9999.
	} else if xmid == -9999. {
		// line only crosses a horizontal gridline
		pNEW[1] = ymid
		dy0 := p0[1] - ymid
		dy1 := p0[1] - p1[1] - dy0
		dx0 := dy0 * (p1[0] - p0[0]) / (dy1 + dy0)
		xmid = p0[0] + dx0
		pNEW[0] = xmid
	} else if ymid == -9999. {
		// line only crosses a vertical gridline
		pNEW[0] = xmid
		dx0 := xmid - p0[0]
		dx1 := p1[0] - p0[0] - dx0
		dy0 := dx0 * (p0[1] - p1[1]) / (dx1 + dx0)
		ymid = p0[1] - dy0
		pNEW[1] = ymid
	}

exitIntersct:
	if push != 0.0 { // NOTE: push insures the next point stays off a grid line: >0 sends to next cell, <0 keeps within current cell, =0 remains on gridline
		sign := func(x float64) float64 {
			if x < 0. {
				return -1.
			}
			return 1.
		}
		pNEW[0] += push * sign(xmid-p0[0])
		pNEW[1] -= push * sign(p0[1]-ymid)
	}

	if gd.Rotation != 0 {
		panic("GD.lineIntersection TODO")
		// If _rotation <> 0.0 Then pNEW.Rotate(_rotation, New Cartesian.XY(_origin.X, _origin.Y))
	}
	return pNEW[0], pNEW[1]
}

// SurroundingCells returns the relative row,col given in units of cell width
func SurroundingCells(unitRadius int) [][]int {
	// Dim gd As New Grid.Definition("t1", 100, 100)
	// For i = 0 To 10
	// 	Dim sb As New Text.StringBuilder
	// 	For Each rc In gd.SurroundingCellSet(i)
	// 		sb.Append(String.Format("{0}{1},{2}{3},", "{", rc.Row, rc.Col, "}"))
	// 	Next
	// 	Console.WriteLine("{0}: {1}{2}{3}", i, "{", Left(sb.ToString, sb.ToString.Length - 1), "},")
	// Next
	d := map[int][][]int{
		0:  {{0, 0}},
		1:  {{-1, 0}, {0, -1}, {0, 0}, {0, 1}, {1, 0}},
		2:  {{-2, 0}, {-1, -1}, {-1, 0}, {-1, 1}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {1, -1}, {1, 0}, {1, 1}, {2, 0}},
		3:  {{-3, 0}, {-2, -2}, {-2, -1}, {-2, 0}, {-2, 1}, {-2, 2}, {-1, -2}, {-1, -1}, {-1, 0}, {-1, 1}, {-1, 2}, {0, -3}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {1, -2}, {1, -1}, {1, 0}, {1, 1}, {1, 2}, {2, -2}, {2, -1}, {2, 0}, {2, 1}, {2, 2}, {3, 0}},
		4:  {{-4, 0}, {-3, -2}, {-3, -1}, {-3, 0}, {-3, 1}, {-3, 2}, {-2, -3}, {-2, -2}, {-2, -1}, {-2, 0}, {-2, 1}, {-2, 2}, {-2, 3}, {-1, -3}, {-1, -2}, {-1, -1}, {-1, 0}, {-1, 1}, {-1, 2}, {-1, 3}, {0, -4}, {0, -3}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {1, -3}, {1, -2}, {1, -1}, {1, 0}, {1, 1}, {1, 2}, {1, 3}, {2, -3}, {2, -2}, {2, -1}, {2, 0}, {2, 1}, {2, 2}, {2, 3}, {3, -2}, {3, -1}, {3, 0}, {3, 1}, {3, 2}, {4, 0}},
		5:  {{-5, 0}, {-4, -3}, {-4, -2}, {-4, -1}, {-4, 0}, {-4, 1}, {-4, 2}, {-4, 3}, {-3, -4}, {-3, -3}, {-3, -2}, {-3, -1}, {-3, 0}, {-3, 1}, {-3, 2}, {-3, 3}, {-3, 4}, {-2, -4}, {-2, -3}, {-2, -2}, {-2, -1}, {-2, 0}, {-2, 1}, {-2, 2}, {-2, 3}, {-2, 4}, {-1, -4}, {-1, -3}, {-1, -2}, {-1, -1}, {-1, 0}, {-1, 1}, {-1, 2}, {-1, 3}, {-1, 4}, {0, -5}, {0, -4}, {0, -3}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {1, -4}, {1, -3}, {1, -2}, {1, -1}, {1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {2, -4}, {2, -3}, {2, -2}, {2, -1}, {2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}, {3, -4}, {3, -3}, {3, -2}, {3, -1}, {3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}, {4, -3}, {4, -2}, {4, -1}, {4, 0}, {4, 1}, {4, 2}, {4, 3}, {5, 0}},
		6:  {{-6, 0}, {-5, -3}, {-5, -2}, {-5, -1}, {-5, 0}, {-5, 1}, {-5, 2}, {-5, 3}, {-4, -4}, {-4, -3}, {-4, -2}, {-4, -1}, {-4, 0}, {-4, 1}, {-4, 2}, {-4, 3}, {-4, 4}, {-3, -5}, {-3, -4}, {-3, -3}, {-3, -2}, {-3, -1}, {-3, 0}, {-3, 1}, {-3, 2}, {-3, 3}, {-3, 4}, {-3, 5}, {-2, -5}, {-2, -4}, {-2, -3}, {-2, -2}, {-2, -1}, {-2, 0}, {-2, 1}, {-2, 2}, {-2, 3}, {-2, 4}, {-2, 5}, {-1, -5}, {-1, -4}, {-1, -3}, {-1, -2}, {-1, -1}, {-1, 0}, {-1, 1}, {-1, 2}, {-1, 3}, {-1, 4}, {-1, 5}, {0, -6}, {0, -5}, {0, -4}, {0, -3}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {1, -5}, {1, -4}, {1, -3}, {1, -2}, {1, -1}, {1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {2, -5}, {2, -4}, {2, -3}, {2, -2}, {2, -1}, {2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}, {2, 5}, {3, -5}, {3, -4}, {3, -3}, {3, -2}, {3, -1}, {3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}, {3, 5}, {4, -4}, {4, -3}, {4, -2}, {4, -1}, {4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}, {5, -3}, {5, -2}, {5, -1}, {5, 0}, {5, 1}, {5, 2}, {5, 3}, {6, 0}},
		7:  {{-7, 0}, {-6, -3}, {-6, -2}, {-6, -1}, {-6, 0}, {-6, 1}, {-6, 2}, {-6, 3}, {-5, -4}, {-5, -3}, {-5, -2}, {-5, -1}, {-5, 0}, {-5, 1}, {-5, 2}, {-5, 3}, {-5, 4}, {-4, -5}, {-4, -4}, {-4, -3}, {-4, -2}, {-4, -1}, {-4, 0}, {-4, 1}, {-4, 2}, {-4, 3}, {-4, 4}, {-4, 5}, {-3, -6}, {-3, -5}, {-3, -4}, {-3, -3}, {-3, -2}, {-3, -1}, {-3, 0}, {-3, 1}, {-3, 2}, {-3, 3}, {-3, 4}, {-3, 5}, {-3, 6}, {-2, -6}, {-2, -5}, {-2, -4}, {-2, -3}, {-2, -2}, {-2, -1}, {-2, 0}, {-2, 1}, {-2, 2}, {-2, 3}, {-2, 4}, {-2, 5}, {-2, 6}, {-1, -6}, {-1, -5}, {-1, -4}, {-1, -3}, {-1, -2}, {-1, -1}, {-1, 0}, {-1, 1}, {-1, 2}, {-1, 3}, {-1, 4}, {-1, 5}, {-1, 6}, {0, -7}, {0, -6}, {0, -5}, {0, -4}, {0, -3}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {1, -6}, {1, -5}, {1, -4}, {1, -3}, {1, -2}, {1, -1}, {1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {2, -6}, {2, -5}, {2, -4}, {2, -3}, {2, -2}, {2, -1}, {2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}, {2, 5}, {2, 6}, {3, -6}, {3, -5}, {3, -4}, {3, -3}, {3, -2}, {3, -1}, {3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}, {3, 5}, {3, 6}, {4, -5}, {4, -4}, {4, -3}, {4, -2}, {4, -1}, {4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}, {4, 5}, {5, -4}, {5, -3}, {5, -2}, {5, -1}, {5, 0}, {5, 1}, {5, 2}, {5, 3}, {5, 4}, {6, -3}, {6, -2}, {6, -1}, {6, 0}, {6, 1}, {6, 2}, {6, 3}, {7, 0}},
		8:  {{-8, 0}, {-7, -3}, {-7, -2}, {-7, -1}, {-7, 0}, {-7, 1}, {-7, 2}, {-7, 3}, {-6, -5}, {-6, -4}, {-6, -3}, {-6, -2}, {-6, -1}, {-6, 0}, {-6, 1}, {-6, 2}, {-6, 3}, {-6, 4}, {-6, 5}, {-5, -6}, {-5, -5}, {-5, -4}, {-5, -3}, {-5, -2}, {-5, -1}, {-5, 0}, {-5, 1}, {-5, 2}, {-5, 3}, {-5, 4}, {-5, 5}, {-5, 6}, {-4, -6}, {-4, -5}, {-4, -4}, {-4, -3}, {-4, -2}, {-4, -1}, {-4, 0}, {-4, 1}, {-4, 2}, {-4, 3}, {-4, 4}, {-4, 5}, {-4, 6}, {-3, -7}, {-3, -6}, {-3, -5}, {-3, -4}, {-3, -3}, {-3, -2}, {-3, -1}, {-3, 0}, {-3, 1}, {-3, 2}, {-3, 3}, {-3, 4}, {-3, 5}, {-3, 6}, {-3, 7}, {-2, -7}, {-2, -6}, {-2, -5}, {-2, -4}, {-2, -3}, {-2, -2}, {-2, -1}, {-2, 0}, {-2, 1}, {-2, 2}, {-2, 3}, {-2, 4}, {-2, 5}, {-2, 6}, {-2, 7}, {-1, -7}, {-1, -6}, {-1, -5}, {-1, -4}, {-1, -3}, {-1, -2}, {-1, -1}, {-1, 0}, {-1, 1}, {-1, 2}, {-1, 3}, {-1, 4}, {-1, 5}, {-1, 6}, {-1, 7}, {0, -8}, {0, -7}, {0, -6}, {0, -5}, {0, -4}, {0, -3}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}, {1, -7}, {1, -6}, {1, -5}, {1, -4}, {1, -3}, {1, -2}, {1, -1}, {1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {2, -7}, {2, -6}, {2, -5}, {2, -4}, {2, -3}, {2, -2}, {2, -1}, {2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}, {2, 5}, {2, 6}, {2, 7}, {3, -7}, {3, -6}, {3, -5}, {3, -4}, {3, -3}, {3, -2}, {3, -1}, {3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}, {3, 5}, {3, 6}, {3, 7}, {4, -6}, {4, -5}, {4, -4}, {4, -3}, {4, -2}, {4, -1}, {4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}, {4, 5}, {4, 6}, {5, -6}, {5, -5}, {5, -4}, {5, -3}, {5, -2}, {5, -1}, {5, 0}, {5, 1}, {5, 2}, {5, 3}, {5, 4}, {5, 5}, {5, 6}, {6, -5}, {6, -4}, {6, -3}, {6, -2}, {6, -1}, {6, 0}, {6, 1}, {6, 2}, {6, 3}, {6, 4}, {6, 5}, {7, -3}, {7, -2}, {7, -1}, {7, 0}, {7, 1}, {7, 2}, {7, 3}, {8, 0}},
		9:  {{-9, 0}, {-8, -4}, {-8, -3}, {-8, -2}, {-8, -1}, {-8, 0}, {-8, 1}, {-8, 2}, {-8, 3}, {-8, 4}, {-7, -5}, {-7, -4}, {-7, -3}, {-7, -2}, {-7, -1}, {-7, 0}, {-7, 1}, {-7, 2}, {-7, 3}, {-7, 4}, {-7, 5}, {-6, -6}, {-6, -5}, {-6, -4}, {-6, -3}, {-6, -2}, {-6, -1}, {-6, 0}, {-6, 1}, {-6, 2}, {-6, 3}, {-6, 4}, {-6, 5}, {-6, 6}, {-5, -7}, {-5, -6}, {-5, -5}, {-5, -4}, {-5, -3}, {-5, -2}, {-5, -1}, {-5, 0}, {-5, 1}, {-5, 2}, {-5, 3}, {-5, 4}, {-5, 5}, {-5, 6}, {-5, 7}, {-4, -8}, {-4, -7}, {-4, -6}, {-4, -5}, {-4, -4}, {-4, -3}, {-4, -2}, {-4, -1}, {-4, 0}, {-4, 1}, {-4, 2}, {-4, 3}, {-4, 4}, {-4, 5}, {-4, 6}, {-4, 7}, {-4, 8}, {-3, -8}, {-3, -7}, {-3, -6}, {-3, -5}, {-3, -4}, {-3, -3}, {-3, -2}, {-3, -1}, {-3, 0}, {-3, 1}, {-3, 2}, {-3, 3}, {-3, 4}, {-3, 5}, {-3, 6}, {-3, 7}, {-3, 8}, {-2, -8}, {-2, -7}, {-2, -6}, {-2, -5}, {-2, -4}, {-2, -3}, {-2, -2}, {-2, -1}, {-2, 0}, {-2, 1}, {-2, 2}, {-2, 3}, {-2, 4}, {-2, 5}, {-2, 6}, {-2, 7}, {-2, 8}, {-1, -8}, {-1, -7}, {-1, -6}, {-1, -5}, {-1, -4}, {-1, -3}, {-1, -2}, {-1, -1}, {-1, 0}, {-1, 1}, {-1, 2}, {-1, 3}, {-1, 4}, {-1, 5}, {-1, 6}, {-1, 7}, {-1, 8}, {0, -9}, {0, -8}, {0, -7}, {0, -6}, {0, -5}, {0, -4}, {0, -3}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}, {0, 9}, {1, -8}, {1, -7}, {1, -6}, {1, -5}, {1, -4}, {1, -3}, {1, -2}, {1, -1}, {1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}, {2, -8}, {2, -7}, {2, -6}, {2, -5}, {2, -4}, {2, -3}, {2, -2}, {2, -1}, {2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}, {2, 5}, {2, 6}, {2, 7}, {2, 8}, {3, -8}, {3, -7}, {3, -6}, {3, -5}, {3, -4}, {3, -3}, {3, -2}, {3, -1}, {3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}, {3, 5}, {3, 6}, {3, 7}, {3, 8}, {4, -8}, {4, -7}, {4, -6}, {4, -5}, {4, -4}, {4, -3}, {4, -2}, {4, -1}, {4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}, {4, 5}, {4, 6}, {4, 7}, {4, 8}, {5, -7}, {5, -6}, {5, -5}, {5, -4}, {5, -3}, {5, -2}, {5, -1}, {5, 0}, {5, 1}, {5, 2}, {5, 3}, {5, 4}, {5, 5}, {5, 6}, {5, 7}, {6, -6}, {6, -5}, {6, -4}, {6, -3}, {6, -2}, {6, -1}, {6, 0}, {6, 1}, {6, 2}, {6, 3}, {6, 4}, {6, 5}, {6, 6}, {7, -5}, {7, -4}, {7, -3}, {7, -2}, {7, -1}, {7, 0}, {7, 1}, {7, 2}, {7, 3}, {7, 4}, {7, 5}, {8, -4}, {8, -3}, {8, -2}, {8, -1}, {8, 0}, {8, 1}, {8, 2}, {8, 3}, {8, 4}, {9, 0}},
		10: {{-10, 0}, {-9, -4}, {-9, -3}, {-9, -2}, {-9, -1}, {-9, 0}, {-9, 1}, {-9, 2}, {-9, 3}, {-9, 4}, {-8, -6}, {-8, -5}, {-8, -4}, {-8, -3}, {-8, -2}, {-8, -1}, {-8, 0}, {-8, 1}, {-8, 2}, {-8, 3}, {-8, 4}, {-8, 5}, {-8, 6}, {-7, -7}, {-7, -6}, {-7, -5}, {-7, -4}, {-7, -3}, {-7, -2}, {-7, -1}, {-7, 0}, {-7, 1}, {-7, 2}, {-7, 3}, {-7, 4}, {-7, 5}, {-7, 6}, {-7, 7}, {-6, -8}, {-6, -7}, {-6, -6}, {-6, -5}, {-6, -4}, {-6, -3}, {-6, -2}, {-6, -1}, {-6, 0}, {-6, 1}, {-6, 2}, {-6, 3}, {-6, 4}, {-6, 5}, {-6, 6}, {-6, 7}, {-6, 8}, {-5, -8}, {-5, -7}, {-5, -6}, {-5, -5}, {-5, -4}, {-5, -3}, {-5, -2}, {-5, -1}, {-5, 0}, {-5, 1}, {-5, 2}, {-5, 3}, {-5, 4}, {-5, 5}, {-5, 6}, {-5, 7}, {-5, 8}, {-4, -9}, {-4, -8}, {-4, -7}, {-4, -6}, {-4, -5}, {-4, -4}, {-4, -3}, {-4, -2}, {-4, -1}, {-4, 0}, {-4, 1}, {-4, 2}, {-4, 3}, {-4, 4}, {-4, 5}, {-4, 6}, {-4, 7}, {-4, 8}, {-4, 9}, {-3, -9}, {-3, -8}, {-3, -7}, {-3, -6}, {-3, -5}, {-3, -4}, {-3, -3}, {-3, -2}, {-3, -1}, {-3, 0}, {-3, 1}, {-3, 2}, {-3, 3}, {-3, 4}, {-3, 5}, {-3, 6}, {-3, 7}, {-3, 8}, {-3, 9}, {-2, -9}, {-2, -8}, {-2, -7}, {-2, -6}, {-2, -5}, {-2, -4}, {-2, -3}, {-2, -2}, {-2, -1}, {-2, 0}, {-2, 1}, {-2, 2}, {-2, 3}, {-2, 4}, {-2, 5}, {-2, 6}, {-2, 7}, {-2, 8}, {-2, 9}, {-1, -9}, {-1, -8}, {-1, -7}, {-1, -6}, {-1, -5}, {-1, -4}, {-1, -3}, {-1, -2}, {-1, -1}, {-1, 0}, {-1, 1}, {-1, 2}, {-1, 3}, {-1, 4}, {-1, 5}, {-1, 6}, {-1, 7}, {-1, 8}, {-1, 9}, {0, -10}, {0, -9}, {0, -8}, {0, -7}, {0, -6}, {0, -5}, {0, -4}, {0, -3}, {0, -2}, {0, -1}, {0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}, {0, 9}, {0, 10}, {1, -9}, {1, -8}, {1, -7}, {1, -6}, {1, -5}, {1, -4}, {1, -3}, {1, -2}, {1, -1}, {1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}, {1, 9}, {2, -9}, {2, -8}, {2, -7}, {2, -6}, {2, -5}, {2, -4}, {2, -3}, {2, -2}, {2, -1}, {2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}, {2, 5}, {2, 6}, {2, 7}, {2, 8}, {2, 9}, {3, -9}, {3, -8}, {3, -7}, {3, -6}, {3, -5}, {3, -4}, {3, -3}, {3, -2}, {3, -1}, {3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}, {3, 5}, {3, 6}, {3, 7}, {3, 8}, {3, 9}, {4, -9}, {4, -8}, {4, -7}, {4, -6}, {4, -5}, {4, -4}, {4, -3}, {4, -2}, {4, -1}, {4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}, {4, 5}, {4, 6}, {4, 7}, {4, 8}, {4, 9}, {5, -8}, {5, -7}, {5, -6}, {5, -5}, {5, -4}, {5, -3}, {5, -2}, {5, -1}, {5, 0}, {5, 1}, {5, 2}, {5, 3}, {5, 4}, {5, 5}, {5, 6}, {5, 7}, {5, 8}, {6, -8}, {6, -7}, {6, -6}, {6, -5}, {6, -4}, {6, -3}, {6, -2}, {6, -1}, {6, 0}, {6, 1}, {6, 2}, {6, 3}, {6, 4}, {6, 5}, {6, 6}, {6, 7}, {6, 8}, {7, -7}, {7, -6}, {7, -5}, {7, -4}, {7, -3}, {7, -2}, {7, -1}, {7, 0}, {7, 1}, {7, 2}, {7, 3}, {7, 4}, {7, 5}, {7, 6}, {7, 7}, {8, -6}, {8, -5}, {8, -4}, {8, -3}, {8, -2}, {8, -1}, {8, 0}, {8, 1}, {8, 2}, {8, 3}, {8, 4}, {8, 5}, {8, 6}, {9, -4}, {9, -3}, {9, -2}, {9, -1}, {9, 0}, {9, 1}, {9, 2}, {9, 3}, {9, 4}, {10, 0}},
	}
	return d[unitRadius]
}

func (gd *Definition) Buffer(cid0 int, cardinal, isActive bool) []int {
	var i []int
	iabs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	r, c := gd.RowCol(cid0)
	for m := -1; m <= 1; m++ {
		for n := -1; n <= 1; n++ {
			if cardinal && iabs(m) == iabs(n) {
				continue
			}
			cid1 := gd.CellID(r+m, c+n)
			if cid1 < 0 || (isActive && !gd.IsActive(cid1)) {
				i = append(i, -1)
				continue
			}
			i = append(i, cid1)
		}
	}
	return i
}
