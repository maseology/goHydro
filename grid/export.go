package grid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/maseology/mmio"
)

// SaveAs writes a grid definition file of format *.gdef
func (gd *Definition) SaveAs(fp string) error {
	switch mmio.GetExtension(fp) {
	case ".gdef":
		t, err := mmio.NewTXTwriter(fp)
		if err != nil {
			return fmt.Errorf(" Definition.SaveAs: %v", err)
		}
		defer t.Close()
		t.WriteLine(fmt.Sprintf("%f", gd.Eorig))
		t.WriteLine(fmt.Sprintf("%f", gd.Norig))
		t.WriteLine(fmt.Sprintf("%f", gd.Rotation))
		t.WriteLine(fmt.Sprintf("%d", gd.Nrow))
		t.WriteLine(fmt.Sprintf("%d", gd.Ncol))
		t.WriteLine(fmt.Sprintf("U%f", gd.Cwidth))

		if gd.Nact > 0 {
			bActive := make([]bool, gd.Ncells())
			for _, i := range gd.Sactives {
				bActive[i] = true
			}
			// fmt.Println(bActive[gd.Sactives[0]])
			// fmt.Println(bActive[gd.Sactives[0]-1])
			// k := 0
			// for i := 0; i < gd.Nrow; i++ {
			// 	for j := 0; j < gd.Ncol; j++ {
			// 		if bActive[k] {
			// 			print("X")
			// 		} // else {
			// 		// 	print("O")
			// 		// }
			// 		k++
			// 	}
			// 	// println()
			// }
			byActives := mmio.BitArrayRev(bActive)
			// t.WriteLine(string(byActives))
			// if err := t.WriteBytes(byActives); err != nil {
			// 	panic(err)
			// }
			binary.Write(t.Writer, binary.LittleEndian, byActives)
		}
		return nil
	case ".hdr":
		t, err := mmio.NewTXTwriter(fp)
		if err != nil {
			return fmt.Errorf(" Definition.SaveAs: %v", err)
		}
		defer t.Close()
		t.WriteLine(fmt.Sprintf("ncols %d", gd.Ncol))
		t.WriteLine(fmt.Sprintf("nrows %d", gd.Nrow))
		t.WriteLine(fmt.Sprintf("xllcorner %f", gd.Eorig))
		t.WriteLine(fmt.Sprintf("yllcorner %f", gd.Norig-float64(gd.Nrow)*gd.Cwidth))
		t.WriteLine(fmt.Sprintf("cellsize %f", gd.Cwidth))
		t.WriteLine("nodata_value -9999")
		t.WriteLine("byteorder i")
		return nil
	default:
		return fmt.Errorf(" Unknown format: %s", fp)
	}
}

// ToASCheader writes ASC grid header info to writer
func (gd *Definition) ToASCheader(t *mmio.TXTwriter) {
	t.WriteLine(fmt.Sprintf("ncols %d", gd.Ncol))
	t.WriteLine(fmt.Sprintf("nrows %d", gd.Nrow))
	t.WriteLine(fmt.Sprintf("xllcorner %f", gd.Eorig))
	t.WriteLine(fmt.Sprintf("yllcorner %f", gd.Norig-float64(gd.Nrow)*gd.Cwidth))
	t.WriteLine(fmt.Sprintf("cellsize %f", gd.Cwidth))
	t.WriteLine(fmt.Sprintf("nodata_value %d", -9999))
}

// ToHDR creates an ESRI-grid based on grid definition header
func (gd *Definition) ToHDR(fp string, nbands, nbits int) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf(" Definition.ToHDR: %v", err)
	}
	defer t.Close()
	t.WriteLine(fmt.Sprintf("ncols %d", gd.Ncol))
	t.WriteLine(fmt.Sprintf("nrows %d", gd.Nrow))
	t.WriteLine(fmt.Sprintf("nbands %d", nbands))
	t.WriteLine(fmt.Sprintf("xllcorner %f", gd.Eorig))
	t.WriteLine(fmt.Sprintf("yllcorner %f", gd.Norig-float64(gd.Nrow)*gd.Cwidth))
	t.WriteLine(fmt.Sprintf("cellsize %f", gd.Cwidth))
	t.WriteLine(fmt.Sprintf("nodata_value %d", -9999))
	t.WriteLine(fmt.Sprintf("nbits %d", nbits))
	t.WriteLine(fmt.Sprintf("pixeltype %s", "signedint"))
	t.WriteLine(fmt.Sprintf("byteorder %s", "i"))
	t.WriteLine(fmt.Sprintf("layout %s", "bil"))
	// t.WriteLine(fmt.Sprintf("byteorder %s", "lsbfirst"))
	// t.WriteLine(fmt.Sprintf("layout %s", "bip"))
	return nil
}

// ToHDRfloat creates an ESRI-grid based on grid definition header for float arrays
func (gd *Definition) ToHDRfloat(fp string, nbands, nbits int) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf(" Definition.ToHToHDRfloatDR: %v", err)
	}
	defer t.Close()
	t.WriteLine(fmt.Sprintf("ncols %d", gd.Ncol))
	t.WriteLine(fmt.Sprintf("nrows %d", gd.Nrow))
	t.WriteLine(fmt.Sprintf("nbands %d", nbands))
	t.WriteLine(fmt.Sprintf("xllcorner %f", gd.Eorig))
	t.WriteLine(fmt.Sprintf("yllcorner %f", gd.Norig-float64(gd.Nrow)*gd.Cwidth))
	t.WriteLine(fmt.Sprintf("cellsize %f", gd.Cwidth))
	t.WriteLine(fmt.Sprintf("nodata_value %d", -9999))
	t.WriteLine(fmt.Sprintf("nbits %d", nbits))
	t.WriteLine(fmt.Sprintf("pixeltype %s", "float"))
	t.WriteLine(fmt.Sprintf("byteorder %s", "i"))
	t.WriteLine(fmt.Sprintf("layout %s", "bil"))
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
	if gd.Nact > 0 {
		m := make(map[int]bool, gd.Nact)
		for _, c := range gd.Sactives {
			m[c] = true
		}
		c := 0
		for i := 0; i < gd.Nrow; i++ {
			for j := 0; j < gd.Ncol; j++ {
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
		for i := 0; i < gd.Nrow; i++ {
			for j := 0; j < gd.Ncol; j++ {
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
	for i := 0; i < gd.Nrow; i++ {
		for j := 0; j < gd.Ncol; j++ {
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

func (gd *Definition) ToBIL(fp string, f32 []float32) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, f32); err != nil {
		panic(err)
	}
	if err := os.WriteFile(fp, buf.Bytes(), 0644); err != nil { // see: https://en.wikipedia.org/wiki/File_system_permissions
		panic(err)
	}
	if err := gd.ToHDRfloat(mmio.RemoveExtension(fp)+".hdr", 1, 32); err != nil {
		panic(err)
	}
}
