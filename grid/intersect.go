package grid

import (
	"log"
	"math"
)

// Intersect returns a mapping from current Definition to inputted Definition
// for now, only Definitions that share the same origin, and cell sizes are mulitples can be considered
func (gd *Definition) Intersect(toGD *Definition) map[int][]int {
	// checks
	if gd.Eorig != toGD.Eorig || gd.Norig != toGD.Norig {
		log.Fatalf("Definition.Intersect error: Definitions must have the same origin")
	}
	if gd.Rotation != toGD.Rotation {
		log.Fatalf("Definition.Intersect error: Definitions not in same orientation (i.e., rotation)")
	}
	intsct := make(map[int][]int, gd.Nact)
	if gd.Cwidth > toGD.Cwidth {
		log.Fatalf("Definition.Intersect TODO")
		log.Fatalf("Definition.Intersect: NNED TO CHECK CODE, not yet used.....")
		if math.Mod(gd.Cwidth, toGD.Cwidth) != 0. {
			log.Fatalf("Definition.Intersect error: Definitions grid definitions are not multiples: fromGD: %f, toGD: %f", gd.Cwidth, toGD.Cwidth)
		}
		scale := int(toGD.Cwidth / gd.Cwidth)
		for _, c := range gd.Sactives {
			i, j := gd.RowCol(c)
			tocid := toGD.CellID(i*scale, j*scale)
			intsct[c] = []int{tocid} // THIS IS INCONSISTENT ++++++++++++++++++++++++++++++++++++++++++++++++++++++
		}
	} else if gd.Cwidth < toGD.Cwidth {
		if math.Mod(toGD.Cwidth, gd.Cwidth) != 0. {
			log.Fatalf("Definition.Intersect error: Definitions grid definitions are not multiples: fromGD: %f, toGD: %f", gd.Cwidth, toGD.Cwidth)
		}
		scale := toGD.Cwidth / gd.Cwidth
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
