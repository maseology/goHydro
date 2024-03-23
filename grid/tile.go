package grid

import (
	"encoding/gob"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"time"

	"github.com/maseology/mmaths/slice"
	"github.com/maseology/mmio"
	"github.com/maseology/wgs84"
	"gonum.org/v1/plot/palette/moreland"
)

const (
	resolution = 256
	dpi        = 96
	maxLat     = 45.4 // the approximate most-northerly latitude needed to estimate minimum pixel size
)

// modified from https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
// X goes from 0 (left edge is 180 °W) to 2^{zoom} − 1 (right edge is 180 °E)
// Y goes from 0 (top edge is 85.0511 °N) to 2^{zoom} − 1 (bottom edge is 85.0511 °S) in a Mercator projection

type Tile struct{ Z, X, Y int }

type TileSet struct {
	Tiles   []Tile
	Cids    [][]int           // cell ids covering tiles
	Clnglat map[int][]float64 // lat-long projected cell centroid locations
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
			mzoom[z] = 156543.03 * math.Cos(maxLat) / math.Pow(2, float64(z)) // https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Resolution_and_Scale
			// fmt.Printf("   pixel size at zoom %d: %.3fm\n", z, mzoom[z])
		}

		tt := time.Now()
		fmt.Printf(" > re-projecting grid-node coordinates.. ")
		gdv := gd.ToVertex()
		clonglats := wgs84.ReprojectMap(gd.CellCentroids(), epsg, 4326)
		vlonglats := wgs84.ReprojectMap(gdv.Nodecoord, epsg, 4326)
		fmt.Printf("%v\n", time.Since(tt))

		tt = time.Now()
		fmt.Printf(" > collecting tiles to cover %s cells.. ", mmio.Thousands(int64(gd.Nact)))
		m := make(map[Tile][]int)
		for vid, crd := range vlonglats {
			for z := zoomMin; z <= zoomMax; z++ {
				var t Tile
				t.FromLatLong(crd[1], crd[0], z)
				m[t] = append(m[t], gdv.Nodecells[vid]...)
			}
		}
		fmt.Printf("%v\n", time.Since(tt))

		tt = time.Now()
		fmt.Printf(" > building indices for %s tiles.. ", mmio.Thousands(int64(len(m))))
		tset = TileSet{
			Tiles:   make([]Tile, len(m)),
			Cids:    make([][]int, len(m)),
			Clnglat: clonglats,
		}
		k := -1
		for t, cs := range m {
			k++
			tset.Tiles[k] = t
			tset.Cids[k] = func() []int {
				o, b := slice.Distinct(cs), false
				for _, c := range o {
					if c < 0 {
						b = true
						break
					}
				}
				if b {
					oo := make([]int, 0, len(o)-1)
					for _, v := range o {
						if v < 0 {
							continue
						}
						oo = append(oo, v)
					}
					return oo
				}
				return o
			}()
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

// ToTiles take a Real grid and builds a set of raster/image tiles for webmapping
func (r *Real) ToTiles(minVal, maxVal float64, zoomMin, zoomMax, epsg int, tileDir string) {
	fmt.Printf("Building image tiles to directory: %s | input cell size: %.3fm\n", tileDir, r.GD.Cwidth)

	ttt := time.Now()
	mmio.MakeDir(tileDir)

	tset := r.GD.BuildTileSet(zoomMin, zoomMax, epsg, mmio.GetFileDir(tileDir)+"/")
	mzoom, minzoom := make(map[int]float64, zoomMax-zoomMin+1), r.GD.Cwidth
	fmt.Printf("  pixel sizes at latidude %.3f:\n", maxLat)
	for z := zoomMin; z <= zoomMax; z++ {
		mzoom[z] = 156543.03 * math.Cos(maxLat) / math.Pow(2, float64(z)) // https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Resolution_and_Scale
		fmt.Printf("   pixel size at zoom %d: %.3fm\n", z, mzoom[z])
		if mzoom[z] < minzoom {
			minzoom = mzoom[z]
		}
	}

	// create colour map
	cmap := moreland.Kindlmann()
	cmap.SetMin(minVal)
	cmap.SetMax(maxVal)
	minCol, _ := cmap.At(minVal)
	maxCol, _ := cmap.At(maxVal)

	fmt.Printf(" > building %s tiles.. ", mmio.Thousands(int64(len(tset.Tiles))))
	tt := time.Now()
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

	fres, cellRad := float64(resolution), r.GD.Cwidth // math.Sqrt(2*r.GD.Cwidth*r.GD.Cwidth)
	gcell := func(l, h, v float64) int { return int(math.Floor((v - l) / (h - l) * fres)) }
	mbrngs := BufferRingsSquare(int(math.Ceil(cellRad / minzoom)))
	for k, t := range tset.Tiles {
		mmio.MakeDir(fmt.Sprintf("%s/%d/%d", tileDir, t.Z, t.X))
		n, a := make([][]float64, resolution), make([][]float64, resolution)
		for i := 0; i < resolution; i++ {
			a[i] = make([]float64, resolution)
			n[i] = make([]float64, resolution)
		}

		latUL, longUL, latLR, longLR := t.ToExtent()
		if mzoom[t.Z] > r.GD.Cwidth { // aggregate
			for _, c := range tset.Cids[k] {
				ll := tset.Clnglat[c]
				x := gcell(longUL, longLR, ll[0])
				y := resolution - gcell(latLR, latUL, ll[1]) - 1
				if x >= 0 && y >= 0 && x < resolution && y < resolution {
					a[x][y] += r.A[c]
					n[x][y]++
				}
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
				for j := 0; j < resolution; j++ {
					a[i][j] = -9999.
				}
			}
			xys := make(map[int][]int, len(tset.Cids[k]))
			for _, c := range tset.Cids[k] {
				ll := tset.Clnglat[c]
				x := gcell(longUL, longLR, ll[0])
				y := resolution - gcell(latLR, latUL, ll[1]) - 1
				xys[c] = []int{x, y}
				if x >= 0 && y >= 0 && x < resolution && y < resolution {
					a[x][y] = r.A[c]
				}
			}
			b := int(math.Ceil(cellRad / mzoom[t.Z]))
			for bb := 1; bb <= b; bb++ {
				for c, xy := range xys {
					for _, mn := range mbrngs[bb] {
						xx, yy := xy[0]+mn[0], xy[1]+mn[1]
						if xx < 0 || yy < 0 || xx >= resolution || yy >= resolution {
							continue
						}
						if a[xx][yy] == -9999 {
							a[xx][yy] = r.A[c]
						}
					}
				}
			}
		}

		saveImg(a, fmt.Sprintf("%s/%d/%d/%d.png", tileDir, t.Z, t.X, t.Y))
	}
	fmt.Printf("%v\nTOTAL TIME: %v\n", time.Since(tt), time.Since(ttt))
}

// ToTiles take a categorical grid and builds a set of raster/image tiles for webmapping
func (g *Indx) ToTiles(cmap map[int]color.RGBA, zoomMin, zoomMax, epsg int, tileDir string) {
	fmt.Printf("Building image tiles to directory: %s | input cell size: %.3fm\n", tileDir, g.GD.Cwidth)

	ttt := time.Now()
	mmio.MakeDir(tileDir)

	tset := g.GD.BuildTileSet(zoomMin, zoomMax, epsg, mmio.GetFileDir(tileDir)+"/")
	mzoom, minzoom := make(map[int]float64, zoomMax-zoomMin+1), g.GD.Cwidth
	fmt.Printf("  pixel sizes at latidude %.3f:\n", maxLat)
	for z := zoomMin; z <= zoomMax; z++ {
		mzoom[z] = 156543.03 * math.Cos(maxLat) / math.Pow(2, float64(z)) // https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames#Resolution_and_Scale
		fmt.Printf("   pixel size at zoom %d: %.3fm\n", z, mzoom[z])
		if mzoom[z] < minzoom {
			minzoom = mzoom[z]
		}
	}

	fmt.Printf(" > building %s tiles.. ", mmio.Thousands(int64(len(tset.Tiles))))
	tt := time.Now()
	saveImg := func(a [][]int, fp string) {
		img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{resolution, resolution}})
		for x := 0; x < resolution; x++ {
			for y := 0; y < resolution; y++ {
				if a[x][y] != -9999. {
					if col, ok := cmap[a[x][y]]; ok {
						img.Set(x, y, col)
					} else {
						panic("indx.saveImg set ERROR")
					}
				}
			}
		}
		f, _ := os.Create(fp)
		png.Encode(f, img)
	}

	fres, cellRad := float64(resolution), g.GD.Cwidth // math.Sqrt(2*g.GD.Cwidth*g.GD.Cwidth)
	gcell := func(l, h, v float64) int { return int(math.Floor((v - l) / (h - l) * fres)) }
	mbrngs := BufferRingsSquare(int(math.Ceil(cellRad / minzoom)))
	for k, t := range tset.Tiles {
		mmio.MakeDir(fmt.Sprintf("%s/%d/%d", tileDir, t.Z, t.X))
		m, a := make([][]map[int]int, resolution), make([][]int, resolution)
		for i := 0; i < resolution; i++ {
			m[i] = make([]map[int]int, resolution)
			a[i] = make([]int, resolution)
		}

		latUL, longUL, latLR, longLR := t.ToExtent()
		if mzoom[t.Z] > g.GD.Cwidth { // aggregate
			for _, c := range tset.Cids[k] {
				ll := tset.Clnglat[c]
				x := gcell(longUL, longLR, ll[0])
				y := resolution - gcell(latLR, latUL, ll[1]) - 1
				if x >= 0 && y >= 0 && x < resolution && y < resolution {
					m[x][y][g.A[c]]++
				}
			}
			for i := 0; i < resolution; i++ {
				for j := 0; j < resolution; j++ {
					if len(m[i][j]) > 0 {
						k1, n1 := -1, -1
						for k, n := range m[i][j] {
							if n > n1 {
								n1 = n
								k1 = k
							}
						}
						a[i][j] = k1
					} else {
						a[i][j] = -9999.
					}
				}
			}
		} else {
			for i := 0; i < resolution; i++ {
				for j := 0; j < resolution; j++ {
					a[i][j] = -9999
				}
			}
			xys := make(map[int][]int, len(tset.Cids[k]))
			for _, c := range tset.Cids[k] {
				ll := tset.Clnglat[c]
				x := gcell(longUL, longLR, ll[0])
				y := resolution - gcell(latLR, latUL, ll[1]) - 1
				xys[c] = []int{x, y}
				if x >= 0 && y >= 0 && x < resolution && y < resolution {
					a[x][y] = g.A[c]
				}
			}
			b := int(math.Ceil(cellRad / mzoom[t.Z]))
			for bb := 1; bb <= b; bb++ {
				for c, xy := range xys {
					for _, mn := range mbrngs[bb] {
						xx, yy := xy[0]+mn[0], xy[1]+mn[1]
						if xx < 0 || yy < 0 || xx >= resolution || yy >= resolution {
							continue
						}
						if a[xx][yy] == -9999 {
							a[xx][yy] = g.A[c]
						}
					}
				}
			}
		}

		saveImg(a, fmt.Sprintf("%s/%d/%d/%d.png", tileDir, t.Z, t.X, t.Y))
	}
	fmt.Printf("%v\nTOTAL TIME: %v\n", time.Since(tt), time.Since(ttt))
}
