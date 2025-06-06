package grid

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"maps"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmaths/slice"
	"github.com/maseology/mmio"
)

// Definition struct of a uniform grid
type Definition struct {
	Coord                          map[int]mmaths.Point
	Act                            map[int]int // [cellID]activeID
	cwidths, cheights              []float64   // variable cell widths and heights
	Sactives                       []int       // an ordered slice of active cell IDs
	Eorig, Norig, Rotation, Cwidth float64     // Xul; Yul; grid rotation about ULorigin; cell width
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
	gd.Act = make(map[int]int, gd.Nact)
	for i := range gd.Nact {
		gd.Sactives[i] = i
		gd.Act[i] = i
	}
	gd.Coord = make(map[int]mmaths.Point, gd.Nact)
	cid := 0
	for i := range gd.Nrow {
		for j := range gd.Ncol {
			p := mmaths.Point{X: gd.Eorig + UniformCellSize*(float64(j)+0.5), Y: gd.Norig - UniformCellSize*(float64(i)+0.5)}
			gd.Coord[cid] = p
			cid++
		}
	}
	return &gd
}

func (gd *Definition) Copy() Definition {

	coord, act := make(map[int]mmaths.Point, len(gd.Coord)), make(map[int]int, len(gd.Act))
	maps.Copy(coord, gd.Coord)
	maps.Copy(act, gd.Act)
	cwidths, cheights := make([]float64, len(gd.cwidths)), make([]float64, len(gd.cheights))
	copy(cwidths, gd.cwidths)
	copy(cheights, gd.cheights)
	sactives := make([]int, len(gd.Sactives))
	copy(sactives, gd.Sactives)

	return Definition{
		Coord:    coord,
		Act:      act,
		cwidths:  cwidths,
		cheights: cheights,
		Sactives: sactives,
		Eorig:    gd.Eorig,
		Norig:    gd.Norig,
		Rotation: gd.Rotation,
		Cwidth:   gd.Cwidth,
		Nrow:     gd.Nrow,
		Ncol:     gd.Ncol,
		Nact:     gd.Nact,
		Name:     gd.Name,
	}
}

func BuildDefinitionFromPoints(nam string, cxy map[int][]float64) *Definition {
	var gd Definition
	gd.Name = nam

	angle := func(p0, p1 []float64) float64 { return math.Atan2(p1[1]-p0[1], p1[0]-p0[0]) }

	rot := angle(cxy[0], cxy[1])
	if math.Abs(rot) < 0.001 {
		gd.Rotation = 0.
		panic("grid.Definition.BuildDefinitionFromPoints (non-rotation) TODO")
	} else {
		isclose := func(f0, f1 float64) bool { return math.Abs((f1-f0)/f1) < .01 }
		dist := func(p0, p1 []float64) float64 {
			return math.Sqrt((p0[0]-p1[0])*(p0[0]-p1[0]) + (p0[1]-p1[1])*(p0[1]-p1[1]))
		}

		ew0 := dist(cxy[0], cxy[1])
		// aew0 := math.Abs(ew0)
		// if cxy[1][1] < cxy[0][1] {
		// 	ew0 *= -1
		// }
		for i := 1; i < len(cxy); i++ {
			if !isclose(dist(cxy[i-1], cxy[i]), dist(cxy[i], cxy[i+1])) {
				gd.Ncol = i + 1
				break
			}
		}
		if math.Mod(float64(len(cxy)), float64(gd.Ncol)) != 0. {
			panic("grid.Definition.BuildDefinitionFromPoints ERROR: point distribution is not square")
		}
		gd.Nrow = len(cxy) / gd.Ncol
		gd.Nact = gd.Nrow * gd.Ncol
		if gd.Nact != len(cxy) {
			panic("grid.Definition.BuildDefinitionFromPoints ERROR: cell count error")
		}
		gd.Rotation = angle(cxy[0], cxy[gd.Ncol-1]) // refine angle
		gd.Sactives = make([]int, gd.Nact)
		gd.Act = make(map[int]int, gd.Nact)
		for i := 0; i < gd.Nact; i++ {
			gd.Sactives[i] = i
			gd.Act[i] = i
		}
		nw0 := dist(cxy[0], cxy[gd.Ncol])
		isuniform := true
		// if !isclose(ew0, nw0) { // commented out as ew0 != nw0 for lat-longs
		// 	isuniform = false
		// 	panic("grid.Definition.BuildDefinitionFromPoints isclose ERROR1")
		// }
		rotate := func(rot, x, y float64) (xp, yp float64) {
			xp = x*math.Cos(rot) - y*math.Sin(rot)
			yp = x*math.Sin(rot) + y*math.Cos(rot)
			return
		}

		xp, yp := rotate(gd.Rotation, -ew0/2, nw0/2)
		gd.Eorig, gd.Norig = cxy[0][0]+xp, cxy[0][1]+yp //  assumes no grid refinements at boundaries
		for i := 0; i < gd.Nrow; i++ {
			for j := 1; j < gd.Ncol; j++ {
				jj := i*gd.Ncol + j
				d1 := dist(cxy[jj-1], cxy[jj])
				if !isclose(d1, ew0) {
					isuniform = false
					panic("grid.Definition.BuildDefinitionFromPoints isclose ERROR2 cols")
				}
			}
		}
		for j := 0; j < gd.Ncol; j++ {
			for i := 1; i < gd.Nrow; i++ {
				d1 := dist(cxy[(i-1)*gd.Ncol+j], cxy[i*gd.Ncol+j])
				if !isclose(d1, nw0) {
					isuniform = false
					panic("grid.Definition.BuildDefinitionFromPoints isclose ERROR3 rows")
				}
			}
		}
		if isuniform {
			gd.Cwidth = -1
			gd.cwidths, gd.cheights = []float64{ew0}, []float64{nw0}
			gd.Coord = make(map[int]mmaths.Point, gd.Nact)
			cid, porig := 0, mmaths.Point{X: gd.Eorig, Y: gd.Norig}
			for i := 0; i < gd.Nrow; i++ {
				for j := 0; j < gd.Ncol; j++ {
					p := mmaths.Point{X: ew0*(float64(j)+0.5) + gd.Eorig, Y: gd.Norig - nw0*(float64(i)+0.5)}
					gd.Coord[cid] = p.Rotate(gd.Rotation, porig)
					cid++
				}
			}
			lst := []string{"cid,lat,lng"}
			for c := 0; c < cid; c++ {
				lst = append(lst, fmt.Sprintf("%d,%f,%f", c, gd.Coord[c].Y, gd.Coord[c].X))
			}
			lst = append(lst, fmt.Sprintf("%d,%f,%f", -1, gd.Norig, gd.Eorig))
			mmio.WriteLines(fmt.Sprintf("M:/_gdfef-test/%s.csv", gd.Name), lst)
			// For j = 0 To _ncol
			// With OrderedXYs(_nrow * (_ncol + 1) + j).Clone
			// 	.Rotate(-_rotation, _origin)
			// 	_ec(j) = .X
			// End With
			// Next
			// For i = 0 To _nrow
			// 	With OrderedXYs(i * (_ncol + 1) + _ncol).Clone
			// 		.Rotate(-_rotation, _origin)
			// 		_nc(i) = .Y
			// 	End With
			// Next
		} else {
			gd.cwidths, gd.cheights = make([]float64, gd.Ncol), make([]float64, gd.Nrow)
			panic("grid.Definition.BuildDefinitionFromPoints non-uniform girds not supported")
		}
	}
	return &gd
}

// ReadGDEF imports a grid definition file
func ReadGDEF(gdeffp string, print bool) (*Definition, error) {
	file, err := os.Open(gdeffp)
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

	parseHeader := func(a []string, print bool) (Definition, bool, error) {
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
		_ = cs
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
			// stErr = append(stErr, " *** Fatal error: ReadGDEF.parseHeader: non-uniform grids currently not supported ***")
			cs = -1.
		}

		// error handling
		if len(stErr) > 0 {
			return Definition{}, uni, fmt.Errorf("fatal error(s): ReadGDEF.parseHeader:\n%s", strings.Join(stErr, "\n"))
		}

		gd := Definition{Eorig: oe, Norig: on, Rotation: rot, Cwidth: cs, Nrow: int(nr), Ncol: int(nc)}
		if print {
			fmt.Printf("\n   opening %s\n", gdeffp)
			fmt.Printf("    xul\t\t%.1f\n", oe)
			fmt.Printf("    yul\t\t%.1f\n", on)
			fmt.Printf("    rotation\t%f\n", rot)
			fmt.Printf("    nrows\t%d\n", nr)
			fmt.Printf("    ncols\t%d\n", nc)
			fmt.Printf("    cell size\t%.3f\n", cs)
			fmt.Printf("    is uniform:\t%t\n", uni)
		}

		return gd, uni, nil
	}
	gd, isuniform, err := parseHeader(a, print)
	if err != nil {
		return nil, err
	}
	if !isuniform {
		// slst := make([]float64, gd.Nrow+gd.Ncol)
		// slst[0], _ = strconv.ParseFloat(a[5], 64)
		// for i := 1; i < gd.Nrow+gd.Ncol; i++ {
		// 	line, _, err := reader.ReadLine()
		// 	if err == io.EOF {
		// 		return nil, fmt.Errorf("ReadTextLines (cell widths): %v", err)
		// 	} else if err != nil {
		// 		return nil, fmt.Errorf("ReadTextLines (cell widths): %v", err)
		// 	}
		// 	slst[i], _ = strconv.ParseFloat(string(line), 64)
		// }
		// gd.VCwidth = slst

		gd.cwidths, gd.cheights = make([]float64, gd.Ncol), make([]float64, gd.Nrow)
		gd.cheights[0], _ = strconv.ParseFloat(a[5], 64)
		for i := 1; i < gd.Nrow; i++ {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				return nil, fmt.Errorf("ReadTextLines (cell widths): %v", err)
			} else if err != nil {
				return nil, fmt.Errorf("ReadTextLines (cell widths): %v", err)
			}
			gd.cheights[i], _ = strconv.ParseFloat(string(line), 64)
		}
		for j := 0; j < gd.Ncol; j++ {
			line, _, err := reader.ReadLine()
			if err == io.EOF {
				return nil, fmt.Errorf("ReadTextLines (cell heights): %v", err)
			} else if err != nil {
				return nil, fmt.Errorf("ReadTextLines (cell heights): %v", err)
			}
			gd.cwidths[j], _ = strconv.ParseFloat(string(line), 64)
		}
	}

	nc := gd.Nrow * gd.Ncol
	cn, cx := 0, nc
	if nc%8 != 0 {
		nc += 8 - nc%8 // add padding
	}
	nc /= 8

	b1 := make([]byte, nc)
	if err := binary.Read(reader, binary.LittleEndian, b1); err != nil { // no active cells
		if err != io.EOF {
			return nil, fmt.Errorf("fatal error: read actives failed: %v", err)
		}
		if print {
			fmt.Printf("    (no active cells)\n")
		}
		gd.Sactives = make([]int, cx)
		gd.Act = make(map[int]int, cx)
		gd.Nact = cx
		for i := 0; i < cx; i++ {
			gd.Sactives[i] = i
			gd.Act[i] = i
		}
		gd.Coord = make(map[int]mmaths.Point, cx)
		cid := 0
		for i := 0; i < gd.Nrow; i++ {
			for j := 0; j < gd.Ncol; j++ {
				p := mmaths.Point{X: gd.Eorig + gd.Cwidth*(float64(j)+0.5), Y: gd.Norig - gd.Cwidth*(float64(i)+0.5)}
				if gd.Rotation != 0 {
					p = p.Rotate(gd.Rotation, mmaths.Point{X: gd.Eorig, Y: gd.Norig})
				}
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
		gd.Act = make(map[int]int, cx)
		for _, b := range b1 {
			for i := uint(0); i < 8; i++ {
				if b&(1<<i)>>i == 1 {
					gd.Sactives = append(gd.Sactives, cn)
					gd.Act[cn] = gd.Nact
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
			fmt.Printf("    %s\tactive cells\n", mmio.Thousands(int64(gd.Nact))) //11,118,568
		}
		gd.Coord = make(map[int]mmaths.Point, gd.Nact)
		cid := 0
		for i := 0; i < gd.Nrow; i++ {
			for j := 0; j < gd.Ncol; j++ {
				if _, ok := gd.Act[cid]; ok {
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

func (gd *Definition) IsSimilar(gd1 *Definition) bool {
	if gd.Ncol != gd1.Ncol {
		return false
	}
	if gd.Nrow != gd1.Nrow {
		return false
	}
	if gd.CellArea() != gd1.CellArea() {
		return false
	}
	return true
}

// IsActive returns whether a cell ID is of an active cell
func (gd *Definition) IsActive(cid int) bool {
	_, ok := gd.Act[cid]
	return ok
}

func (gd *Definition) ResetActives(cids []int) {
	gd.Nact = len(cids)
	gd.Act = make(map[int]int, gd.Nact)
	gd.Sactives = make([]int, gd.Nact)
	copy(gd.Sactives, cids)
	sort.Ints(gd.Sactives)
	for i, c := range gd.Sactives {
		gd.Act[c] = i
	}
}

func (gd *Definition) RemoveActives(rmcids []int) {
	drm := make(map[int]int, len(rmcids))
	for _, rm := range rmcids {
		drm[rm]++
	}
	newSactives := make([]int, 0, len(gd.Sactives)-len(drm))
	for _, c := range gd.Sactives {
		if _, ok := drm[c]; ok {
			continue
		}
		newSactives = append(newSactives, c)
	}
	gd.ResetActives(newSactives)
}

// // Actives returns a slice of active cell IDs
// func (gd *Definition) Actives() []int {
// 	out, i := make([]int, gd.Nact), 0
// 	for k, v := range gd.Act {
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

// Extents Left Up Right Down
func (gd *Definition) Extents() []float64 {
	return []float64{gd.Eorig, gd.Norig, gd.Eorig + gd.Cwidth*float64(gd.Ncol), gd.Norig - gd.Cwidth*float64(gd.Nrow)} // Left Up Right Down
}

func (gd *Definition) CellOriginUL(cid int) (r, c int, x0, y0 float64) {
	r, c = gd.RowCol(cid)
	if len(gd.cheights) == 0 { // uniform cells
		return r, c, gd.Eorig + float64(c)*gd.Cwidth, gd.Norig - float64(r)*gd.Cwidth
	}

	sr, sc := gd.Norig, gd.Eorig
	for i := 0; i < r; i++ {
		sr -= gd.cheights[i]
	}
	for j := 0; j < c; j++ {
		sc += gd.cwidths[j]
	}
	return r, c, sc, sr
}

func (gd *Definition) CellCentroid(cid int) []float64 {
	if len(gd.Coord) > 0 {
		if p, ok := gd.Coord[cid]; ok {
			return []float64{p.X, p.Y}
		}
	}

	if len(gd.cheights) == 0 { // uniform cells
		r, c := gd.RowCol(cid)
		return []float64{gd.Eorig + (float64(c)+.5)*gd.Cwidth, gd.Norig - (float64(r)+.5)*gd.Cwidth}
	}

	r, c, sc, sr := gd.CellOriginUL(cid)
	return []float64{sc + gd.cwidths[c]/2., sr - gd.cheights[r]/2.}
}

func (gd *Definition) CellCentroids() map[int][]float64 {
	m := make(map[int][]float64, gd.Nact)
	if len(gd.Coord) > 0 {
		for c, p := range gd.Coord {
			m[c] = []float64{p.X, p.Y}
		}
	} else {
		for _, c := range gd.Sactives {
			m[c] = gd.CellCentroid(c)
		}
	}
	return m
}

func (gd *Definition) CellPerimeter(cid int) [][]float64 {
	// p1---p2   y       0---nc
	//  | c |    |       |       clockwise, left-top-right-bottom
	// p0---p3   0---x   nr
	if len(gd.cheights) == 0 { // uniform cells
		cw2 := gd.Cwidth / 2
		ctrd := gd.CellCentroid(cid)
		return [][]float64{
			{ctrd[0] - cw2, ctrd[1] - cw2},
			{ctrd[0] - cw2, ctrd[1] + cw2},
			{ctrd[0] + cw2, ctrd[1] + cw2},
			{ctrd[0] + cw2, ctrd[1] - cw2},
			{ctrd[0] - cw2, ctrd[1] - cw2}, // same as first point
		}
	}

	r, c, sc, sr := gd.CellOriginUL(cid)
	p0 := []float64{sc, sr - gd.cheights[r]}
	p1 := []float64{sc, sr}
	p2 := []float64{sc + gd.cwidths[c], sr}
	p3 := []float64{sc + gd.cwidths[c], sr - gd.cheights[r]}

	return [][]float64{p0, p1, p2, p3, p0}
}

// CellIndexXR returns a mapping of cell id to an array index
func (gd *Definition) CellIndexXR() map[int]int {
	m := make(map[int]int, len(gd.Sactives))
	for i, c := range gd.Sactives {
		m[c] = i
	}
	return m
}

func (gd *Definition) NullArray(nodatavalue float64) []float64 {
	nc := gd.Ncells()
	o := make([]float64, nc)
	for i := 0; i < nc; i++ {
		o[i] = nodatavalue
	}
	return o
}

func (gd *Definition) NullFloat32(nodatavalue float32) []float32 {
	nc := gd.Ncells()
	o := make([]float32, nc)
	for i := 0; i < nc; i++ {
		o[i] = nodatavalue
	}
	return o
}

func (gd *Definition) NullInt32(nodatavalue int32) []int32 {
	nc := gd.Ncells()
	o := make([]int32, nc)
	for i := 0; i < nc; i++ {
		o[i] = nodatavalue
	}
	return o
}

func (gd *Definition) ExtentToCellIDs(ext mmaths.Extent) []int {
	var wg sync.WaitGroup
	wg.Add(2)
	var iul, jul, ilr, jlr int
	go func() {
		iul, jul = gd.PointToRowCol(ext.Xn, ext.Yx)
		wg.Done()
	}()
	go func() {
		ilr, jlr = gd.PointToRowCol(ext.Xx, ext.Yn)
		wg.Done()
	}()
	wg.Wait()
	cids := []int{}
	for i := iul; i <= ilr; i++ {
		for j := jul; j <= jlr; j++ {
			cids = append(cids, gd.CellID(i, j))
		}
	}
	return cids
}

// PointToCellID returns the cell id that contains the xy coordinates
func (gd *Definition) PointToCellID(x, y float64) int {
	return gd.CellID(gd.PointToRowCol(x, y))
}

func (gd *Definition) PointToActiveCellID(x, y float64) int {
	cid := gd.CellID(gd.PointToRowCol(x, y))
	if gd.IsActive(cid) {
		return cid
	}
	return -1
}

// PointToRowCol returns the row and column grid cell that contains the xy coordinates
func (gd *Definition) PointToRowCol(x, y float64) (row, col int) {
	if gd.Rotation != 0 {
		panic("rotated grids not supported")
	}
	if gd.Cwidth < 0 {
		panic("non-uniform grid not supported")
	}

	var wg sync.WaitGroup
	wg.Add(2)

	if gd.Rotation != 0. {
		log.Fatalf(" Definition.PointToRowCol todo")
	}

	row = -1
	col = -1
	go func() {
		defer wg.Done()
		if y > gd.Norig {
			return
		}
		for {
			row++
			if row > gd.Nrow {
				row = -1 // gd.Nrow +1
				return
			}
			if gd.Norig-float64(row+1)*gd.Cwidth <= y {
				break
			}
		}
	}()
	go func() {
		defer wg.Done()
		if x < gd.Eorig {
			return
		}
		for {
			col++
			if col > gd.Ncol {
				col = -1 // gd.Ncol +1
				return
			}
			if gd.Eorig+float64(col+1)*gd.Cwidth >= x {
				return
			}
		}
	}()

	wg.Wait()
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

func (gd *Definition) LineToCellIDs(x0, y0, x1, y1 float64) []int {
	if gd.Rotation != 0 {
		panic("GD.LineToCellIDs TODO")
	}

	// see (for example) Lindsay 2016 The practice of DEM stream burning revisited
	ln0 := mmaths.LineSegment{
		P0: mmaths.Point{X: x0, Y: y0},
		P1: mmaths.Point{X: x1, Y: y1},
	}
	ln0.Build()

	r0, c0 := gd.PointToRowCol(x0, y0)
	r1, c1 := gd.PointToRowCol(x1, y1)
	if r0 < 0 || c0 < 0 || r1 < 0 || c1 < 0 { // line segment intersecting grid, but has point(s) outside -- adjust row cols,
		ext := gd.Extents() // Left Up Right Down
		lnL := mmaths.LineSegment{
			P0: mmaths.Point{X: ext[0], Y: ext[3]},
			P1: mmaths.Point{X: ext[0], Y: ext[1]},
		}
		lnR := mmaths.LineSegment{
			P0: mmaths.Point{X: ext[2], Y: ext[3]},
			P1: mmaths.Point{X: ext[2], Y: ext[1]},
		}
		lnU := mmaths.LineSegment{
			P0: mmaths.Point{X: ext[0], Y: ext[1]},
			P1: mmaths.Point{X: ext[2], Y: ext[1]},
		}
		lnD := mmaths.LineSegment{
			P0: mmaths.Point{X: ext[0], Y: ext[3]},
			P1: mmaths.Point{X: ext[2], Y: ext[3]},
		}

		var intersectionPoints []*mmaths.Point
		var dists []float64
		appendPoint := func(ln *mmaths.LineSegment) {
			if p, dist := ln0.Intersection2D(ln); !math.IsNaN(dist) {
				intersectionPoints = append(intersectionPoints, p)
				dists = append(dists, dist)
			}
		}
		appendPoint(&lnL)
		appendPoint(&lnU)
		appendPoint(&lnR)
		appendPoint(&lnD)

		switch len(intersectionPoints) {
		case 0:
			return nil
		case 1:
			if r0 < 0 || c0 < 0 {
				pnew := intersectionPoints[0]
				r0, c0 = gd.PointToRowCol(pnew.X, pnew.Y)
			} else if r1 < 0 || c1 < 0 {
				pnew := intersectionPoints[0]
				r1, c1 = gd.PointToRowCol(pnew.X, pnew.Y)
			} else {
				panic("should not occur 0")
			}
		case 2:
			pnew0, pnew1 := intersectionPoints[0], intersectionPoints[1]
			d00 := ln0.P0.Distance(*pnew0)
			d01 := ln0.P0.Distance(*pnew1)
			// d10 := ln0.P1.Distance(pnew0)
			// d11 := ln0.P1.Distance(pnew1)
			if d00 < d01 {
				r0, c0 = gd.PointToRowCol(pnew0.X, pnew0.Y)
				r1, c1 = gd.PointToRowCol(pnew1.X, pnew1.Y)
			} else {
				r1, c1 = gd.PointToRowCol(pnew0.X, pnew0.Y)
				r0, c0 = gd.PointToRowCol(pnew1.X, pnew1.Y)
			}
		default:
			return nil
		}
	}

	var aIntersect []*mmaths.Point
	if c0 != c1 {
		j0, j1 := 0, 0
		if c1 > c0 {
			j0 = c0 + 1
			j1 = c1
		} else {
			j0 = c1 + 1
			j1 = c0
		}
		for j := j0; j <= j1; j++ {
			lni := ln0.IntersectionX(gd.Eorig + gd.Cwidth*(float64(j)+.5))
			if lni != nil {
				aIntersect = append(aIntersect, lni)
			}
		}
	}
	if r0 != r1 {
		i0, i1 := 0, 0
		if r1 > r0 {
			i0 = r0 + 1
			i1 = r1
		} else {
			i0 = r1 + 1
			i1 = r0
		}
		for i := i0; i <= i1; i++ {
			lni := ln0.IntersectionY(gd.Norig - gd.Cwidth*(float64(i)+.5))
			if lni != nil {
				aIntersect = append(aIntersect, lni)
			}
		}
	}

	// sort intersections along polyline
	dX, dY := make(map[int]float64, len(aIntersect)), make(map[int]float64, len(aIntersect))
	for i, xy1 := range aIntersect {
		dX[i] = xy1.X
		dY[i] = xy1.Y
	}
	dy, dx := math.Abs(y0-y1), math.Abs(x0-x1)
	if y0 == y1 || dx > dy {
		aIntersect = make([]*mmaths.Point, len(dX))
		if len(dX) == 1 {
			for k, v := range dX {
				aIntersect[0] = &mmaths.Point{X: v, Y: dY[k]}
			}
		} else {
			ii, _ := mmaths.SortMapFloat(dX, x0 > x1)
			for k, iii := range ii {
				aIntersect[k] = &mmaths.Point{X: dX[iii], Y: dY[iii]}
			}
		}
	} else { //if x0 == x1 {
		aIntersect = make([]*mmaths.Point, len(dY))
		if len(dY) == 1 {
			for k, v := range dY {
				aIntersect[0] = &mmaths.Point{X: dX[k], Y: v}
			}
		} else {
			ii, _ := mmaths.SortMapFloat(dY, y0 > y1)
			for k, iii := range ii {
				aIntersect[k] = &mmaths.Point{X: dX[iii], Y: dY[iii]}
			}
		}
	}

	// d := make(map[int]bool, len(aIntersect))
	// for _, p := range aIntersect {
	// 	if gd.Rotation != 0.0 {
	// 		panic("GD.LineToCellIDs TODO")
	// 		// p.Rotate(_rotation, False)
	// 	}
	// 	d[gd.PointToCellID(p.X, p.Y)] = true
	// }
	// o := make([]int,  len(d))
	// for k := range d {
	// 	o = append(o, k)
	// }
	// sort.Ints(o)
	o := make([]int, len(aIntersect)+2)
	o[0] = gd.PointToCellID(x0, y0)
	o[len(aIntersect)+1] = gd.PointToCellID(x1, y1)
	for i, p := range aIntersect {
		if gd.Rotation != 0.0 {
			panic("GD.LineToCellIDs TODO")
		}
		o[i+1] = gd.PointToCellID(p.X, p.Y)
	}
	o = slice.Distinct(o)
	return o
}

func (gd *Definition) PolygonToCellIDs(vertices [][]float64) ([]int, int) {
	var pr PolygonRasterizer
	return pr.InteriorCellIDs(gd, vertices) // cids, ncells
}

func (gd *Definition) Buffers(cardinal, isActive bool) map[int][]int {
	o := make(map[int][]int, gd.Nact)
	if gd.Act == nil {
		isActive = false
	}
	for _, c := range gd.Sactives {
		o[c] = gd.Buffer(c, cardinal, isActive)
	}
	return o
}

func (gd *Definition) Buffer(cid0 int, cardinal, isActive bool) []int {
	i := make([]int, 0, 8)
	// 0 1 2
	// 3   4
	// 5 6 7
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
			if m == 0 && n == 0 {
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

func (gd *Definition) CropToActives() (*Definition, map[int]int) {

	rn, rx, cn, cx := 1000000, -1, 1000000, -1
	for _, cid := range gd.Sactives {
		r, c := gd.RowCol(cid)
		if r > rx {
			rx = r
		}
		if r < rn {
			rn = r
		}
		if c > cx {
			cx = c
		}
		if c < cn {
			cn = c
		}
	}
	nnr, nnc := rx-rn+1, cx-cn+1

	ogd := NewDefinition(gd.Name+"-cropped", nnr, nnc, gd.Cwidth)
	_, _, ogd.Eorig, ogd.Norig = gd.CellOriginUL(gd.CellID(rn, cn))
	ogd.Nact = gd.Nact
	ogd.Sactives = make([]int, ogd.Nact)
	ogd.Act = make(map[int]int, nnr*nnc)
	xr := make(map[int]int, nnr*nnc)
	for i, cid := range gd.Sactives {
		r, c := gd.RowCol(cid)
		cidn := ogd.CellID(r-rn, c-cn)
		ogd.Sactives[i] = cidn
		ogd.Coord[cidn] = gd.Coord[cid]
		ogd.Act[cidn] = i
		xr[cid] = cidn
	}

	return ogd, xr
}

func (gd *Definition) Crop(xn, xx, yn, yx float64, buffer int) (*Definition, int, int) {
	rn, rx, cn, cx := -1, -1, -1, -1
	for c := 0; c < gd.Nrow; c++ {
		x := float64(c)*gd.Cwidth + gd.Eorig
		if cn < 0 && x > xn {
			cn = c - 1
		}
		if cx < 0 && x > xx {
			cx = c
			break
		}
	}
	for r := gd.Ncol - 1; r >= 0; r-- {
		y := gd.Norig - float64(r)*gd.Cwidth
		if rx < 0 && y > yn {
			rx = r + 1
		}
		if rn < 0 && y > yx {
			rn = r
			break
		}
	}
	cn -= buffer
	cx += buffer
	rn -= buffer
	rx += buffer
	nnr, nnc := rx-rn+1, cx-cn+1

	ogd := NewDefinition(gd.Name+"-cropped", nnr, nnc, gd.Cwidth)
	_, _, ogd.Eorig, ogd.Norig = gd.CellOriginUL(gd.CellID(rn, cn))
	return ogd, rn, cn
}
