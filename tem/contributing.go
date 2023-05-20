package tem

import "container/list"

// ContributingAreaIDs returns a list of upslope cell IDs that make up the contributing area to cid0
func (t *TEM) ContributingAreaIDs(cid0 int) []int {
	cs := t.climb(cid0)
	a, i := make([]int, len(cs)), 0
	for c := range cs {
		a[i] = c
		i++
	}
	return a
}

// DownslopeContributingAreaIDs returns a list of upslope cell IDs that make up the contributing area to cid0,
// yet ordered in the downslope (topologically-safe) direction. cid0 < 0 returns entire TEM
func (t *TEM) DownslopeContributingAreaIDs(cid0 int) ([]int, map[int]int) {
	queue := list.New()
	eval := make(map[int]bool, len(t.TEC))
	proceed := func(cid int) bool {
		if v, ok := t.USlp[cid]; ok {
			for _, u := range v { // returns true if all upslope cells have been evaluated
				if !eval[u] {
					return false
				}
			}
		}
		return true
	}

	dsa := t.downslopes() // from{to}
	c, ds, i := make([]int, len(t.TEC)), make(map[int]int), 0
	for _, k := range t.Peaks(cid0) {
		queue.PushBack(k) // initial enqueue
	}

	for queue.Len() > 0 {
		e := queue.Front() // first element
		c[i] = e.Value.(int)
		eval[c[i]] = true
		if v, ok := dsa[c[i]]; ok {
			ds[c[i]] = v
			if v != cid0 && proceed(v) {
				queue.PushBack(v) // enqueue
			}
		}
		queue.Remove(e) // dequeue
		i++
	}
	if cid0 < 0 {
		return c, ds
	}
	c[i] = cid0
	ds[cid0] = -1
	u := make([]int, i+1)
	copy(u, c)

	// cktopo := make(map[int]bool, len(u))
	// for _, i := range u {
	// 	if _, ok := cktopo[i]; ok {
	// 		log.Fatalf(" TEM.DownslopeContributingAreaIDs error: cell %d occured more than once, possible cycle", i)
	// 	}
	// 	if _, ok := ds[i]; !ok {
	// 		log.Fatalf(" TEM.DownslopeContributingAreaIDs error: cell %d not given dowslope id", i)
	// 	}
	// 	if _, ok := cktopo[ds[i]]; ok {
	// 		log.Fatalf(" TEM.DownslopeContributingAreaIDs error: cell %d out of topological order", i)
	// 	}
	// 	cktopo[i] = true
	// }

	return u, ds
}

// UnitContributingArea computes the (unit) contributing area (count) to a given cell id
func (t *TEM) UnitContributingArea(cid int) int {
	return t.UpCnt(cid)
}

// ContributingCellMap returns a map of upslope TEC count for every TEC in TEM cascading to cellID cid0.
func (t *TEM) ContributingCellMap(cid0 int) map[int]int {
	o, m := t.DownslopeContributingAreaIDs(cid0)
	mcnt := make(map[int]int, len(o))
	for _, c := range o {
		mcnt[c] = 1
	}
	for _, c := range o {
		if v, ok := m[c]; ok { // outlet/farfield cells would not be included here
			if v > -1 {
				mcnt[v] += mcnt[c]
			}
		}
	}
	return mcnt
}

// ContributingCellCounts returns a map of upslope TEC count for every TEC in TEM
func (t *TEM) ContributingCellCounts() map[int]int {
	mcnt := make(map[int]int, len(t.TEC))
	var climbRecurs func(int) int
	for _, rt := range t.Outlets() {
		climbRecurs = func(cid int) int {
			mcnt[cid] = 1
			if us, ok := t.USlp[cid]; ok {
				for _, p := range us {
					if _, ok := mcnt[p]; ok {
						continue
					}
					mcnt[cid] += climbRecurs(p)
				}
			}
			return mcnt[cid]
		}
		climbRecurs(rt)
	}
	return mcnt
}
