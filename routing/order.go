package routing

import (
	"github.com/maseology/mmaths"
)

// SetDirectionFromRoots rts: Root nodes. elements of rts also belong to ntwk
func SetDirectionFromRoots(ntwk map[*mmaths.Node][]*mmaths.Node, rts []*mmaths.Node) {
	eval, cnt := make(map[*mmaths.Node]bool, len(ntwk)), 0
	var walkToJunctions func(*mmaths.Node)
	var w []*mmaths.Node
	walkToJunctions = func(n *mmaths.Node) {
		if v, ok := eval[n]; ok {
			if !v {
				print("")
			}
		} else {
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

			if v, ok := eval[q]; ok {
				if !v {
					print("")
				}
			} else {
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
				print("")
			} else {
				n.I = append(n.I, -1) // isolated segments
			}
		}
	}
}

// // SetOrder nds: All nodes; rts: Root nodes. elements of rts also belong to nsd
// func SetOrder(nds, rts []mmaths.Node) {
// 	for _, r := range rts {
// 		topologicalSort(nds, r) ///////////////////// go routine ??????????
// 	}
// }

// func topologicalSort(nds []mmaths.Node, startNode mmaths.Node) {
// 	queue := []mmaths.Node{startNode}
// 	ocol := make(map[*mmaths.Node]bool)
// 	for {
// 		if len(queue) == 0 {
// 			break
// 		}
// 		// pop
// 		nn := queue[0]
// 		queue = queue[1:]

// 		if _, ok := ocol[&nn]; ok {
// 			continue
// 		}
// 		walkToJuctions(nn)

// 	}
// 	// Dim osv = _o, cnt = 0
// 	// Dim ocoll As New Dictionary(Of Integer, Boolean)
// 	// With New Queue(Of Node)
// 	// 	.Enqueue(FromNode)
// 	// 	_o = New Dictionary(Of Integer, Boolean)
// 	// 	Do While .Count > 0
// 	// 		'_o = New Dictionary(Of Integer, Boolean)
// 	// 		Dim nn = .Dequeue
// 	// 		Dim nnid = _n.IndexOf(nn)
// 	// 		If ocoll.ContainsKey(nnid) Then Continue Do

// 	// 		Me.WalkToJunctions(nn)
// 	// 		For Each o In _o
// 	// 			If ocoll.ContainsKey(o.Key) Then Continue For
// 	// 			If _n(o.Key).Indices.Count = 0 Then _n(o.Key).Indices.Add(cnt)
// 	// 			If _n(o.Key).Indices(0) < cnt Then _n(o.Key).Indices(0) = cnt
// 	// 			If o.Value AndAlso _n(o.Key).Edges.Count > 1 Then .Enqueue(_n(o.Key)) Else ocoll.Add(o.Key, False)
// 	// 		Next
// 	// 		cnt += 1
// 	// 	Loop
// 	// End With
// 	// For Each o In osv
// 	// 	ocoll.Add(o.Key, o.Value)
// 	// Next
// 	// _o = ocoll
// }

// func walkToJuctions(n mmaths.Node) {
// 	// (*eval)[&n] = false
// 	// for _, un := range n.US {
// 	// 	if _, ok := (*eval)[&un]; ok {
// 	// 		continue
// 	// 	}
// 	// }
// 	i := 1
// 	_ = i
// }
