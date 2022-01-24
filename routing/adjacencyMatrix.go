package routing

import (
	"fmt"
	"math"
)

// am: <0 connection at last node, >0 connection at first node (VERY SLOW)
func BuildAdjacencyMatrix(polylines [][][3]float64, searchRadius float64) (am [][]int, epfs, epls [][3]float64) {
	// collect endpoints
	sr2 := searchRadius * searchRadius
	epfs, epls = make([][3]float64, len(polylines)+1), make([][3]float64, len(polylines)+1)
	// tooshorts := map[int]bool{}
	for i, f := range polylines {
		epf, epl := f[0], f[len(f)-1]
		// if dist2(epf, epl) < sr2 {
		// 	tooshorts[i+1] = true
		// }
		epfs[i+1] = epf // first node connections
		epls[i+1] = epl // last node connections
	}

	// build Adjacency matrix: <0 connection at last node, >0 connection at first node
	am = make([][]int, len(polylines)+1)
	distinct := func(input [4]float64) []float64 {
		u := make([]float64, 0, 4)
		m := make(map[float64]bool)
		for _, val := range input {
			if _, ok := m[val]; !ok {
				m[val] = true
				u = append(u, val)
			}
		}
		return u
	}
	shortest := func(d [4]float64) (int, float64) {
		ix, fx := -1, math.MaxFloat64
		for i, f := range d {
			if f < fx {
				fx = f
				ix = i
			}
		}
		return ix, fx
	}
	for i := 1; i <= len(polylines); i++ { // need to offset from index-0
		am[i] = []int{}
		// if _, ok := tooshorts[i]; ok {
		// 	continue
		// }

		for j := 1; j <= len(polylines); j++ {
			// if _, ok := tooshorts[j]; ok {
			// 	continue
			// }

			d := [4]float64{}
			d[0] = dist2(epfs[i], epls[j])
			d[1] = dist2(epls[i], epfs[j])
			d[2] = dist2(epfs[i], epfs[j])
			d[3] = dist2(epls[i], epls[j])

			if len(distinct(d)) == 2 {
				// do nothing
				// crude. all simple cycles should have this property
			} else {
				if ii, f := shortest(d); f < sr2 {
					switch ii {
					case 0, 2:
						am[i] = append(am[i], j)
					case 1, 3:
						am[i] = append(am[i], -j)
					}
				}
			}
		}
	}
	fmt.Printf("  %d size of adjacency matrix - ", len(am))
	// printAdjacencyMatrixGeojson(polylines, am)
	return am, epfs, epls
}

// func printAdjacencyMatrixGeojson(polylines [][][3]float64, am [][]int) {
// 	fc := geojson.NewFeatureCollection()
// 	for i, pln := range polylines {
// 		ll := make([][]float64, len(pln))
// 		for j, vs := range pln {
// 			ll[j] = []float64{vs[0], vs[1]}
// 		}
// 		f := geojson.NewLineStringFeature(ll)
// 		f.SetProperty("ftid", i)
// 		f.SetProperty("topol", fmt.Sprintf("%d  %d", i+1, am[i+1]))
// 		fc.AddFeature(f)
// 	}

// 	rawJSON, err := fc.MarshalJSON()
// 	if err != nil {
// 		log.Fatalf("routing.Print: %v\n", err)
// 	}
// 	if err := ioutil.WriteFile("goAM.geojson", rawJSON, 0644); err != nil {
// 		log.Fatalf("routing.Print: %v\n", err)
// 	}
// }
