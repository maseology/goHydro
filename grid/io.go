package grid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"

	"github.com/maseology/mmio"
)

// ToASC creates an ascii-grid of Indx.
func (x *Indx) ToASC(fp string, ignoreActives bool) error {
	t, err := mmio.NewTXTwriter(fp)
	if err != nil {
		return fmt.Errorf("Indx ToASC: %v", err)
	}
	defer t.Close()
	x.GD.ToASCheader(t)
	if x.GD.Nact > 0 && ignoreActives {
		m := make(map[int]bool, x.GD.Nact)
		for _, c := range x.GD.Sactives {
			m[c] = true
		}
		c := 0
		for i := 0; i < x.GD.Nrow; i++ {
			for j := 0; j < x.GD.Ncol; j++ {
				if _, ok := m[c]; ok {
					t.Write(fmt.Sprintf("%d ", x.A[c]))
				} else {
					t.Write("-9999 ")
				}
				c++
			}
			t.Write("\n")
		}
	} else {
		c := 0
		for i := 0; i < x.GD.Nrow; i++ {
			for j := 0; j < x.GD.Ncol; j++ {
				if v, ok := x.A[c]; ok {
					t.Write(fmt.Sprintf("%d ", v))
				} else {
					t.Write("-9999 ")
				}
				c++
			}
			t.Write("\n")
		}
	}
	return nil
}

// ToBinary writes to binary array
func (x *Indx) ToBinary(fp string) error {
	a := make([]int32, x.GD.Nact)
	for i, c := range x.GD.Sactives {
		if xac, ok := x.A[c]; ok {
			a[i] = int32(xac)
		} else {
			return fmt.Errorf("Indx.ToBinary() error")
		}
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, a); err != nil {
		return fmt.Errorf("Indx.ToBinary() failed1: %v", err)
	}
	if err := ioutil.WriteFile(fp, buf.Bytes(), 0644); err != nil { // see: https://en.wikipedia.org/wiki/File_system_permissions
		return fmt.Errorf("Indx.ToBinary() failed2: %v", err)
	}
	return nil
}
