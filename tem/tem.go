package tem

// TEM topologic elevation model
type TEM struct {
	TECs map[int]TEC
	us   map[int][]int
	c    map[int]bool
}

// NumCells number of cells that make up the TEM
func (t *TEM) NumCells() int {
	return len(t.TECs)
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
