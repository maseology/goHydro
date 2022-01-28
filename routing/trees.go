package routing

import tp "github.com/maseology/mmaths/topology"

func JunctionToJunctionFromRoots(roots []*tp.Node) []*tp.Node {
	o := []*tp.Node{}
	for itree, r := range roots {
		sws := r.Climb()

		var xys [][][]float64

		queue := []*tp.Node{r}

		// find junction and leaves
		jns := tp.Junctions(sws)
		isjn := make(map[*tp.Node]bool, len(jns))
		for _, jn := range jns {
			isjn[jn] = true
			queue = append(queue, jn) // push
		}

		lvs := tp.Leaves(sws)
		islf := make(map[*tp.Node]bool, len(lvs))
		for _, lf := range lvs {
			islf[lf] = true
		}

		walkUpstreamToJuntion := func(startNode *tp.Node) (*tp.Node, [][]float64) {
			xys := [][]float64{}
			var nlast *tp.Node
			var recurs func(*tp.Node)
			recurs = func(n *tp.Node) {
				xys = append(xys, []float64{n.S[0], n.S[1], n.S[2]})
				nlast = n
				if !isjn[n] && !islf[n] {
					recurs(n.US[0])
				}
			}
			recurs(startNode)
			return nlast, xys
		}

		isegtree := 0
		ups, dns := map[*tp.Node][]int{}, map[*tp.Node][]int{}
		for len(queue) > 0 {
			q := queue[0] // pop
			queue = queue[1:]

			for _, u := range q.US {
				xys = append(xys, [][]float64{{q.S[0], q.S[1], q.S[2]}})
				nl, nxys := walkUpstreamToJuntion(u)
				dns[nl] = append(dns[nl], isegtree)
				ups[q] = append(ups[q], isegtree)
				xys[isegtree] = append(xys[isegtree], nxys...)
				isegtree++
			}
		}

		// create tree (segments as nodes)
		oo := make([]*tp.Node, len(xys))
		for isegtree, v := range xys {
			lverts := make([]float64, 2*len(v))
			for i, c := range v {
				lverts[2*i] = c[0]
				lverts[2*i+1] = c[1]
			}
			oo[isegtree] = &tp.Node{
				I: []int{2, itree, isegtree}, // dimensionned at 2 (dropping elevation)
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
		// sca := mmio.NewCSVwriter("junction.vertices.csv")
		// sca.WriteHead("x,y,jid,dwn")
		// for i, jn := range jns {
		// 	sca.WriteLine(jn.S[0], jn.S[1], i, fmt.Sprintf("%d %d>%d", i, ups[jn], dns[jn]))
		// }
		// sca.Close()

		// // print for testing
		// Strahler(oo)
		// PrintNetwork("M:/segments.geojson", oo)
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
