package lia

// SetHeads allows and instantaneous change in ghost node heads
func (d *Domain) SetHeads(m map[int]float64) {
	for f, h := range m {
		d.gns[f].h = h
	}
}
