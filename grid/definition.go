package grid

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
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
