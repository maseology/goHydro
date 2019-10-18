package grid

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

// Definition struct
type Definition struct {
	Coord                 map[int]mmaths.Point
	act                   map[int]bool
	eorig, norig, rot, cs float64
	nr, nc, na            int
	Name                  string
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

	nc := gd.nr * gd.nc
	cn, cx := 0, nc

	gd.act = make(map[int]bool, cx)
	if nc%8 != 0 {
		nc += 8 - nc%8 // add padding
	}
	nc /= 8

	b1 := make([]byte, nc)
	if err := binary.Read(reader, binary.LittleEndian, b1); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("Fatal error: read actives failed: %v", err)
		}
		// gd.na = cx
		// for i := 0; i < cx; i++ {
		// 	gd.act[i] = true
		// }
		if print {
			fmt.Printf(" %d cells (no actives)\n", gd.na)
		}
	} else { // active cells
		t := make([]byte, 1)
		if v, _ := reader.Read(t); v != 0 {
			return nil, fmt.Errorf("Fatal error: EOF not reached when expected")
		}
		for _, b := range b1 {
			for i := uint(0); i < 8; i++ {
				if b&(1<<i)>>i == 1 {
					gd.act[cn] = true
					gd.na++
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
		if gd.na > 0 && print {
			fmt.Printf(" %s actives\n", mmio.Thousands(int64(gd.na))) //11,118,568
		}
	}
	fmt.Println()

	gd.Coord = make(map[int]mmaths.Point, gd.na)
	cid := 0
	for i := 0; i < gd.nr; i++ {
		for j := 0; j < gd.nc; j++ {
			if v, ok := gd.act[cid]; ok {
				if v {
					p := mmaths.Point{X: gd.eorig + gd.cs*(float64(j)+0.5), Y: gd.norig - gd.cs*(float64(i)+0.5)}
					gd.Coord[cid] = p
				}
			}
			cid++
		}
	}
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

	gd := Definition{eorig: oe, norig: on, rot: rot, cs: cs, nr: int(nr), nc: int(nc)}
	if print {
		fmt.Printf(" %f\n", oe)
		fmt.Printf(" %f\n", on)
		fmt.Printf(" %f\n", rot)
		fmt.Println("", nr)
		fmt.Println("", nc)
		fmt.Printf(" %f\n", cs)
		fmt.Println("", uni)
	}

	return gd, nil
}

// Actives returns a slice of active cell IDs
func (gd *Definition) Actives() []int {
	out, i := make([]int, gd.na), 0
	for k, v := range gd.act {
		if v {
			out[i] = k
			i++
		}
	}
	return out
}

// Sactives returns a SORTED slice of active cell IDs
func (gd *Definition) Sactives() []int {
	a := gd.Actives()
	sort.Ints(a)
	return a
}

// RowCol returns row and column index for a given cell ID
func (gd *Definition) RowCol(cid int) (row, col int) {
	if cid < 0 || cid > gd.nr*gd.nc {
		log.Fatalf("Definition.RowCol error: invalid cell ID: %d", cid)
	}
	row = int(float64(cid) / float64(gd.nc))
	col = cid - row*gd.nc
	return
}

// CellID returns cell ID for a given row and column index
func (gd *Definition) CellID(row, col int) int {
	if row < 0 || row >= gd.nr || col < 0 || col >= gd.nc {
		log.Fatalf("Definition.CellID error: invalid [row,col]: [%d,%d]", row, col)
	}
	return row*gd.nc + col
}

// Nactives returns the count of active grid cells
func (gd *Definition) Nactives() int {
	return gd.na
}

// Ncells returns the count of grid cells
func (gd *Definition) Ncells() int {
	return gd.nc * gd.nr
}

// CellWidth returns the (uniform) width of the grid cells
func (gd *Definition) CellWidth() float64 {
	return gd.cs
}

// CellArea returns the (uniform) area of the grid cells
func (gd *Definition) CellArea() float64 {
	return gd.cs * gd.cs
}

// SaveAs writes a grid definition file of format *.gdef
func (gd *Definition) SaveAs(fp string) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf(" Definition.SaveAs: %v", err)
	}
	defer t.Close()
	t.WriteLine(fmt.Sprintf("%f", gd.eorig))
	t.WriteLine(fmt.Sprintf("%f", gd.norig))
	t.WriteLine(fmt.Sprintf("%f", gd.rot))
	t.WriteLine(fmt.Sprintf("%d", gd.nr))
	t.WriteLine(fmt.Sprintf("%d", gd.nc))
	t.WriteLine(fmt.Sprintf("U%f", gd.cs))
	return nil
}

// ToASCheader writes ASC grid header info to writer
func (gd *Definition) ToASCheader(t *mmio.TXTwriter) {
	t.WriteLine(fmt.Sprintf("ncols %d", gd.nc))
	t.WriteLine(fmt.Sprintf("nrows %d", gd.nr))
	t.WriteLine(fmt.Sprintf("xllcorner %f", gd.eorig))
	t.WriteLine(fmt.Sprintf("yllcorner %f", gd.norig-float64(gd.nr)*gd.cs))
	t.WriteLine(fmt.Sprintf("cellsize %f", gd.cs))
	t.WriteLine(fmt.Sprintf("nodata_value %d", -9999))
}

// ToASC creates an ascii-grid based on grid definition.
// If the grid definition contains active cells,
// they will be given a value of 1 in the raster.
func (gd *Definition) ToASC(fp string) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf(" Definition.ToASC: %v", err)
	}
	defer t.Close()
	gd.ToASCheader(t)
	if gd.na > 0 {
		c := 0
		for i := 0; i < gd.nr; i++ {
			for j := 0; j < gd.nc; j++ {
				if gd.act[c] {
					t.Write("1 ")
				} else {
					t.Write("-9999 ")
				}
				c++
			}
			t.Write("\n")
		}
	} else {
		for i := 0; i < gd.nr; i++ {
			for j := 0; j < gd.nc; j++ {
				t.Write("-9999 ")
			}
			t.Write("\n")
		}
	}
	return nil
}

// ToAscData converts a map referenced to cell id to an ASCII grid
func (gd *Definition) ToAscData(fp string, d map[int]float64) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf("GDEF ToASC: %v", err)
	}
	defer t.Close()
	gd.ToASCheader(t)
	cid := 0
	for i := 0; i < gd.nr; i++ {
		for j := 0; j < gd.nc; j++ {
			if v, ok := d[cid]; ok {
				t.Write(fmt.Sprintf("%.6f ", v))
			} else {
				t.Write("-9999 ")
			}
			cid++
		}
		t.Write("\n")
	}
	return nil
}

// Intersect returns a mapping from current Definition to inputted Definition
// for now, only Definitions that share the same origin, and cell sizes are mulitples can be considered
func (gd *Definition) Intersect(toGD *Definition) map[int][]int {
	// checks
	if gd.eorig != toGD.eorig || gd.norig != toGD.norig {
		log.Fatalf("Definition.Intersect error: Definitions must have the same origin")
	}
	if gd.rot != toGD.rot {
		log.Fatalf("Definition.Intersect error: Definitions not in same orientation (i.e., rotation)")
	}
	intsct := make(map[int][]int, gd.Nactives())
	if gd.cs > toGD.cs {
		log.Fatalf("Definition.Intersect TODO")
		log.Fatalf("Definition.Intersect: NNED TO CHECK CODE, not yet used.....")
		if math.Mod(gd.cs, toGD.cs) != 0. {
			log.Fatalf("Definition.Intersect error: Definitions grid definitions are not multiples: fromGD: %f, toGD: %f", gd.cs, toGD.cs)
		}
		scale := int(toGD.cs / gd.cs)
		for _, c := range gd.Actives() {
			i, j := gd.RowCol(c)
			tocid := toGD.CellID(i*scale, j*scale)
			intsct[c] = []int{tocid} // THIS IS INCONSISTENT ++++++++++++++++++++++++++++++++++++++++++++++++++++++
		}
	} else if gd.cs < toGD.cs {
		if math.Mod(toGD.cs, gd.cs) != 0. {
			log.Fatalf("Definition.Intersect error: Definitions grid definitions are not multiples: fromGD: %f, toGD: %f", gd.cs, toGD.cs)
		}
		scale := toGD.cs / gd.cs
		for _, c := range gd.Actives() {
			i, j := gd.RowCol(c)
			tocid := toGD.CellID(int(float64(i)/scale), int(float64(j)/scale))
			intsct[c] = []int{tocid}
		}
	} else {
		log.Fatalf("Definition.Intersect TODO")
	}
	return intsct
}
