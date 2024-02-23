package grid

import (
	"fmt"

	"github.com/maseology/mmio"
)

func (r *Real) ImportRaster(fp string) error {
	switch mmio.GetExtension(fp) {
	case ".asc":
		return r.ImportAsc(fp)
	case ".bil":
		return r.ImportBil(fp)
	}
	return fmt.Errorf("unknown Raster type: %s", fp)
}
