package routing

import (
	"log"

	"github.com/maseology/mmaths"
)

// SetDirectionFromRoots rts: Root nodes. elements of rts also belong to ntwk
func SetDirectionFromRoots(ntwk map[*mmaths.Node][]*mmaths.Node, rts []*mmaths.Node) {
	eval, cnt := make(map[*mmaths.Node]bool, len(ntwk)), 0
	var walkToJunctions func(*mmaths.Node)
	var w []*mmaths.Node
	walkToJunctions = func(n *mmaths.Node) {
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
		queue := []*mmaths.Node{r} // push
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
				w = []*mmaths.Node{q} // initialize
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
