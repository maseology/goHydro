package grid

// SurroundingCells returns the relative row,col given in units of cell width
func SurroundingCells(unitRadius int) [][]int {
	o := [][]int{}
	trsh := unitRadius * unitRadius
	for m := -unitRadius; m <= unitRadius; m++ {
		for n := -unitRadius; n <= unitRadius; n++ {
			if m*m+n*n <= trsh {
				o = append(o, []int{m, n})
			}
		}
	}
	return o
}

func SurroundingRing(unitRadius int) [][]int {
	if unitRadius < 1 {
		panic("grid.SurroundingRing ERROR: unitRadius must be > 0")
	}
	type mn struct{ m, n int }
	oinner := SurroundingCells(unitRadius - 1)
	minner := make(map[mn]bool, len(oinner))
	for _, a := range oinner {
		minner[mn{a[0], a[1]}] = true
	}
	oouter := SurroundingCells(unitRadius)
	o := make([][]int, 0, len(oouter)-len(oinner))
	for _, a := range oouter {
		if _, ok := minner[mn{a[0], a[1]}]; !ok {
			o = append(o, a)
		}
	}
	return o
}

func BufferRings(nrings int) map[int][][]int {
	o := make(map[int][][]int, nrings)
	for k := 1; k <= nrings; k++ {
		o[k] = SurroundingRing(k)
	}
	return o
}

func BufferRingsSquare(nrings int) map[int][][]int {
	o := make(map[int][][]int, nrings)
	for k := 1; k <= nrings; k++ {
		o[k] = make([][]int, 0, 4*2*k)
		for m := 1 - k; m <= k-1; m++ {
			o[k] = append(o[k], []int{m, -k})
			o[k] = append(o[k], []int{m, k})
			o[k] = append(o[k], []int{k, m})
			o[k] = append(o[k], []int{-k, m})
		}
		// corners
		o[k] = append(o[k], []int{-k, -k})
		o[k] = append(o[k], []int{k, -k})
		o[k] = append(o[k], []int{k, k})
		o[k] = append(o[k], []int{-k, k})
	}
	return o
}
