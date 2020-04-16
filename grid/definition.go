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
	Coord                 map[int]mmaths.Point
	act                   map[int]bool
	Sactives              []int // a sorted slice of active cell IDs
	eorig, norig, rot, Cw float64
	Nr, Nc, Na            int
	Name                  string
}

// NewDefinition constructs a basic grid definition
func NewDefinition(nam string, nr, nc int, UniformCellSize float64) *Definition {
	var gd Definition
	gd.Name = nam
	gd.Nr, gd.Nc, gd.Na = nr, nc, nr*nc
	gd.Cw = UniformCellSize
	gd.Sactives = make([]int, gd.Na)
	gd.act = make(map[int]bool, gd.Na)
	for i := 0; i < gd.Na; i++ {
		gd.Sactives[i] = i
		gd.act[i] = true
	}
	gd.Coord = make(map[int]mmaths.Point, gd.Na)
	cid := 0
	for i := 0; i < gd.Nr; i++ {
		for j := 0; j < gd.Nc; j++ {
			p := mmaths.Point{X: gd.eorig + gd.Cw*(float64(j)+0.5), Y: gd.norig - gd.Cw*(float64(i)+0.5)}
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

	gd, err := parseHeader(a, print)
	if err != nil {
		return nil, err
	}
	if gd.rot != 0. {
		return nil, fmt.Errorf("ReadGDEF error: rotation no yet supported")
	}

	nc := gd.Nr * gd.Nc
	cn, cx := 0, nc
	if nc%8 != 0 {
		nc += 8 - nc%8 // add padding
	}
	nc /= 8

	b1 := make([]byte, nc)
	if err := binary.Read(reader, binary.LittleEndian, b1); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("Fatal error: read actives failed: %v", err)
		}
		if print {
			fmt.Printf(" (no active cells)\n")
		}
		gd.Sactives = make([]int, cx)
		gd.act = make(map[int]bool, cx)
		gd.Na = cx
		for i := 0; i < cx; i++ {
			gd.Sactives[i] = i
			gd.act[i] = true
		}
		gd.Coord = make(map[int]mmaths.Point, cx)
		cid := 0
		for i := 0; i < gd.Nr; i++ {
			for j := 0; j < gd.Nc; j++ {
				p := mmaths.Point{X: gd.eorig + gd.Cw*(float64(j)+0.5), Y: gd.norig - gd.Cw*(float64(i)+0.5)}
				gd.Coord[cid] = p
				cid++
			}
		}
	} else { // active cells
		t := make([]byte, 1)
		if v, _ := reader.Read(t); v != 0 {
			return nil, fmt.Errorf("Fatal error: EOF not reached when expected")
		}
		gd.Sactives = []int{}
		gd.act = make(map[int]bool, cx)
		for _, b := range b1 {
			for i := uint(0); i < 8; i++ {
				if b&(1<<i)>>i == 1 {
					gd.Sactives = append(gd.Sactives, cn)
					gd.act[cn] = true
					gd.Na++
				}
				cn++
				if cn >= cx {
					break
				}
			}
		}
		if cn != cx {
			return nil, fmt.Errorf("Fatal error(s): ReadGDEF:\n   number of cells found (%d) not equal to total (%d): %v", cn, cx, err)
		}
		if gd.Na > 0 && print {
			fmt.Printf(" %s active cells\n", mmio.Thousands(int64(gd.Na))) //11,118,568
		}
		gd.Coord = make(map[int]mmaths.Point, gd.Na)
		cid := 0
		for i := 0; i < gd.Nr; i++ {
			for j := 0; j < gd.Nc; j++ {
				if _, ok := gd.act[cid]; ok {
					p := mmaths.Point{X: gd.eorig + gd.Cw*(float64(j)+0.5), Y: gd.norig - gd.Cw*(float64(i)+0.5)}
					gd.Coord[cid] = p
				}
				cid++
			}
		}
	}
	fmt.Println()
	return &gd, nil
}

func parseHeader(a []string, print bool) (Definition, error) {
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
		return Definition{}, fmt.Errorf("Fatal error(s): ReadGDEF.parseHeader:\n%s", strings.Join(stErr, "\n"))
	}

	gd := Definition{eorig: oe, norig: on, rot: rot, Cw: cs, Nr: int(nr), Nc: int(nc)}
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

// IsActive returns whether a cell ID is of an active cell
func (gd *Definition) IsActive(cid int) bool {
	return gd.act[cid]
}

// // Actives returns a slice of active cell IDs
// func (gd *Definition) Actives() []int {
// 	out, i := make([]int, gd.na), 0
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

// RowCol returns row and column index for a given cell ID
func (gd *Definition) RowCol(cid int) (row, col int) {
	if cid < 0 || cid > gd.Nr*gd.Nc {
		log.Fatalf("Definition.RowCol error: invalid cell ID: %d", cid)
	}
	row = int(float64(cid) / float64(gd.Nc))
	col = cid - row*gd.Nc
	return
}

// CellID returns cell ID for a given row and column index
func (gd *Definition) CellID(row, col int) int {
	if row < 0 || row >= gd.Nr || col < 0 || col >= gd.Nc {
		log.Fatalf("Definition.CellID error: invalid [row,col]: [%d,%d]", row, col)
	}
	return row*gd.Nc + col
}

// Ncells returns the count of grid cells
func (gd *Definition) Ncells() int {
	return gd.Nc * gd.Nr
}

// CellArea returns the (uniform) area of the grid cells
func (gd *Definition) CellArea() float64 {
	return gd.Cw * gd.Cw
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
	if gd.rot != 0. {
		log.Fatalf(" Definition.PointToRowCol todo")
	}
	for {
		row++
		if gd.norig-float64(row+1)*gd.Cw <= y {
			break
		}
	}
	for {
		col++
		if gd.eorig+float64(col+1)*gd.Cw >= x {
			break
		}
	}
	return
}

// ConatainsPoint returns whether a point exists within a grid definition, with a specified buffer
func (gd *Definition) ConatainsPoint(x, y, buf float64) bool {
	if x < gd.eorig-buf {
		return false
	}
	if x > gd.eorig+float64(gd.Nc)*gd.Cw+buf {
		return false
	}
	if y > gd.norig+buf {
		return false
	}
	if y < gd.norig-float64(gd.Nr)*gd.Cw-buf {
		return false
	}
	return true
}
