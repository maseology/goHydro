package routing

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/maseology/mmaths"
	geojson "github.com/paulmach/go.geojson"
)

func PrintNetwork(fp string, nds []*mmaths.Node) {
	fc := geojson.NewFeatureCollection()
	for _, n := range nds {
		nd := n.I[0] // number of spatial dimensions recorded in set of vertices
		nv := len(n.S) / nd
		vs := make([][]float64, nv)
		for j := 0; j < nv; j++ {
			vs[j] = make([]float64, nd)
			for d := 0; d < nd; d++ {
				vs[j][d] = n.S[j*nd+d]
			}
		}
		ups, dns, dni := []int{}, []int{}, -1
		for _, u := range n.US {
			ups = append(ups, u.I[2])
		}
		for _, d := range n.DS {
			dns = append(dns, d.I[2])
			if d.I[0] > dni {
				dni = d.I[4]
			}
		}
		f := geojson.NewLineStringFeature(vs)
		f.SetProperty("segmentID", n.I[4])
		f.SetProperty("downID", dni)
		f.SetProperty("treeID", n.I[1])
		f.SetProperty("treesegID", n.I[2])
		f.SetProperty("topol", fmt.Sprintf("%d %d %d", ups, n.I[2], dns))
		f.SetProperty("order", n.I[3])
		fc.AddFeature(f)
	}
	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		log.Fatalf("routing.PrintSegments: %v\n", err)
	}
	if err := ioutil.WriteFile(fp, rawJSON, 0644); err != nil {
		log.Fatalf("routing.PrintSegments: %v\n", err)
	}
}
