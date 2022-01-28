package routing

import (
	"io/ioutil"
	"log"

	tp "github.com/maseology/mmaths/topology"
	geojson "github.com/paulmach/go.geojson"
)

func LoadNetwork(fp string) []*tp.Node {
	fstreams, err := ioutil.ReadFile(fp)
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
		if f.Geometry.Type != "MultiLineString" {
			log.Fatalln("todo")
		}
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
				panic("TODO: indexing assumption not valid")
			}
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
