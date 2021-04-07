package tem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/maseology/mmaths"
)

type hdemReader struct {
	Z, G, A float64
}

type uhdemReader struct {
	I             int32
	X, Y, Z, G, A float64
}

func (h *hdemReader) toTEC() TEC {
	var t TEC
	t.New(h.Z, h.G, h.A) //, -1)
	return t
}

func (u *uhdemReader) uhdemRead(b *bytes.Reader) {
	err := binary.Read(b, binary.LittleEndian, u)
	if err != nil {
		log.Fatalln("Fatal error: uhdemRead failed", err)
	}
}

func (u *uhdemReader) toTEC() (mmaths.Point, TEC) {
	var t TEC
	t.New(u.Z, u.G, u.A) //, -1)
	xy := mmaths.Point{X: u.X, Y: u.Y}
	return xy, t
}

// type fpReader struct {
// 	I, Nds int32
// 	Ids    []int32
// 	F      []float64
// }

type fpReader struct {
	I, Nds, Ids int32
	F           float64
}

func (f *fpReader) fpRead(b *bytes.Reader) error {
	err := binary.Read(b, binary.LittleEndian, f)
	if err != nil {
		return fmt.Errorf("Fatal error: fpRead failed: %v", err)
	}
	if f.Nds != 1 {
		return fmt.Errorf("Fatal error: fpRead currently only supports singular downslope IDs (i.e., tree-graph topology only). %v", err)
	}
	return nil
}
