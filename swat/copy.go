package swat

// Copy (deep) a SubBasin struct
func (b *SubBasin) Copy() *SubBasin {
	sb := *b
	sb.hru = make([]HRU, 0)
	for _, m := range b.hru {
		mnew := m
		mnew.sz = make([]SoilLayer, len(m.sz))
		copy(mnew.sz, m.sz)
		sb.hru = append(sb.hru, mnew)
	}
	return &sb
}
