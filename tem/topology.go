package tem

import "log"

// Outlets returns cells that flow to farfield
func (t *TEM) Outlets() []int {
	var o []int
	// for i, t := range t.TEC {
	// 	if t.Ds < 0 {
	// 		o = append(o, i)
	// 	}
	// }
	nds := make(map[int]int, len(t.TEC))
	for i := range t.TEC {
		if us, ok := t.USlp[i]; ok {
			for _, u := range us {
				nds[u]++
			}
		}
	}
	for i := range t.TEC {
		if _, ok := nds[i]; !ok {
			o = append(o, i)
		}
	}
	return o
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

// // UpIDs returns a list of upslope cell IDs
// func (t *TEM) UpIDs(cid int) []int {
// 	return t.USlp[cid]
// }

// UpCnt returns a list of upslope cell IDs
func (t *TEM) UpCnt(cid int) int {
	return len(t.climb(cid))
}

func (t *TEM) climb(cid int) map[int]bool {
	c := make(map[int]bool)
	var climbRecurs func(int)
	climbRecurs = func(cid int) {
		if _, ok := c[cid]; ok {
			return
		}
		c[cid] = true
		for _, i := range t.USlp[cid] {
			climbRecurs(i)
		}
	}
	climbRecurs(cid)
	return c
}

// SubSet returns a subset topologic elevation model from a given outlet cell
// fromid < 0 return a pointer to the referenced TEM
func (t *TEM) SubSet(fromid int) (*TEM, []int) {
	uids := t.ContributingAreaIDs(fromid)
	if fromid < 0 {
		return t, uids
	}
	tss, uss := make(map[int]TEC, len(uids)), make(map[int][]int, len(uids))
	for _, c := range uids {
		tss[c] = t.TEC[c]
		uss[c] = t.USlp[c]
	}
	return &TEM{TEC: tss, USlp: uss}, uids
}

func (t *TEM) Downslopes() map[int]int {
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
