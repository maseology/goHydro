package hechms

type Domain struct {
	SB      []SubBasin
	Xr, Mxr map[int]int
	Order   []int
}

type SubBasin struct {
	Name string
	Ia, Fimp, CN,
	FlowPathLen, FlowPathSlope,
	BasinSlope, BasinRelief, BasinRelRatio,
	DrainDensity, Elongation, Area float64
	MetID, Swsid, Dsws int
}

func Build(ws []SubBasin, mxr map[int]int) *Domain {
	Usws, xr := make(map[int][]int), make(map[int]int, len(ws))
	for i, s := range ws {
		xr[s.Swsid] = i
		Usws[s.Swsid] = []int{}
	}
	for _, s := range ws {
		if _, ok := xr[s.Dsws]; ok {
			Usws[s.Dsws] = append(Usws[s.Dsws], s.Swsid)
		}
	}
	eval, q := make(map[int]int), []int{}
	for _, s := range ws {
		if len(Usws[s.Swsid]) == 0 {
			if _, ok := xr[s.Swsid]; !ok {
				panic("hechms.Build err1")
			}
			q = append(q, xr[s.Swsid])
		}
	}

	ord := make([]int, 0, len(ws))
	for len(q) > 0 {
		qq := q[0]
		w := ws[qq]
		ord = append(ord, qq)
		eval[w.Swsid]++
		q = q[1:]
		if us, ok := Usws[w.Dsws]; ok {
			if func() bool {
				for _, u := range us {
					if _, ok := eval[u]; !ok {
						return false
					}
				}
				return true
			}() {
				q = append(q, xr[w.Dsws])
			}
		}
	}

	return &Domain{
		SB:    ws,
		Xr:    xr,
		Mxr:   mxr,
		Order: ord,
	}
}
