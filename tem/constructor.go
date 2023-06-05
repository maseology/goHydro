package tem

import (
	"fmt"
	"math"
	"path/filepath"
)

// NewTEM loads TEM
func NewTEM(fp string) (*TEM, error) {
	var t TEM
	err := t.New(fp)
	return &t, err
}

// New constructor
func (t *TEM) New(fp string) error {
	var err error
	var ds map[int]int // down-slope IDs = map[from]to
	switch filepath.Ext(fp) {
	case ".uhdem", ".bin":
		_, ds, err = t.loadUHDEM(fp)
		if err != nil {
			return err
		}
	case ".hdem":
		_, ds, err = t.loadHDEM(fp)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf(" error: unknown TEM file type used")
	}

	t.checkVals()
	t.buildUpslopes(ds)
	return nil
}

func (t *TEM) checkVals() {
	for k, v := range t.TEC {
		v1 := v
		if v.G < 0.0001 {
			v1.G = 0.0001
		}
		if v.A < -math.Pi {
			v1.A = 0.
		}
		t.TEC[k] = v1
	}
}
