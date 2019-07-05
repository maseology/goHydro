package swat

import (
	"log"
	"math"
)

const nearzero = 1e-8

// Update state (all in [mm])
func (bsn *SubBasin) Update(vin, p, ep float64) (r, i, a, g, b, vout float64) {
	r, i, a, g = 0., 0., 0., 0.
	sl := bsn.Storage()
	for _, m := range bsn.hru {
		slt := m.storage()
		swprfl := m.drainableStorage() // soil water content of the entire profile excluding the water held in the profile at wilting point [mm] (pg.104)

		rgen := m.cn.Update(p, swprfl, false) // generated runoff
		inf := math.Max(0., p-rgen)           // infiltration
		m.sz[0].sw += inf                     // add infiltration to soil zone
		r += rgen * m.f                       // accumulate subbasin generated runoff
		i += inf * m.f                        // accumulate subbasin infiltration

		s0 := m.storage()
		at := m.evap(ep) // actual et
		s1 := m.storage()
		wbalevap := s0 - (at + s1)
		s0 = s1
		if math.Abs(wbalevap) > nearzero {
			log.Fatalf("HRU wbal error: |wbalevap| = %f\n", wbalevap)
		}
		a += at * m.f // accumulate subbasin evaporation

		gt := m.percolate() // gw recharge
		s1 = m.storage()
		wbalperc := s0 - (gt + s1)
		s0 = s1
		if math.Abs(wbalperc) > nearzero {
			log.Fatalf("HRU wbal error: |wbalperc| = %f\n", wbalperc)
		}
		g += gt * m.f // accumulate subbasin gw recharge

		wbal := p + slt - s0 - at - gt - rgen
		if math.Abs(wbal) > nearzero {
			log.Fatalf("HRU wbal error: |wbal| = %f\n", wbal)
		}
	}
	s0 := bsn.Storage()
	rbsn := bsn.surfRunoffLag(r) // basin lagged runoff
	s1 := bsn.Storage()
	wbalro := s0 + r - (rbsn + s1)
	s0 = s1
	if math.Abs(wbalro) > nearzero {
		log.Fatalf("SubBasin runoff wbal error: |wbalro| = %f\n", wbalro)
	}

	b = bsn.baseflow(g) // baseflow
	s1 = bsn.Storage()
	wbalbf := s0 + g - (b + s1)
	s0 = s1
	if math.Abs(wbalbf) > nearzero {
		log.Fatalf("SubBasin baseflow wbal error: |wbalbf| = %f\n", wbalbf)
	}

	wbal := p + sl - s0 - a - rbsn - b // subbasin wb
	if math.Abs(wbal) > nearzero {
		log.Fatalf("SubBasin wbal error: |wbal| = %f\n", wbal)
	}

	vout = bsn.chn.Route(vin + (rbsn+b)*bsn.Ca*1000.) // SubBasin daily average outflow [m³/d]
	s1 = bsn.Storage()
	vinmm := vin / bsn.Ca / 1000.
	voutmm := vout / bsn.Ca / 1000.
	wbalrte := s0 + vinmm + rbsn + b - (voutmm + s1)
	s0 = s1
	if math.Abs(wbalrte) > nearzero {
		log.Fatalf("SubBasin rte wbal error: |wbalrte| = %f\n", wbalrte)
	}

	wbal2 := p + vinmm + sl - s0 - a - voutmm
	if math.Abs(wbal2) > nearzero {
		log.Fatalf("SubBasin wbal2 (post routing) error: |wbal2| = %f\n", wbal2)
	}
	return
}
