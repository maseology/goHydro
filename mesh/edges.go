package mesh

type Edge struct{ N0, N1, E0, E1 int }

func (sl *Slice) GetEdges() []Edge {

	intersection := func(i1, i2 []int) (o []int) {
		m := make(map[int]bool)
		for _, i := range i1 {
			m[i] = true
		}
		for _, i := range i2 {
			if m[i] {
				o = append(o, i)
			}
		}
		return
	}

	o := make([]Edge, 0, len(sl.Nodes))
	for eid, nids := range sl.Elements {
		if len(nids) != 3 {
			panic("Slice.GetEdges: only supporting triangluar meshes")
		}
		for i := 0; i < 3; i++ {
			newE, c := Edge{N0: nids[i], N1: nids[(i+1)%3], E0: eid, E1: -1}, 0
			for _, sel := range intersection(sl.NExr[newE.N0], sl.NExr[newE.N1]) {
				if sel == eid {
					continue
				}
				newE.E1 = sel
				c++
			}
			if c > 1 {
				panic("Slice.GetEdges: shared element error")
			}
			o = append(o, newE)
		}
	}
	return o
}
