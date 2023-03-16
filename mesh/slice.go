package mesh

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/maseology/mmio"
)

// Slice struct of a uniform grid
type Slice struct {
	Nodes    [][]float64
	Elements [][]int
	Name     string
}

// // NewSlice constructs a basic fem mesh
// func NewSlice(nam string, nr, nc int, UniformCellSize float64) *Slice {
// 	var gd Slice
// 	gd.Name = nam
// 	gd.Nrow, gd.Ncol, gd.Nact = nr, nc, nr*nc
// 	gd.Cwidth = UniformCellSize
// 	gd.Sactives = make([]int, gd.Nact)
// 	gd.act = make(map[int]bool, gd.Nact)
// 	for i := 0; i < gd.Nact; i++ {
// 		gd.Sactives[i] = i
// 		gd.act[i] = true
// 	}
// 	gd.Coord = make(map[int]mmaths.Point, gd.Nact)
// 	cid := 0
// 	for i := 0; i < gd.Nrow; i++ {
// 		for j := 0; j < gd.Ncol; j++ {
// 			p := mmaths.Point{X: gd.Eorig + UniformCellSize*(float64(j)+0.5), Y: gd.Norig - UniformCellSize*(float64(i)+0.5)}
// 			gd.Coord[cid] = p
// 			cid++
// 		}
// 	}
// 	return &gd
// }

// ReadAlgomesh imports a Algomesh grids in .ah2 or .ah3
func ReadAlgomesh(fp string, print bool) (*Slice, error) {
	file, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("ReadGDEF: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	line, _, err := reader.ReadLine()
	if err != nil {
		return nil, fmt.Errorf("ReadTextLines: %v", err)
	}
	nn, err := strconv.ParseInt(string(line), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ReadTextLines: %v", err)
	}

	f64 := func(s string) float64 {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic("ReadAlgomesh f64")
		}
		return f
	}
	i64 := func(s string) int {
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			panic("ReadAlgomesh i64")
		}
		return int(i)
	}

	nds := make([][]float64, nn)
	for i := 0; i < int(nn); i++ {
		line, _, err := reader.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("ReadTextLines: %v", err)
		}
		sp := strings.Split(string(line), " ")
		if len(sp) != 3 {
			panic("ReadAlgomesh 1")
		}
		nds[i] = []float64{f64(sp[0]), f64(sp[1]), f64(sp[2])}
	}

	ne, err := strconv.ParseInt(string(line), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ReadTextLines: %v", err)
	}

	els := make([][]int, ne)
	for i := 0; i < int(ne); i++ {
		line, _, err := reader.ReadLine()
		if err != nil {
			return nil, fmt.Errorf("ReadTextLines: %v", err)
		}
		sp := strings.Split(string(line), " ")
		if len(sp) != 3 {
			panic("ReadAlgomesh 2")
		}
		els[i] = []int{i64(sp[0]), i64(sp[1]), i64(sp[2])}
	}

	if _, _, err := reader.ReadLine(); err != io.EOF {
		panic("ReadAlgomesh 3")
	}

	return &Slice{Name: mmio.FileName(fp, false), Nodes: nds, Elements: els}, nil
}
