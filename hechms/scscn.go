package hechms

func (s *basin) scscn(p float64) float64 {
	s.Pe += p
	if s.Pe > s.ia {
		peia := s.Pe - s.ia
		qi := peia*peia/(peia+s.scn) - s.Q
		s.Q += qi
		return qi
	}
	return 0.
}
