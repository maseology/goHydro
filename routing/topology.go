package routing

import (
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strconv"

	"github.com/maseology/mmaths"
	geojson "github.com/paulmach/go.geojson"
)

func Print(nodes []*mmaths.Node, fp string) {
	fc := geojson.NewFeatureCollection()
	for _, n := range nodes {
		f := geojson.NewPointFeature([]float64{n.S[0], n.S[1]}) //, n.S[2]})
		f.SetProperty("featureID", n.I[0])
		f.SetProperty("nid", n.I[1])
		topo := func() string {
			lin := func() (o []int) {
				us := n.US
				o = make([]int, len(us))
				for i, n := range us {
					o[i] = n.I[1]
				}
				return
			}
			lout := func() (o []int) {
				ds := n.DS
				o = make([]int, len(ds))
				for i, n := range ds {
					o[i] = n.I[1]
				}
				return
			}
			return fmt.Sprint(lin()) + ">" + strconv.Itoa(n.I[1]) + ">" + fmt.Sprint(lout())
		}
		f.SetProperty("order", n.I[len(n.I)-1])
		f.SetProperty("topol", topo())
		f.SetProperty("address", fmt.Sprintf("%p", n))
		fc.AddFeature(f)
	}
	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		log.Fatalf("routing.Print: %v\n", err)
	}
	if err := ioutil.WriteFile(fp, rawJSON, 0644); err != nil {
		log.Fatalf("routing.Print: %v\n", err)
	}
}

// func PrintWithCoords(nodes []mmaths.Node, coords [][3]float64, fp string) {
// 	// csvw := mmio.NewCSVwriter(fp)
// 	// csvw.WriteHead("nid,x,y,z,i")
// 	// for i, n := range nodes {
// 	// 	csvw.WriteLine(i, coords[i][0], coords[i][1], coords[i][2], n.ID)
// 	// }
// 	// csvw.Close()

// 	fc := geojson.NewFeatureCollection()
// 	for i, n := range nodes {
// 		f := geojson.NewPointFeature([]float64{coords[i][0], coords[i][1]}) //, coords[i][2]})
// 		f.SetProperty("nid", i)
// 		topo := func() string {
// 			lin := func() (o []int) {
// 				us := n.US
// 				o = make([]int, len(us))
// 				for i, n := range us {
// 					o[i] = n.I[0]
// 				}
// 				return
// 			}
// 			lout := func() (o []int) {
// 				ds := n.DS
// 				o = make([]int, len(ds))
// 				for i, n := range ds {
// 					o[i] = n.I[0]
// 				}
// 				return
// 			}
// 			return fmt.Sprint(lin()) + ">" + strconv.Itoa(i) + ">" + fmt.Sprint(lout())
// 		}
// 		f.SetProperty("order", n.I[1])
// 		f.SetProperty("topol", topo())
// 		fc.AddFeature(f)
// 	}
// 	rawJSON, err := fc.MarshalJSON()
// 	if err != nil {
// 		log.Fatalf("routing.Print: %v\n", err)
// 	}
// 	if err := ioutil.WriteFile(fp, rawJSON, 0644); err != nil {
// 		log.Fatalf("routing.Print: %v\n", err)
// 	}
// }

// Strahler, A.N., 1952. Hypsometric (area-altitude) analysis of erosional topology, Geological Society of America Bulletin 63(11): 1117–1142.
// the Horton–Strahler system: Horton, R.E., 1945. Erosional Development of Streams and Their Drainage Basins: Hydrophysical Approach To Quantitative Morphology Geological Society of America Bulletin, 56(3):275-370.
func Strahler(nodes []*mmaths.Node) {
	queue, nI := make([]*mmaths.Node, 0), -1
	for _, n := range nodes {
		n.I = append(n.I, 0)
		if nI == -1 {
			nI = len(n.I)
		} else if nI != len(n.I) {
			log.Fatalln(" Strahler error: dimensioning error")
		}
	}
	nI-- // to 0-index
	for _, ln := range mmaths.Leaves(nodes) {
		ln.I[nI] = 1
		queue = append(queue, ln) // sinks/leaves/headwaters
	}
	jns := mmaths.Junctions(nodes)
	isjn := make(map[*mmaths.Node]bool, len(jns))
	for _, jn := range jns {
		isjn[jn] = true
	}

	for {
		if len(queue) == 0 {
			break
		}

		// pop
		q := queue[0]
		queue = queue[1:]

		if len(q.DS) > 1 { // bifurcating (assuming cycle)
			for _, dn := range q.DS {
				if q.I[nI] < 0 {
					dn.I[nI] = q.I[nI] // consecutive cycles
				} else {
					dn.I[nI] = -q.I[nI]
				}
				queue = append(queue, dn) // push
			}
		} else {
			for _, dn := range q.DS {
				if _, ok := isjn[dn]; ok {
					uORD := []int{}
					for _, un := range dn.US {
						if un.I[nI] == 0 {
							uORD = []int{}
							break
						}
						uORD = append(uORD, un.I[nI])
					}
					if len(uORD) == 1 {
						dn.I[nI] = q.I[nI]
						for _, dn := range dn.DS { // bifurcating (cycle?)
							dn.I[nI] = -q.I[nI]
							queue = append(queue, dn) // push
						}
					} else if len(uORD) > 1 { // merging
						sort.Ints(uORD)
						if uORD[0] < 0 { // cycle
							dn.I[nI] = -uORD[0]
						} else {
							mmaths.Rev(uORD)
							if uORD[0] == uORD[1] {
								dn.I[nI] = uORD[0] + 1
							} else {
								dn.I[nI] = uORD[0]
							}
						}
						queue = append(queue, dn) // push
					}
				} else {
					dn.I[nI] = q.I[nI]
					queue = append(queue, dn) // push
				}
			}
		}
	}

	for _, n := range nodes {
		if n.I[nI] < 0 {
			n.I[nI] = -n.I[nI]
		}
	}
}

// Shreve R.L., 1966. Statistical Law of Stream Numbers. The Journal of Geology 74(1): 17-37.
func Shreve(nodes []*mmaths.Node) {
	queue, nI := make([]*mmaths.Node, 0), 0
	for _, n := range nodes {
		n.I = append(n.I, 0)
		nI = len(n.I)
	}
	nI-- // to 0-index
	for _, ln := range mmaths.Leaves(nodes) {
		ln.I[nI] = 1
		queue = append(queue, ln) // sinks/leaves/headwaters
	}

	for {
		if len(queue) == 0 {
			break
		}

		// pop
		x := queue[0]
		queue = queue[1:]

		// push
		for _, dn := range x.DS {
			dn.I[nI] += 1
			queue = append(queue, dn)
		}
	}

	// norder := make([]int, len(nodes))
	// for i, n := range nodes {
	// 	norder[i] = n.I[nI]
	// }
	// return norder
}
