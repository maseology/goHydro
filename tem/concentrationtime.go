package tem

func (t *TEM) concentrationTime() map[int]int {
	cnt := make(map[int]int, len(t.TEC))
	var climb func(int, int)
	climb = func(i, c int) {
		cnt[i] = c
		for _, us := range t.USlp[i] {
			if _, ok := cnt[us]; ok {
				continue
			}
			climb(us, c+1)
		}
	}
	for _, rt := range t.Outlets() {
		climb(rt, 1)
	}
	return cnt
}
