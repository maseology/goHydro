package tem

func (t *TEM) buildDs(bufs map[int][]int) map[int]int {
	ds := make(map[int]int, len(t.TEC))
	for c, tt := range t.TEC {
		ds[c] = func() int { // Single direction steepest decent (D8)
			ii := -1
			zmin := tt.Z
			for _, bc := range bufs[c] {
				if bc < 0 {
					continue
				}
				if bt, ok := t.TEC[bc]; ok {
					if bt.Z < zmin {
						zmin = bt.Z
						ii = bc
					}
				} else {
					panic("buildDs err1")
				}
			}
			return ii
		}()
	}
	return ds
}
