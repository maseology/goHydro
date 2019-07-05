package swat

import (
	"log"
)

// topo holds subbasin topology
var topo map[int][]int

func topoOrder() []int {
	eval := make(map[int]bool, len(topo))
	for i := range topo {
		eval[i] = false
	}

	fromto := make(map[int]int, len(topo))
	for to, froms := range topo {
		for _, from := range froms {
			eval[from] = true
			if from < 0 { // headwaters
				continue
			}
			if _, ok := fromto[from]; ok {
				log.Fatalf("swat.topoOrder error: subbasin %d going to %d and %d, must have only one destination\n", from, fromto[from], to)
			}
			fromto[from] = to
		}
	}

	var ord []int
	var recur func(int)
	for r, b := range eval {
		if !b { // roots
			recur = func(b int) {
				ord = append(ord, b)
				for _, f := range topo[b] {
					if f < 0 {
						continue
					}
					recur(f)
				}
			}
			recur(r)
		}
	}

	//reverse order
	for i, j := 0, len(ord)-1; i < j; i, j = i+1, j-1 {
		ord[i], ord[j] = ord[j], ord[i]
	}

	return ord
}

// TopoToOutlet returns an ordered set of subbasin IDs leading to an outlet
func TopoToOutlet(outlet int) []int {
	if _, ok := topo[outlet]; !ok {
		log.Fatalf("swat.topoToOutlet error: no subbasin ID %d in model\n", outlet)
	}
	var ord []int
	var recur func(int)
	recur = func(b int) {
		ord = append(ord, b)
		for _, f := range topo[b] {
			if f < 0 {
				continue
			}
			recur(f)
		}
	}
	recur(outlet)

	// reverse order
	for i, j := 0, len(ord)-1; i < j; i, j = i+1, j-1 {
		ord[i], ord[j] = ord[j], ord[i]
	}

	return ord
}
