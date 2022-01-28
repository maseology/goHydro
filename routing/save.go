package routing

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
