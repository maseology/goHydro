package routing

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"

	"github.com/maseology/mmaths"
	geojson "github.com/paulmach/go.geojson"
)

type edge struct {
	n0, n1 *mmaths.Node
	cond   float64
}

// CollectNodes takes the spatial arrangement of polyline features and determine head/tail connectivity (topology with no direction/ordering)
// returns map[node][nodes node is connected to]; by "node" I mean pointers to nodes
func CollectNodes(polylines [][][3]float64, searchRadius float64) map[*mmaths.Node][]*mmaths.Node {
	am, enf, enl := buildAdjacencyMatrix(polylines, searchRadius)
	// printAdjacencyMatrix(polylines, am) // for testing
	eval := make(map[int]bool, len(am))
	nvert := func() int {
		i := 0
		for _, f := range polylines {
			i += len(f)
		}
		return i
	}()

	edges := make([]*edge, 0, nvert)
	var collectEdges func(int, int)
	collectEdges = func(aid int, fromid int) {
		eval[aid] = false
		blF := func() bool {
			if fromid != 0 {
				for _, c := range am[aid] {
					if c == fromid { // previous connected to my start point
						return true
					}
					if -c == fromid { // previous connected to my end point, count backwards
						return false
					}
				}
				log.Fatalf("routing.CollectNodes: this line should not occur %d %d %d\n", aid, fromid, am[aid])
			}
			return am[aid][0] < 0 // first connection I have is at the end of feature aid, then forward we go
		}()

		// create nodes
		if blF { // forward progression
			if fromid > 0 && enf[aid] != enf[fromid] && enf[aid] != enl[fromid] {
				fmt.Printf("  warning: segments %d and %d do not share nodes (forward progression)\n", aid, fromid)
			}
			n0 := enf[aid]
			for i := 1; i < len(polylines[aid-1])-2; i++ {
				n1 := &mmaths.Node{
					I: []int{aid},
					S: polylines[aid-1][i][:],
				}
				edges = append(edges, &edge{n0: n0, n1: n1})
				n0 = n1
			}
			if n0 != enl[aid] {
				edges = append(edges, &edge{n0: n0, n1: enl[aid]})
			}
		} else { // backward progression
			if fromid > 0 && enl[aid] != enf[fromid] && enl[aid] != enl[fromid] {
				fmt.Printf("  warning: segments %d and %d do not share nodes (backward progression)\n", aid, fromid)
			}
			n0 := enl[aid]
			for i := len(polylines[aid-1]) - 2; i > 0; i-- {
				n1 := &mmaths.Node{
					I: []int{aid},
					S: polylines[aid-1][i][:],
				}
				edges = append(edges, &edge{n0: n0, n1: n1})
				n0 = n1
			}
			if n0 != enf[aid] {
				edges = append(edges, &edge{n0: n0, n1: enf[aid]})
			}
		}

		// recursion
		for _, c := range am[aid] {
			if c < 0 {
				if _, ok := eval[-c]; ok {
					continue
				}
				if blF {
					collectEdges(-c, aid)
				}
			} else {
				if _, ok := eval[c]; ok {
					continue
				}
				if !blF {
					collectEdges(c, aid)
				}
			}
		}
	}

	for i, a := range am {
		if i == 0 {
			continue
		}
		if len(a) == 0 {
			continue
		}
		if v, ok := eval[i]; ok && !v {
			continue
		}
		for j := 1; j < len(a); j++ {
			if a[j-1]*a[j] <= 0 {
				//fmt.Println("routing.CollectNode: check: may need increase node threshold")
				continue
			}
		}

		// collect segments (edges)
		collectEdges(i, 0)
	}

	// return connectivity
	return func() map[*mmaths.Node][]*mmaths.Node {
		ne := make(map[*mmaths.Node][]*mmaths.Node, len(edges)*2)
		for _, ed := range edges {
			if _, ok := ne[ed.n0]; !ok {
				ne[ed.n0] = []*mmaths.Node{ed.n1}
			} else {
				ne[ed.n0] = append(ne[ed.n0], ed.n1)
			}
			if _, ok := ne[ed.n1]; !ok {
				ne[ed.n1] = []*mmaths.Node{ed.n0}
			} else {
				ne[ed.n1] = append(ne[ed.n1], ed.n0)
			}
		}
		return ne
	}()
}

// am: <0 connection at last node, >0 connection at first node
func buildAdjacencyMatrix(polylines [][][3]float64, searchRadius float64) ([][]int, []*mmaths.Node, []*mmaths.Node) {
	// collect endpoints
	sr2 := searchRadius * searchRadius
	// tooshorts := map[int]bool{}
	dist2 := func(p0, p1 [3]float64) float64 {
		s2 := 0.
		for j := 0; j < 3; j++ {
			ds := p0[j] - p1[j]
			s2 += ds * ds
		}
		return s2
	}
	epfs, epls := make([][3]float64, len(polylines)+1), make([][3]float64, len(polylines)+1)
	for i, f := range polylines {
		epf, epl := f[0], f[len(f)-1]
		// if dist2(epf, epl) < sr2 {
		// 	tooshorts[i+1] = true
		// }
		epfs[i+1] = epf // first node connections
		epls[i+1] = epl // last node connections
	}

	// build Adjacency matrix: <0 connection at last node, >0 connection at first node
	am := make([][]int, len(polylines)+1)
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

	// merge nodes where intersection is detected
	enf, enl := make([]*mmaths.Node, len(polylines)+1), make([]*mmaths.Node, len(polylines)+1)
	for i := 1; i <= len(polylines); i++ {
		if len(am[i]) == 0 {
			continue
		}
		if len(am[i]) == 1 && am[i][0] == i {
			am[i] = []int{}
			continue
		}
		if enf[i] == nil {
			enf[i] = &mmaths.Node{I: []int{i}, S: epfs[i][:]}
		}
		if enl[i] == nil {
			enl[i] = &mmaths.Node{I: []int{i}, S: epls[i][:]}
		}

		if enf[i] == enl[i] {
			fmt.Printf("  warning: segment %d non routing: enf==enl\n", i)
		} else {
			ddd := dist2([...]float64{enf[i].S[0], enf[i].S[1], 0.}, [...]float64{enl[i].S[0], enl[i].S[1], 0.})
			if ddd < .001 {
				fmt.Printf("  warning: segment %d non routing: start point = endpoint, d = %f\n", i, ddd)
			}
		}

		for _, c := range am[i] {
			ca, blN := func() (int, bool) { // false: c connected at the end of am[i]
				if c < 0 {
					return -c, true
				}
				return c, false
			}()
			if ca < i {
				continue
			}

			blF := func() bool { // true: progression follows from first node to last
				for _, cc := range am[ca] {
					if cc == i {
						return true
					}
				}
				return false
			}()

			if blN {
				if blF {
					if enf[ca] == nil {
						enf[ca] = enl[i]
					} else {
						enl[i] = enf[ca]
					}
				} else {
					if enl[ca] == nil {
						enl[ca] = enl[i]
					} else {
						enl[i] = enl[ca]
					}
				}
			} else {
				if blF {
					if enf[ca] == nil {
						enf[ca] = enf[i]
					} else {
						enf[i] = enf[ca]
					}
				} else {
					if enl[ca] == nil {
						enl[ca] = enf[i]
					} else {
						enf[i] = enl[ca]
					}
				}
			}
		}
	}

	return am, enf, enl
}

func printAdjacencyMatrix(polylines [][][3]float64, am [][]int) {
	fc := geojson.NewFeatureCollection()
	for i, pln := range polylines {
		ll := make([][]float64, len(pln))
		for j, vs := range pln {
			ll[j] = []float64{vs[0], vs[1]}
		}
		f := geojson.NewLineStringFeature(ll)
		f.SetProperty("ftid", i)
		f.SetProperty("topol", fmt.Sprintf("%d  %d", i+1, am[i+1]))
		fc.AddFeature(f)
	}

	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		log.Fatalf("routing.Print: %v\n", err)
	}
	if err := ioutil.WriteFile("M:/goAM.geojson", rawJSON, 0644); err != nil {
		log.Fatalf("routing.Print: %v\n", err)
	}
}
