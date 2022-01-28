package routing

import (
	"log"
	"sort"

	"github.com/maseology/mmaths/slice"
	tp "github.com/maseology/mmaths/topology"
)

// SetDirectionFromRoots rts: Root nodes. elements of rts also belong to ntwk
func SetDirectionFromRoots(ntwk map[*tp.Node][]*tp.Node, rts []*tp.Node) {
	eval, cnt := make(map[*tp.Node]bool, len(ntwk)), 0
	var walkToJunctions func(*tp.Node)
	var w []*tp.Node
	walkToJunctions = func(n *tp.Node) {
		if _, ok := eval[n]; !ok {
			n.I = append(n.I, cnt)
			cnt++
		}
		eval[n] = false
		w = append(w, n)

		if len(ntwk[n]) == 2 {
			for _, c := range ntwk[n] {
				if v, ok := eval[c]; ok && !v {
					continue
				}
				walkToJunctions(c)
			}
		} else {
			eval[n] = true
		}
	}

	for _, r := range rts {
		queue := []*tp.Node{r} // push
		for len(queue) > 0 {
			q := queue[0] // pop
			queue = queue[1:]

			if _, ok := eval[q]; !ok {
				q.I = append(q.I, cnt)
				cnt++
			}
			eval[q] = false

			for _, c := range ntwk[q] {
				if _, ok := eval[c]; ok {
					continue
				}
				w = []*tp.Node{q} // initialize
				walkToJunctions(c)
				for i := 1; i < len(w); i++ {
					d, n := w[i-1], w[i]
					d.US = append(d.US, n)
					n.DS = append(n.DS, d)
				}
				queue = append(queue, w[len(w)-1])
			}
		}
	}
	for n := range ntwk {
		if len(n.I) != len(rts[0].I) {
			if len(n.I) > len(rts[0].I) {
				log.Fatalln("SetDirectionFromRoots dimensioning error")
			} else {
				n.I = append(n.I, -1) // isolated segments
			}
		}
	}
}

// Strahler, A.N., 1952. Hypsometric (area-altitude) analysis of erosional topology, Geological Society of America Bulletin 63(11): 1117–1142.
// the Horton–Strahler system: Horton, R.E., 1945. Erosional Development of Streams and Their Drainage Basins: Hydrophysical Approach To Quantitative Morphology Geological Society of America Bulletin, 56(3):275-370.
func Strahler(nodes []*tp.Node) {
	queue, nI := make([]*tp.Node, 0), -1
	for _, n := range nodes {
		n.I = append(n.I, 0)
		if nI == -1 {
			nI = len(n.I)
		} else if nI != len(n.I) {
			log.Fatalln(" Strahler error: dimensioning error")
		}
	}
	nI-- // to 0-index
	for _, ln := range tp.Leaves(nodes) {
		ln.I[nI] = 1
		queue = append(queue, ln) // sinks/leaves/headwaters
	}
	jns := tp.Junctions(nodes)
	isjn := make(map[*tp.Node]bool, len(jns))
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
							slice.Rev(uORD)
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

	// quick fix for tough cycles
	for _, ln := range tp.Leaves(nodes) {
		queue = append(queue, ln) // sinks/leaves/headwaters
	}
	for {
		if len(queue) == 0 {
			break
		}

		// pop
		q := queue[0]
		queue = queue[1:]

		for _, dn := range q.DS {
			if dn.I[nI] < q.I[nI] {
				dn.I[nI] = q.I[nI]
			}
			queue = append(queue, dn) // push
		}
	}
}

// Shreve R.L., 1966. Statistical Law of Stream Numbers. The Journal of Geology 74(1): 17-37.
func Shreve(nodes []*tp.Node) {
	queue, nI := make([]*tp.Node, 0), 0
	for _, n := range nodes {
		n.I = append(n.I, 0)
		nI = len(n.I)
	}
	nI-- // to 0-index
	for _, ln := range tp.Leaves(nodes) {
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
