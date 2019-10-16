package met

import "log"

func (h *Header) locationSize() int {
	if h.lc == 0 {
		return 4
	} else if h.lc > 0 {
		n := 4
		if h.lc == 1 {
			if h.nloc == 1 {
				n += 4
			} else {
				n += int(h.nloc) * 8
			}
		} else if h.lc == 12 {
			n += int(h.nloc) * 20
		} else {
			log.Panicf("location code %d currently not supported\n", h.lc)
		}
		return n
	} else {
		log.Panicf("location code %d currently not supported\n", h.lc)
	}
	return -1
}
