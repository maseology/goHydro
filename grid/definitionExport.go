package grid

import (
	"fmt"

	"github.com/maseology/mmio"
)

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

// ToHDR creates an ESRI-grid based on grid definition header
func (gd *Definition) ToHDR(fp string, nbands int) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf(" Definition.ToASC: %v", err)
	}
	defer t.Close()
	t.WriteLine(fmt.Sprintf("ncols %d", gd.nc))
	t.WriteLine(fmt.Sprintf("nrows %d", gd.nr))
	t.WriteLine(fmt.Sprintf("nbands %d", nbands))
	t.WriteLine(fmt.Sprintf("xllcorner %f", gd.eorig))
	t.WriteLine(fmt.Sprintf("yllcorner %f", gd.norig-float64(gd.nr)*gd.cs))
	t.WriteLine(fmt.Sprintf("cellsize %f", gd.cs))
	t.WriteLine(fmt.Sprintf("nodata_value %d", -32768))
	t.WriteLine(fmt.Sprintf("nbits %d", 16))
	t.WriteLine(fmt.Sprintf("pixeltype %s", "signedint"))
	t.WriteLine(fmt.Sprintf("byteorder %s", "lsbfirst"))
	t.WriteLine(fmt.Sprintf("layout %s", "bip"))
	return nil
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
		m := make(map[int]bool, gd.na)
		for _, c := range gd.Sactives {
			m[c] = true
		}
		c := 0
		for i := 0; i < gd.nr; i++ {
			for j := 0; j < gd.nc; j++ {
				if _, ok := m[c]; ok {
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
