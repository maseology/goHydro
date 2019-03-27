package tem

import (
	"container/list"
	"log"
)

// TEM topologic elevation model
type TEM struct {
	TEC map[int]TEC
	us  map[int][]int
	c   map[int]bool
}

// NumCells number of cells that make up the TEM
func (t *TEM) NumCells() int {
	return len(t.TEC)
}

// Peaks returns list of peak cell IDs (cells that do not receive cascading runon) cascading to cellID cid0. cid0<0 returns all peaks.
func (t *TEM) Peaks(cid0 int) []int {
	p := make([]int, 0)
	if cid0 < 0 {
		for i, v := range t.us {
			if len(v) == 0 {
				p = append(p, i)
			}
		}
		return p
	}
	c := t.ContributingAreaIDs(cid0)
	for _, i := range c {
		if len(t.us[i]) == 0 {
			p = append(p, i)
		}
	}
	return p
}

// UpIDs returns a list of upslope cell IDs
func (t *TEM) UpIDs(cid int) []int {
	return t.us[cid]
}

// ContributingAreaIDs returns a list of upslope cell IDs that make up the contributing area to cid0
func (t *TEM) ContributingAreaIDs(cid0 int) []int {
	t.c = make(map[int]bool)
	t.climb(cid0)
	a, i := make([]int, len(t.c)), 0
	for c := range t.c {
		a[i] = c
		i++
	}
	return a
}

// DownslopeContributingAreaIDs returns a list of upslope cell IDs that make up the contributing area to cid0, yet ordered in the downslope direction
func (t *TEM) DownslopeContributingAreaIDs(cid0 int) ([]int, map[int]int) {
	queue := list.New()
	eval := make(map[int]bool, len(t.us))
	proceed := func(cid int) bool {
		for _, u := range t.us[cid] { // returns true if all upslope cells have been evaluated
			if !eval[u] {
				return false
			}
		}
		return true
	}

	dsa := t.downslopes()
	c, ds, i := make([]int, len(t.us)), make(map[int]int, len(t.us)), 0
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
	c[len(c)-1] = cid0
	ds[cid0] = -1
	return c, ds
}

// UpCnt returns a list of upslope cell IDs
func (t *TEM) UpCnt(cid int) int {
	t.c = make(map[int]bool)
	t.climb(cid)
	return len(t.c)
}

// UnitContributingArea computes the (unit) contributing area from a given cell id
func (t *TEM) UnitContributingArea(cid int) float64 {
	return float64(t.UpCnt(cid))
}

func (t *TEM) climb(cid int) {
	t.c[cid] = true
	for _, i := range t.us[cid] {
		t.climb(i)
	}
}

func (t *TEM) downslopes() map[int]int {
	ds := make(map[int]int, len(t.us))
	for to, v := range t.us {
		for _, from := range v {
			if _, ok := ds[from]; ok {
				log.Fatalln(" TEM.downslopes() error: expecting a tree graph")
			}
			ds[from] = to
		}
	}
	return ds // from{to}
}

// SubSet returns a subset topologic elevation model from a given outlet cell
func (t *TEM) SubSet(fromid int) TEM {
	uids := t.ContributingAreaIDs(fromid)
	tss, uss := make(map[int]TEC, len(uids)), make(map[int][]int, len(uids))
	for _, c := range uids {
		tss[c] = t.TEC[c]
		uss[c] = t.us[c]
	}
	return TEM{TEC: tss, us: uss}
}
