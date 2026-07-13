package drainagenetwork

import (
	"log"
	"os"

	tp "github.com/maseology/mmaths/topology"
	geojson "github.com/paulmach/go.geojson"
)

type Basin struct {
	Name                           string
	OutletSegID, TreeID, TreeSegID int
}

func LoadNetwork(fp string) ([]*tp.Node, map[int]Basin) {
	fstreams, err := os.ReadFile(fp)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	gstreams, err := geojson.UnmarshalFeatureCollection(fstreams)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	var nds []*tp.Node
	basins := make(map[int]Basin)
	nf := 0
	for _, f := range gstreams.Features {
		switch f.Geometry.Type {
		case "LineString":
			ff := f.Geometry.LineString
			dim := len(ff[0])

			nwsid := f.Properties["WsID"]
			if nwsid != nil {
				basins[int(nwsid.(float64))] = Basin{
					Name:        f.Properties["newWsname"].(string),
					OutletSegID: int(f.Properties["WsOut"].(float64)),
					TreeID:      int(f.Properties["treeID"].(float64)),
					TreeSegID:   int(f.Properties["treesegID"].(float64)),
				}
			}

			nds = append(nds, &tp.Node{
				S: func() []float64 {
					a := make([]float64, len(ff)*dim)
					for i, c := range ff {
						for d := range dim {
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
							for d := range 2 {
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

	// basinTableJSON(basins, fp[:len(fp)-8]+".json")

	// topological sort
	for ius := range nds {
		ids := nds[ius].I[2]
		if ids == -1 {
			continue
		}
		nds[ius].DS = append(nds[ius].DS, nds[ids])
		nds[ids].US = append(nds[ids].US, nds[ius])
	}

	return nds, basins
}

// type basinRow struct {
// 	WsID   int    `json:"wsID"`
// 	WsName string `json:"wsName"`
// 	Outlet int    `json:"outlet"`
// }

// func basinTableJSON(basins map[int]Basin, filename string) error {
// 	rows := make([]basinRow, 0, len(basins))
// 	for wsID, b := range basins {
// 		rows = append(rows, basinRow{
// 			WsID:   wsID,
// 			WsName: b.Name,
// 			Outlet: b.OutletSegID,
// 		})
// 	}

// 	data, err := json.MarshalIndent(rows, "", "  ")
// 	if err != nil {
// 		return err
// 	}

// 	return os.WriteFile(filename, data, 0644)
// }
