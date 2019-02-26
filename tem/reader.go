package tem

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/maseology/mmaths"
)

type uhdemReader struct {
	I             int32
	X, Y, Z, S, A float64
}

func (u *uhdemReader) uhdemRead(b *bytes.Reader) {
	err := binary.Read(b, binary.LittleEndian, u)
	if err != nil {
		log.Fatalln("Fatal error: uhdemRead failed", err)
	}
}

func (u *uhdemReader) toTEC() (mmaths.Point, TEC) {
	var t TEC
	t.New(u.Z, u.S, u.A, -1)
	xy := mmaths.Point{X: u.X, Y: u.Y}
	return xy, t
}

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
