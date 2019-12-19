package lia // VERSION 2

import (
	"math"
)

// type kv struct {
// 	k int
// 	v flux
// }

func (d *Domain) updateFluxes() {
	q := make([]float64, len(d.WKR))
	for k := range d.WKR {
		q[k] = d.WKR[k].getFlux(d.Theta, d.dt, d.dx)
	}
	for k := range d.WKR {
		d.WKR[k].q.q = q[k]
	}
}

// func (d *Domain) updateFluxes() {
// 	var wg sync.WaitGroup
// 	ch := make(chan kv, len(d.WKR))
// 	for k := range d.WKR {
// 		wg.Add(1)
// 		go func(k int) {
// 			defer wg.Done()
// 			// if k == 0 {
// 			// 	fmt.Printf(" b%10.6f  --", d.WKR[k].nb.h)
// 			// }
// 			ch <- kv{k, d.WKR[k].getFlux(d.Theta, d.dt, d.dx)}
// 		}(k)
// 	}
// 	wg.Wait()
// 	close(ch)
// 	for i := 0; i < len(d.WKR); i++ {
// 		kv := <-ch
// 		*d.WKR[kv.k].q = kv.v
// 		// if kv.k == 2001 {
// 		// 	for i, v := range cf {
// 		// 		fmt.Printf(" %d %d:%f ", i, v, d.qs[v])
// 		// 	}
// 		// 	fmt.Printf("   %f\n", n.h)
// 		// 	time.Sleep(500 * time.Millisecond)
// 		// }
// 	}
// }

func (d *Domain) updateHeads() float64 {
	d1 := d.dt / d.dx // math.Pow(d.dt/d.dx,2.) // error in equation 11, see eq. 20 in de Almeda etal 2012
	dhx, adhx := 0., 0.
	for i := 0; i < d.GF.GD.Na; i++ {
		cf := d.GF.CellFace[i]
		dh := d1 * (d.qs[cf[2]].q - d.qs[cf[0]].q + d.qs[cf[3]].q - d.qs[cf[1]].q) // eq.11
		if d.r != nil {
			print("updateHeads TODO: nodal source/sinks (rain/infiltration")
			// if v, ok := d.r[k]; ok {
			// 	dh += v
			// }
		}
		d.ns[i].h += dh
		adh := math.Abs(dh)
		if adh > adhx {
			adhx = adh
			dhx = dh
		}
	}
	return dhx
}

// func (d *Domain) updateHeads() float64 {
// 	d1 := d.dt / d.dx // math.Pow(d.dt/d.dx,2.) // error in equation 11, see eq. 20 in de Almeda etal 2012
// 	ch := make(chan float64, d.GF.GD.Na)
// 	for i := 0; i < d.GF.GD.Na; i++ {
// 		go func(i int, n *node) {
// 			//   1
// 			// 2   0
// 			//   3
// 			cf := d.GF.CellFace[i]
// 			dh := d1 * float64(d.qs[cf[2]]-d.qs[cf[0]]+d.qs[cf[3]]-d.qs[cf[1]]) // eq.11
// 			if d.r != nil {
// 				print("updateHeads TODO: nodal source/sinks (rain/infiltration")
// 				// if v, ok := d.r[k]; ok {
// 				// 	dh += v
// 				// }
// 			}
// 			// ch <- kv{i, dh}
// 			n.h += dh
// 			// if i == 0 {
// 			// 	for _, v := range cf {
// 			// 		fmt.Printf(" [%d]:%f ", v, d.qs[v])
// 			// 	}
// 			// 	fmt.Printf("   %10.6f  --", n.h)
// 			// 	time.Sleep(500 * time.Millisecond)
// 			// }
// 			ch <- dh
// 		}(i, &d.ns[i])
// 		// if i == 0 {
// 		// 	fmt.Printf("   %10.6f [%10.6f] {%p %p} --", d.ns[i].h, d.WKR[i].nb.h, &d.ns[i], d.WKR[i].nb)
// 		// }
// 	}

// 	dhx, adhx := 0., 0.
// 	for i := 0; i < d.GF.GD.Na; i++ {
// 		dh := <-ch
// 		adh := math.Abs(dh)
// 		if adh > adhx {
// 			adhx = adh
// 			dhx = dh
// 		}
// 	}
// 	close(ch)

// 	return dhx
// }
