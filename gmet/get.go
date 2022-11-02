package gmet

func (g *GMET) GetData(stationid int, varname string) []float64 {
	f := make([]float64, len(g.Ts))
	k := func() int {
		for i, s := range g.Snams {
			if varname == s {
				return i
			}
		}
		return -1
	}()

	for j := range g.Ts {
		f[j] = g.Dat[stationid][j].Dat[k]
	}
	return f
}

func (g *GMET) GetAllData(varname string) [][]float64 {
	ff := make([][]float64, g.Nsta)
	k := func() int {
		for i, s := range g.Snams {
			if varname == s {
				return i
			}
		}
		return -1
	}()
	for i := range g.Sids {
		f := make([]float64, g.Nts)
		for j := range g.Ts {
			f[j] = g.Dat[i][j].Dat[k]
		}
		ff[i] = f
	}
	return ff
}
