package tem

import (
	"container/list"
	"encoding/gob"
	"log"
	"os"
)

// TEM topologic elevation model
type TEM struct {
	TEC  map[int]TEC
	USlp map[int][]int
}

// NumCells number of cells that make up the TEM
func (t *TEM) NumCells() int {
	return len(t.TEC)
}

// SaveGob TEM to gob
func (t *TEM) SaveGob(fp string) error {
	f, err := os.Create(fp)
	defer f.Close()
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(t)
	if err != nil {
		return err
	}
	return nil
}

// LoadGob TEM gob
func LoadGob(fp string) (TEM, error) {
	var t TEM
	f, err := os.Open(fp)
	defer f.Close()
	if err != nil {
		return t, err
	}
	enc := gob.NewDecoder(f)
	err = enc.Decode(&t)
	if err != nil {
		return t, err
	}
	return t, nil
}

// Peaks returns list of peak cell IDs (cells that do not receive cascading runon) cascading to cellID cid0. cid0<0 returns all peaks.
func (t *TEM) Peaks(cid0 int) []int {
	p := make([]int, 0)
	if cid0 < 0 {
		for i := range t.TEC {
			if len(t.USlp[i]) == 0 {
				p = append(p, i)
			}
		}
		return p
	}
	c := t.ContributingAreaIDs(cid0)
	for _, i := range c {
		if len(t.USlp[i]) == 0 {
			p = append(p, i)
		}
	}
	return p
}

// UpIDs returns a list of upslope cell IDs
func (t *TEM) UpIDs(cid int) []int {
	return t.USlp[cid]
}

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

	dsa := t.downslopes()
	c, ds, i := make([]int, len(t.TEC)), make(map[int]int, len(dsa)), 0
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

// UpCnt returns a list of upslope cell IDs
func (t *TEM) UpCnt(cid int) int {
	return len(t.climb(cid))
}

// UnitContributingArea computes the (unit) contributing area to a given cell id
func (t *TEM) UnitContributingArea(cid int) int {
	return t.UpCnt(cid)
}

func (t *TEM) climb(cid int) map[int]bool {
	c := make(map[int]bool)
	var climbRecurs func(int)
	climbRecurs = func(cid int) {
		c[cid] = true
		for _, i := range t.USlp[cid] {
			climbRecurs(i)
		}
	}
	climbRecurs(cid)
	return c
}

func (t *TEM) downslopes() map[int]int {
	ds := make(map[int]int, len(t.USlp))
	for to, v := range t.USlp {
		for _, from := range v {
			if _, ok := ds[from]; ok {
				log.Fatalln(" TEM.downslopes() error: expecting a tree graph")
			}
			ds[from] = to
		}
	}
	return ds // from{to}
}

// ContributingCellMap returns a map of upslope TEC count for every TEC in TEM
func (t *TEM) ContributingCellMap() map[int]int {
	mcnt := make(map[int]int, len(t.TEC))
	for c := range t.TEC {
		mcnt[c] = 1
	}
	o, m := t.DownslopeContributingAreaIDs(-1)
	for _, c := range o {
		if v, ok := m[c]; ok { // outlet/farfield cells would not be included here
			mcnt[v] += mcnt[c]
		}
	}
	return mcnt
}

// SubSet returns a subset topologic elevation model from a given outlet cell
func (t *TEM) SubSet(fromid int) TEM {
	uids := t.ContributingAreaIDs(fromid)
	tss, uss := make(map[int]TEC, len(uids)), make(map[int][]int, len(uids))
	for _, c := range uids {
		tss[c] = t.TEC[c]
		uss[c] = t.USlp[c]
	}
	return TEM{TEC: tss, USlp: uss}
}
