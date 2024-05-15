package tem

// // BuildUpslopes re-builds upslope mapping
// func (t *TEM) BuildUpslopes(bufs map[int][]int) {
// 	ds := t.buildDsFromNeighbours(bufs)
// 	t.buildUpslopes(ds)
// }

func (t *TEM) buildUpslopes(ds map[int]int) {
	tu := make(map[int][]int)
	for i := range t.TEC {
		if di, ok := ds[i]; ok && di >= 0 {
			if di > -1 {
				tu[di] = append(tu[di], i)
			}
		}
	}
	t.USlp = make(map[int][]int, len(tu))
	for k, v := range tu {
		t.USlp[k] = v
	}
}
