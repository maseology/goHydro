package drainagenetwork

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/maseology/mmaths/topology"
	geojson "github.com/paulmach/go.geojson"
)

func topoString(n *topology.Node) string {
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

func SaveVertices(nodes []*topology.Node, fp string, fNames []string) {
	fc := geojson.NewFeatureCollection()
	for _, n := range nodes {
		if n.I[0] != 2 || len(n.S) != 2 {
			panic("tested only for nodes as points")
		}
		f := geojson.NewPointFeature([]float64{n.S[0], n.S[1]}) //, n.S[2]})
		// f.SetProperty("dim", n.I[0]) // dimension
		f.SetProperty("vertexID", n.I[1])
		// f.SetProperty("order", n.I[len(n.I)-1])
		fmt.Println(topoString(n))
		f.SetProperty("FromAtTo", topoString(n))
		// f.SetProperty("address", fmt.Sprintf("%p", n))

		for i, fn := range fNames {
			f.SetProperty(fn, n.I[i+2])
		}
		// for i := 2; i < len(n.I); i++ {
		// 	f.SetProperty(fmt.Sprintf("ID%02d", i), n.I[i])
		// }

		fc.AddFeature(f)
	}
	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		log.Fatalf("routing.Print: %v\n", err)
	}
	if err := os.WriteFile(fp, rawJSON, 0644); err != nil {
		log.Fatalf("routing.Print: %v\n", err)
	}
}

func SaveSegments(nodes []*topology.Node, fp string, fNames []string) {
	fc := geojson.NewFeatureCollection()
	for _, n := range nodes {
		if n.I[0] != 2 || len(n.S) != 2 {
			panic("tested only for nodes as points")
		}
		p0 := []float64{n.S[0], n.S[1]}
		for _, ds := range n.DS {
			p1 := []float64{ds.S[0], ds.S[1]}
			f := geojson.NewLineStringFeature([][]float64{p0, p1})
			// f.SetProperty("dim", n.I[0]) // dimension
			f.SetProperty("vertexID", n.I[1])
			f.SetProperty("dwnVrtID", ds.I[1])

			// f.SetProperty("order", n.I[len(n.I)-1])
			f.SetProperty("FromAtTo", topoString(n))
			// f.SetProperty("address", fmt.Sprintf("%p", n))

			for i, fn := range fNames {
				f.SetProperty(fn, n.I[i+2])
			}
			// for i := 2; i < len(n.I); i++ {
			// 	f.SetProperty(fmt.Sprintf("ID%02d", i), n.I[i])
			// }

			fc.AddFeature(f)
		}
	}
	rawJSON, err := fc.MarshalJSON()
	if err != nil {
		log.Fatalf("routing.Print: %v\n", err)
	}
	if err := os.WriteFile(fp, rawJSON, 0644); err != nil {
		log.Fatalf("routing.Print: %v\n", err)
	}
}

// func Print(nodes []*tp.Node, fp string) {
// 	fc := geojson.NewFeatureCollection()
// 	for _, n := range nodes {
// 		f := geojson.NewPointFeature([]float64{n.S[0], n.S[1]}) //, n.S[2]})
// 		f.SetProperty("featureID", n.I[0])
// 		f.SetProperty("nid", n.I[1])
// 		topo := func() string {
// 			lin := func() (o []int) {
// 				us := n.US
// 				o = make([]int, len(us))
// 				for i, n := range us {
// 					o[i] = n.I[1]
// 				}
// 				return
// 			}
// 			lout := func() (o []int) {
// 				ds := n.DS
// 				o = make([]int, len(ds))
// 				for i, n := range ds {
// 					o[i] = n.I[1]
// 				}
// 				return
// 			}
// 			return fmt.Sprint(lin()) + ">" + strconv.Itoa(n.I[1]) + ">" + fmt.Sprint(lout())
// 		}
// 		f.SetProperty("order", n.I[len(n.I)-1])
// 		f.SetProperty("topol", topo())
// 		f.SetProperty("address", fmt.Sprintf("%p", n))
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

// // func PrintWithCoords(nodes []tp.Node, coords [][3]float64, fp string) {
// // 	// csvw := mmio.NewCSVwriter(fp)
// // 	// csvw.WriteHead("nid,x,y,z,i")
// // 	// for i, n := range nodes {
// // 	// 	csvw.WriteLine(i, coords[i][0], coords[i][1], coords[i][2], n.ID)
// // 	// }
// // 	// csvw.Close()

// // 	fc := geojson.NewFeatureCollection()
// // 	for i, n := range nodes {
// // 		f := geojson.NewPointFeature([]float64{coords[i][0], coords[i][1]}) //, coords[i][2]})
// // 		f.SetProperty("nid", i)
// // 		topo := func() string {
// // 			lin := func() (o []int) {
// // 				us := n.US
// // 				o = make([]int, len(us))
// // 				for i, n := range us {
// // 					o[i] = n.I[0]
// // 				}
// // 				return
// // 			}
// // 			lout := func() (o []int) {
// // 				ds := n.DS
// // 				o = make([]int, len(ds))
// // 				for i, n := range ds {
// // 					o[i] = n.I[0]
// // 				}
// // 				return
// // 			}
// // 			return fmt.Sprint(lin()) + ">" + strconv.Itoa(i) + ">" + fmt.Sprint(lout())
// // 		}
// // 		f.SetProperty("order", n.I[1])
// // 		f.SetProperty("topol", topo())
// // 		fc.AddFeature(f)
// // 	}
// // 	rawJSON, err := fc.MarshalJSON()
// // 	if err != nil {
// // 		log.Fatalf("routing.Print: %v\n", err)
// // 	}
// // 	if err := ioutil.WriteFile(fp, rawJSON, 0644); err != nil {
// // 		log.Fatalf("routing.Print: %v\n", err)
// // 	}
// // }
