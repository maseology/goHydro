package hechms

func (s *state) scscn(p float64) float64 {
	s.Pe += p
	if s.Pe > s.ia {
		// scn := 1000. / cn - 10. // inches
		scn := 25.4/s.cn - 0.254 // m
		iacn := .2 * scn
		if s.Pe > iacn {
			peia := s.Pe - iacn
			qi := peia*peia/(peia+scn) - s.Q
			s.Q += qi
			return qi
		}
	}
	return 0.
}
