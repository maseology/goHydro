package gmet

import (
	"encoding/json"
	"fmt"
)

func toJson(ds []DSet) string {
	var iv interface{}
	iv = ds

	jsonMsg, err := json.Marshal(iv)
	if err != nil {
		return fmt.Sprintf("ERROR (GMET.pointToJson): %v", err)
	}
	return string(jsonMsg)
}

func (g *GMET) StaToJson(xsid int) string {
	// library(ggplot2)
	// library(jsonlite)
	// library(dplyr)

	// df <- fromJSON("E:/YCDBnetCDF4.gob.json") %>% na_if(-999)
	// View(df)

	// ggplot(df,aes(as.Date(Date))) +
	//   geom_line(aes(y=Tx)) +
	//   geom_line(aes(y=Tn))

	return toJson(g.Dat[xsid])
}

// func (g *GMET) StasToJson(xsids []int) string {

// 	ds := make([]dset, len(g.Ts))
// 	div0 := func(n, d float64) float64 {
// 		if d != 0. {
// 			return n / d
// 		}
// 		return -999.
// 	}
// 	for i, t := range g.Ts {
// 		sum, cnt := dset{}, dset{}
// 		for _, j := range xsids {
// 			if v := g.Dat[j][i].Tx; v > -100. {
// 				sum.Tx += v
// 				cnt.Tx++
// 			}
// 			if v := g.Dat[j][i].Tn; v > -100. {
// 				sum.Tn += v
// 				cnt.Tn++
// 			}
// 			if v := g.Dat[j][i].Rf; v > -100. {
// 				sum.Rf += v
// 				cnt.Rf++
// 			}
// 			if v := g.Dat[j][i].Sf; v > -100. {
// 				sum.Sf += v
// 				cnt.Sf++
// 			}
// 			if v := g.Dat[j][i].Sm; v > -100. {
// 				sum.Sm += v
// 				cnt.Sm++
// 			}
// 			if v := g.Dat[j][i].Pa; v > -100. {
// 				sum.Pa += v
// 				cnt.Pa++
// 			}
// 		}
// 		ds[i] = dset{
// 			Date: t.Format("2006-01-02"),
// 			Tx:   div0(sum.Tx, cnt.Tx),
// 			Tn:   div0(sum.Tn, cnt.Tn),
// 			Rf:   div0(sum.Rf, cnt.Rf),
// 			Sf:   div0(sum.Sf, cnt.Sf),
// 			Sm:   div0(sum.Sm, cnt.Sm),
// 			Pa:   div0(sum.Pa, cnt.Pa),
// 		}
// 	}

// 	return toJson(ds)
// }
