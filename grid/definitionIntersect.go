package grid

import (
	"log"
	"math"
)

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
	intsct := make(map[int][]int, gd.Na)
	if gd.Cw > toGD.Cw {
		log.Fatalf("Definition.Intersect TODO")
		log.Fatalf("Definition.Intersect: NNED TO CHECK CODE, not yet used.....")
		if math.Mod(gd.Cw, toGD.Cw) != 0. {
			log.Fatalf("Definition.Intersect error: Definitions grid definitions are not multiples: fromGD: %f, toGD: %f", gd.Cw, toGD.Cw)
		}
		scale := int(toGD.Cw / gd.Cw)
		for _, c := range gd.Sactives {
			i, j := gd.RowCol(c)
			tocid := toGD.CellID(i*scale, j*scale)
			intsct[c] = []int{tocid} // THIS IS INCONSISTENT ++++++++++++++++++++++++++++++++++++++++++++++++++++++
		}
	} else if gd.Cw < toGD.Cw {
		if math.Mod(toGD.Cw, gd.Cw) != 0. {
			log.Fatalf("Definition.Intersect error: Definitions grid definitions are not multiples: fromGD: %f, toGD: %f", gd.Cw, toGD.Cw)
		}
		scale := toGD.Cw / gd.Cw
		for _, c := range gd.Sactives {
			i, j := gd.RowCol(c)
			tocid := toGD.CellID(int(float64(i)/scale), int(float64(j)/scale))
			intsct[c] = []int{tocid}
		}
	} else {
		log.Fatalf("Definition.Intersect TODO")
	}
	return intsct
}
