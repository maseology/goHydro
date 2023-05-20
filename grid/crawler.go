package grid

type Crawler struct {
	adj map[int][]int
}

var bufrc = [][]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}}

func (gd *Definition) ToCrawler() *Crawler {
	a := make(map[int][]int, gd.Nact)
	for _, c := range gd.Sactives {
		a[c] = make([]int, 0, 8)

		r, c := gd.RowCol(c)
		for _, drc := range bufrc {
			sc := gd.CellID(r+drc[0], c+drc[1])
			if gd.IsActive(sc) {
				a[c] = append(a[c], sc)
			}
		}
	}

	return &Crawler{a}
}
