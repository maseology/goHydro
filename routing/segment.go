package routing

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/maseology/mmaths"
	geojson "github.com/paulmach/go.geojson"
)

func BuildSegment(roots []*mmaths.Node) []*mmaths.Node {
	o := []*mmaths.Node{}
	for itree, r := range roots {
		sws := r.Climb()

		var xys [][][]float64

		queue := []*mmaths.Node{r}

		// find junction and leaves
		jns := mmaths.Junctions(sws)
		isjn := make(map[*mmaths.Node]bool, len(jns))
		for _, jn := range jns {
			isjn[jn] = true
			queue = append(queue, jn) // push
		}

		lvs := mmaths.Leaves(sws)
		islf := make(map[*mmaths.Node]bool, len(lvs))
		for _, lf := range lvs {
			islf[lf] = true
		}

		walkUpstreamToJuntion := func(startNode *mmaths.Node) (*mmaths.Node, [][]float64) {
			xys := [][]float64{}
			var nlast *mmaths.Node
			var recurs func(*mmaths.Node)
			recurs = func(n *mmaths.Node) {
				xys = append(xys, []float64{n.S[0], n.S[1], n.S[2]})
				nlast = n
				if !isjn[n] && !islf[n] {
					recurs(n.US[0])
				}
			}
			recurs(startNode)
			return nlast, xys
		}

		iseg := 0
		ups, dns := map[*mmaths.Node][]int{}, map[*mmaths.Node][]int{}
		for len(queue) > 0 {
			q := queue[0] // pop
			queue = queue[1:]

			for _, u := range q.US {
				xys = append(xys, [][]float64{{q.S[0], q.S[1], q.S[2]}})
				nl, nxys := walkUpstreamToJuntion(u)
				dns[nl] = append(dns[nl], iseg)
				ups[q] = append(ups[q], iseg)
				xys[iseg] = append(xys[iseg], nxys...)
				iseg++
			}
		}

		// create tree (segments as nodes)
		oo := make([]*mmaths.Node, len(xys))
		for iseg, v := range xys {
			lverts := make([]float64, 2*len(v))
			for i, c := range v {
				lverts[2*i] = c[0]
				lverts[2*i+1] = c[1]
			}
			oo[iseg] = &mmaths.Node{
				I: []int{2, itree, iseg}, // dimensionned at 2
				S: lverts,
			}
		}
		// topology
		for _, jn := range jns {
			for _, u := range ups[jn] {
				for _, d := range dns[jn] {
					oo[u].DS = append(oo[u].DS, oo[d])
					oo[d].US = append(oo[d].US, oo[u])
				}
			}
		}
		o = append(o, oo...)

		// //////////////////////////////////////
		// // print for testing
		// sca := mmio.NewCSVwriter("M:/segments.vertices.csv")
		// sca.WriteHead("x,y,jid,dwn")
		// for i, jn := range jns {
		// 	sca.WriteLine(jn.S[0], jn.S[1], i, fmt.Sprintf("%d %d>%d", i, ups[jn], dns[jn]))
		// }
		// sca.Close()

		// // print for testing
		// Strahler(oo)
		// PrintSegments("M:/segments.geojson", oo)
		// // fc := geojson.NewFeatureCollection()
		// // for i, vs := range xys {
		// // 	f := geojson.NewLineStringFeature(vs)
		// // 	f.SetProperty("segmentID", i)
		// // 	fc.AddFeature(f)
		// // }
		// // rawJSON, err := fc.MarshalJSON()
		// // if err != nil {
		// // 	log.Fatalf("routing.Print: %v\n", err)
		// // }
		// // if err := ioutil.WriteFile("M:/segments.geojson", rawJSON, 0644); err != nil {
		// // 	log.Fatalf("routing.Print: %v\n", err)
		// // }

	}
	return o
}

func PrintSegments(fp string, nds []*mmaths.Node) {
	fc := geojson.NewFeatureCollection()
	for i, n := range nds {
		nd := n.I[0]
		nv := len(n.S) / nd
		vs := make([][]float64, nv)
		for j := 0; j < nv; j++ {
			vs[j] = make([]float64, nd)
			for d := 0; d < nd; d++ {
				vs[j][d] = n.S[j*nd+d]
			}
		}
		ups, dns := []int{}, []int{}
		for _, u := range n.US {
			ups = append(ups, u.I[2])
		}
		for _, d := range n.DS {
			dns = append(dns, d.I[2])
		}
		f := geojson.NewLineStringFeature(vs)
		f.SetProperty("treeID", n.I[1])
		f.SetProperty("segmentID", i)
		f.SetProperty("topol", fmt.Sprintf("%d %d %d", ups, i, dns))
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
