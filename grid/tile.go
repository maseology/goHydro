package grid

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"github.com/maseology/mmio"
	"github.com/maseology/wgs84"
	"gonum.org/v1/plot/palette/moreland"
)

const (
	resolution = 256
	dpi        = 96
)

// modified from https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
// X goes from 0 (left edge is 180 °W) to 2^{zoom} − 1 (right edge is 180 °E)
// Y goes from 0 (top edge is 85.0511 °N) to 2^{zoom} − 1 (bottom edge is 85.0511 °S) in a Mercator projection

type Tile struct{ Z, X, Y int }

type TileSet struct {
	Tiles []Tile
	Cxr   [][][]int
}

func (t *Tile) FromLatLong(lat, long float64, z int) {
	n := math.Exp2(float64(z))
	t.X = int(math.Floor((long + 180.) / 360. * n))
	if float64(t.X) >= n {
		t.X = int(n - 1)
	}
	t.Y = int(math.Floor((1. - math.Log(math.Tan(lat*math.Pi/180.)+1./math.Cos(lat*math.Pi/180.))/math.Pi) / 2. * n))
	t.Z = z
}

func (t *Tile) ToLatLong() (lat, long float64) {
	n := math.Pi - 2.*math.Pi*float64(t.Y)/math.Exp2(float64(t.Z))
	lat = 180. / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(t.X)/math.Exp2(float64(t.Z))*360. - 180.
	return lat, long
}

func (t *Tile) ToExtent() (latUL, longUL, latLR, longLR float64) {
	latUL, longUL = t.ToLatLong()
	tLR := Tile{t.Z, t.X + 1, t.Y + 1}
	latLR, longLR = tLR.ToLatLong()
	return latUL, longUL, latLR, longLR
}

// https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Resolution_and_Scale
func (t *Tile) Resolution(lat float64) float64 {
	// 40075.016686 * 1000 / 256 ≈ 6378137.0 * 2 * pi / 256 ≈ 156543.03 // zoom 0: 1 pixel = 156543.03m
	return 156543.03 * math.Cos(lat) / math.Pow(2, float64(t.Z))
}
func (t *Tile) Scale(lat float64) float64 {
	// return (1 : scale)
	return float64(dpi) / .0254 * t.Resolution(lat)
}

////////////////////////////////////////////////
////////////////////////////////////////////////

func (gd *Definition) BuildTileSet(zoomMin, zoomMax, epsg int, outDir string) (tset TileSet) {
	ttt := time.Now()
	gobFP := outDir + gd.Name + ".TileSet.gob"
	if _, ok := mmio.FileExists(gobFP); ok {
		f, _ := os.Open(gobFP)
		enc := gob.NewDecoder(f)
		enc.Decode(&tset)
		f.Close()
	} else {
		mzoom := make(map[int]float64, zoomMax-zoomMin+1)
		for z := zoomMin; z <= zoomMax; z++ {
			mzoom[z] = 156543.03 * math.Cos(44.) / math.Pow(2, float64(z))
			fmt.Printf("   pixel size at zoom %d: %.3fm\n", z, mzoom[z])
		}

		tt := time.Now()
		fmt.Printf(" > converting grid coordinates.. ")
		latlongs := gd.CellCentroidsLatLongs(epsg)
		fmt.Printf("%v\n", time.Since(tt))

		tt = time.Now()
		fmt.Printf(" > collecting tiles to cover %s cells.. ", mmio.Thousands(int64(gd.Nact)))
		m := make(map[Tile][]int)
		for _, c := range gd.Sactives {
			for z := zoomMin; z <= zoomMax; z++ {
				var t Tile
				t.FromLatLong(latlongs[c][0], latlongs[c][1], z)
				m[t] = append(m[t], c)
			}
		}
		fmt.Printf("%v\n", time.Since(tt))

		tt = time.Now()
		fmt.Printf(" > building indices for %s tiles.. ", mmio.Thousands(int64(len(m))))
		tset = TileSet{
			Tiles: make([]Tile, len(m)),
			Cxr:   make([][][]int, len(m)),
		}
		fres := float64(resolution)
		gcell := func(l, h, v float64) int { return int(math.Floor((v - l) / (h - l) * fres)) }
		k := -1
		for t, cs := range m {
			k++
			tset.Tiles[k] = t
			tset.Cxr[k] = make([][]int, resolution*resolution)
			latUL, longUL, latLR, longLR := t.ToExtent()
			if mzoom[t.Z] > gd.Cwidth {
				for _, c := range cs {
					ll := latlongs[c]
					x := gcell(longUL, longLR, ll[1])
					y := resolution - gcell(latLR, latUL, ll[0]) - 1
					ii := x*resolution + y
					tset.Cxr[k][ii] = append(tset.Cxr[k][ii], c)
				}
			} else {
				// for i := 0; i < resolution; i++ {
				// 	lat := latUL - (latUL-latLR)/fres*(float64(i)+.5)
				// 	for j := 0; j < resolution; j++ {
				// 		lng := (longLR-longUL)/fres*(float64(j)+.5) + longUL
				// 		e, n, _ := wgs84.To(wgs84.EPSG().Code(epsg))(lng, lat, 0)
				// 		cid := gd.PointToCellID(e, n)
				// 		ii := j*resolution + i
				// 		tset.Cxr[k][ii] = []int{cid}
				// 	}
				// }
				type result struct{ i, c int }
				ch := make(chan result, 128)
				go func() {
					for i := 0; i < resolution; i++ {
						lat := latUL - (latUL-latLR)/fres*(float64(i)+.5)
						for j := 0; j < resolution; j++ {
							lng := (longLR-longUL)/fres*(float64(j)+.5) + longUL
							e, n, _ := wgs84.To(wgs84.EPSG().Code(epsg))(lng, lat, 0)
							cid := gd.PointToCellID(e, n)
							ii := i*resolution + j
							ch <- result{ii, cid}
						}
					}
					close(ch)
				}()
				for res := range ch {
					tset.Cxr[k][res.i] = []int{res.c}
				}
			}
		}
		fmt.Printf("%v\n", time.Since(tt))

		tt = time.Now()
		fmt.Printf(" > saving to %s.. ", gobFP)
		f, _ := os.Create(gobFP)
		enc := gob.NewEncoder(f)
		enc.Encode(tset)
		f.Close()
		fmt.Printf("%v\n", time.Since(tt))
	}
	fmt.Printf("TOTAL TIME: %v\n", time.Since(ttt))
	return
}

// ////////////////////////////////////////////////
// ////////////////////////////////////////////////
// ////////////////////////////////////////////////
// ////////////////////////////////////////////////

func getdLatLongXR(r *Real, epsg int, outDir string) (m map[int][]float64) {
	gobFP := mmio.GetFileDir(outDir) + "/" + r.GD.Name + ".LatLong.gob"
	if _, ok := mmio.FileExists(gobFP); ok {
		f, _ := os.Open(gobFP)
		enc := gob.NewDecoder(f)
		enc.Decode(&m)
		f.Close()
	} else {
		m = r.GD.CellCentroidsLatLongs(epsg)
		f, _ := os.Create(gobFP)
		enc := gob.NewEncoder(f)
		enc.Encode(m)
		f.Close()
	}
	return
}

func getTileSet(gd *Definition, latlongs map[int][]float64, zoomMin, zoomMax int, tileDir string, saveToGob bool) (m map[Tile][]int) {
	gobFP := mmio.GetFileDir(tileDir) + "/" + gd.Name + ".Tiles.gob"
	if _, ok := mmio.FileExists(gobFP); saveToGob && ok {
		f, _ := os.Open(gobFP)
		enc := gob.NewDecoder(f)
		enc.Decode(&m)
		f.Close()
		zn, zx := 24, 0
		for t := range m {
			mmio.MakeDir(fmt.Sprintf("%s/%d/%d", tileDir, t.Z, t.X))
			if t.Z > zx {
				zx = t.Z
			}
			if t.Z < zn {
				zn = t.Z
			}
		}
		if zn != zoomMin || zx != zoomMax {
			log.Fatalf("Error: %s does not match zoom levels specified. Need to re-create gob\n", gobFP)
		}
	} else {
		m = make(map[Tile][]int)
		for _, c := range gd.Sactives {
			for z := zoomMin; z <= zoomMax; z++ {
				var t Tile
				t.FromLatLong(latlongs[c][0], latlongs[c][1], z)
				mmio.MakeDir(fmt.Sprintf("%s/%d/%d", tileDir, z, t.X))
				m[t] = append(m[t], c)
			}
		}
		// lst := make([]string, 0, len(m)+1)
		// lst = append(lst, "z,x,y,latitude,longitude")
		// for t := range m {
		// 	lat, long := t.ToLatLong()
		// 	lst = append(lst, fmt.Sprintf("%d,%d,%d,%f,%f", t.Z, t.X, t.Y, lat, long))
		// }
		// mmio.LinesToAscii("tiles.csv", lst)
		// os.Exit(22)
		if saveToGob {
			f, _ := os.Create(gobFP)
			enc := gob.NewEncoder(f)
			enc.Encode(m)
			f.Close()
		}
	}
	return
}

// ToTiles take a Real grid and builds a set of raster/image tiles for webmapping
func (r *Real) ToTiles(minVal, maxVal float64, zoomMin, zoomMax, epsg int, tileDir string) {
	fmt.Printf("Building image tiles to directory: '%v'\n   input cell size: %.3fm\n", tileDir, r.GD.Cwidth)
	mzoom := make(map[int]float64, zoomMax-zoomMin+1)
	// minzoom := r.GD.Cwidth
	for z := zoomMin; z <= zoomMax; z++ {
		mzoom[z] = 156543.03 * math.Cos(44.) / math.Pow(2, float64(z))
		fmt.Printf("   pixel size at zoom %d: %.3fm\n", z, mzoom[z])
		// if mzoom[z] < minzoom {
		// 	minzoom = mzoom[z]
		// }
	}
	// mbrngs := BufferRings(int(math.Floor(r.GD.Cwidth / minzoom)))

	ttt := time.Now()
	mmio.MakeDir(tileDir)

	tt := time.Now()
	fmt.Printf(" > converting grid coordinates.. ")
	lls := getdLatLongXR(r, epsg, tileDir)
	fmt.Printf("%v\n", time.Since(tt))

	tt = time.Now()
	fmt.Printf(" > collecting tiles to cover %s cells.. ", mmio.Thousands(int64(len(r.A))))
	mtls := getTileSet(r.GD, lls, zoomMin, zoomMax, tileDir, true)
	fmt.Printf("%v\n", time.Since(tt))

	// create colour map
	cmap := moreland.Kindlmann()
	cmap.SetMin(minVal)
	cmap.SetMax(maxVal)
	minCol, _ := cmap.At(minVal)
	maxCol, _ := cmap.At(maxVal)

	fmt.Printf(" > building %s tiles.. ", mmio.Thousands(int64(len(mtls))))
	tt = time.Now()
	saveImg := func(a [][]float64, fp string) {
		img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{resolution, resolution}})
		for x := 0; x < resolution; x++ {
			for y := 0; y < resolution; y++ {
				if a[x][y] != -9999. {
					// fmt.Println(x, y, rn, rx, a[x][y], fp)
					if col, err := cmap.At(a[x][y]); err != nil {
						if a[x][y] < minVal {
							img.Set(x, y, minCol)
						} else if a[x][y] > maxVal {
							img.Set(x, y, maxCol)
						} else {
							panic(err)
						}
					} else {
						img.Set(x, y, col)
					}
				}
			}
		}
		f, _ := os.Create(fp)
		png.Encode(f, img)
	}

	fres := float64(resolution)
	gcell := func(l, h, v float64) int {
		// fmt.Println(l, h, v, (v-l)/(h-l), math.Floor((v-l)/(h-l)*fres))
		return int(math.Floor((v - l) / (h - l) * fres))
	}
	for t, cs := range mtls {
		n, a := make([][]float64, resolution), make([][]float64, resolution)
		for i := 0; i < resolution; i++ {
			a[i] = make([]float64, resolution)
			n[i] = make([]float64, resolution)
		}

		latUL, longUL, latLR, longLR := t.ToExtent()
		if mzoom[t.Z] > r.GD.Cwidth {
			for _, c := range cs {
				ll := lls[c]
				x := gcell(longUL, longLR, ll[1])
				y := resolution - gcell(latLR, latUL, ll[0]) - 1
				a[x][y] += r.A[c]
				n[x][y]++
			}
			for i := 0; i < resolution; i++ {
				for j := 0; j < resolution; j++ {
					if n[i][j] > 0 {
						a[i][j] /= n[i][j]
					} else {
						a[i][j] = -9999.
					}
				}
			}
		} else {
			for i := 0; i < resolution; i++ {
				lat := latUL - (latUL-latLR)/fres*(float64(i)+.5)
				for j := 0; j < resolution; j++ {
					lng := (longLR-longUL)/fres*(float64(j)+.5) + longUL
					e, n, _ := wgs84.To(wgs84.EPSG().Code(epsg))(lng, lat, 0)
					cid := r.GD.PointToCellID(e, n)
					a[j][i] = r.A[cid]
				}
			}
			// for i := 0; i < resolution; i++ {
			// 	for j := 0; j < resolution; j++ {
			// 		a[i][j] = -9999.
			// 	}
			// }
			// xys := make(map[int][]int, len(cs))
			// for _, c := range cs {
			// 	ll := lls[c]
			// 	x := gcell(longUL, longLR, ll[1])
			// 	y := resolution - gcell(latLR, latUL, ll[0]) - 1
			// 	xys[c] = []int{x, y}
			// 	a[x][y] = r.A[c]
			// }
			// b := int(math.Ceil(r.GD.Cwidth/mzoom[t.Z])) + 1
			// for bb := 1; bb <= b; bb++ {
			// 	for c, xy := range xys {
			// 		for _, mn := range mbrngs[bb] {
			// 			xx, yy := xy[0]+mn[0], xy[1]+mn[1]
			// 			if xx < 0 || yy < 0 || xx >= resolution || yy >= resolution {
			// 				continue
			// 			}
			// 			if a[xx][yy] == -9999 {
			// 				a[xx][yy] = r.A[c]
			// 			}
			// 		}
			// 	}
			// }
		}

		saveImg(a, fmt.Sprintf("%s/%d/%d/%d.png", tileDir, t.Z, t.X, t.Y))
	}
	fmt.Printf("%v\nTOTAL TIME: %v\n", time.Since(tt), time.Since(ttt))
}
