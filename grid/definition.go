package grid

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/maseology/mmio"
)

// Definition struct
type Definition struct {
	act                   map[int]bool
	eorig, norig, rot, cs float64
	nr, nc, na            int
}

// ReadGDEF imports a grid definition file
func ReadGDEF(fp string) (*Definition, error) {
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

	gd, err := parseHeader(a)
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

	for {
		b1 := make([]byte, nc)
		n1, err := reader.Read(b1)
		if err == io.EOF {
			if gd.na == 0 {
				fmt.Println(" no active cells")
			}
			break
		} else if err != nil {
			return nil, fmt.Errorf("Fatal error(s): ReadGDEF:\n   failed to read 'actives': %v", err)
		} else {
			for i := 0; i < n1; i++ {
				// fmt.Println(i, b1[i], mmio.BitArray1(b1[i]))
				if b1[i] == 0 {
					for p := 0; p < 8; p++ {
						gd.act[cn] = false
						cn++
						if cn >= cx {
							break
						}
					}
				} else if b1[i] == 255 {
					for p := 0; p < 8; p++ {
						gd.act[cn] = true
						gd.na++
						cn++
						if cn >= cx {
							break
						}
					}
				} else {
					ba := mmio.BitArray1(b1[i])
					for p := 0; p < 8; p++ {
						gd.act[cn] = ba[p]
						if ba[p] {
							gd.na++
						}
						cn++
						if cn >= cx {
							break
						}
					}
				}
			}
		}
	}
	if cn != cx {
		return nil, fmt.Errorf("Fatal error(s): ReadGDEF:\n   number of cells found (%d) not equal to total (%d): %v", cn, cx, err)
	}
	if gd.na > 0 {
		fmt.Printf(" %d actives\n", gd.na) //11,118,568
	}
	fmt.Println()

	return &gd, nil
}

func parseHeader(a []string) (Definition, error) {
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
	fmt.Printf(" %f\n", oe)
	fmt.Printf(" %f\n", on)
	fmt.Printf(" %f\n", rot)
	fmt.Println("", nr)
	fmt.Println("", nc)
	fmt.Printf(" %f\n", cs)
	fmt.Println("", uni)

	return gd, nil
}

// CellWidth returns the (uniform) width of the grid cells
func (gd *Definition) CellWidth() float64 {
	return gd.cs
}

// CellArea returns the (uniform) area of the grid cells
func (gd *Definition) CellArea() float64 {
	return gd.cs * gd.cs
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
		return fmt.Errorf("GDEF ToASC: %v", err)
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
