package hechms

type Domain struct {
	SBP       []SubBasinProperties
	Xr, MetXr map[int]int
	Order     []int
	Area      float64
	TSmin     int
}

type SubBasinProperties struct {
	Name string
	Ia, Fimp, CN,
	FlowPathLen, CentFlowPathLen, FlowPathSlope,
	BasinSlope, BasinRelief, BasinRelRatio,
	DrainDensity, Elongation, Area float64
	MetID, Swsid, Dsws int
}

func Build(ws []SubBasinProperties, metxr map[int]int, tsmin int) *Domain {
	Usws, xr := make(map[int][]int), make(map[int]int, len(ws))

	if func() bool {
		for _, s := range ws {
			if s.Swsid != 0 {
				return true
			}
		}
		return false
	}() {
		for i, s := range ws {
			xr[s.Swsid] = i
			Usws[s.Swsid] = []int{}
		}
	} else {
		for i := range ws {
			ws[i].Swsid = i
			xr[i] = i
			Usws[i] = []int{}
		}
	}

	tarea := 0.
	for _, s := range ws {
		tarea += s.Area
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

	if metxr == nil {
		metxr = make(map[int]int, len(ws))
		for i := range ws {
			metxr[i] = 0
		}
	}

	return &Domain{
		SBP:   ws,
		Xr:    xr,
		MetXr: metxr,
		Order: ord,
		Area:  tarea,
		TSmin: tsmin,
	}
}
