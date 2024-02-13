package routing

import (
	"log"
	"os"

	tp "github.com/maseology/mmaths/topology"
	geojson "github.com/paulmach/go.geojson"
)

func LoadNetwork(fp string) []*tp.Node {
	fstreams, err := os.ReadFile(fp)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	gstreams, err := geojson.UnmarshalFeatureCollection(fstreams)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	var nds []*tp.Node

	nf := 0
	for _, f := range gstreams.Features {
		switch f.Geometry.Type {
		case "LineString":
			ff := f.Geometry.LineString
			dim := len(ff[0])
			nds = append(nds, &tp.Node{
				S: func() []float64 {
					a := make([]float64, len(ff)*dim)
					for i, c := range ff {
						for d := 0; d < dim; d++ {
							a[i*dim+d] = c[d]
						}
					}
					return a
				}(),
				I: []int{
					dim, // dimension
					int(f.Properties["segmentID"].(float64)),
					int(f.Properties["downID"].(float64)),
					int(f.Properties["treeID"].(float64)),
					int(f.Properties["treesegID"].(float64)),
					int(f.Properties["order"].(float64)),
				},
			})
			if nds[nf].I[1] != nf {
				panic("TODO: LineString indexing assumption not valid")
			}
			nf += 1
		case "MultiLineString":
			for i, ln := range f.Geometry.MultiLineString {
				nf += len(ln)
				nds = append(nds, &tp.Node{
					S: func() []float64 {
						a := make([]float64, len(ln)*2)
						for i, c := range ln {
							for d := 0; d < 2; d++ {
								a[i*2+d] = c[d]
							}
						}
						return a
					}(),
					I: []int{
						2, // dimension
						int(f.Properties["segmentID"].(float64)),
						int(f.Properties["downID"].(float64)),
						int(f.Properties["treeID"].(float64)),
						int(f.Properties["treesegID"].(float64)),
						int(f.Properties["order"].(float64)),
					},
				})
				if nds[i].I[1] != i {
					panic("TODO: MultiLineString indexing assumption not valid")
				}
			}
		default:
			log.Fatalf("Routing.LoadNetwork: unsupported type, given %v\n", f.Geometry.Type)
		}
	}

	// topological sort
	for ius := range nds {
		ids := nds[ius].I[2]
		if ids == -1 {
			continue
		}
		nds[ius].DS = append(nds[ius].DS, nds[ids])
		nds[ids].US = append(nds[ids].US, nds[ius])
	}

	return nds
}
