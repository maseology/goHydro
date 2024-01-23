package grid

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"time"

	"github.com/maseology/mmio"
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

// ToTiles take a Real grid and builds a set of raster/image tiles for webmapping
func (r *Real) ToTiles(zoomMin, zoomMax, epsg int, outDir string) {
	fmt.Printf("Building image tiles to directory: '%v'\n    input cell size: %f\n", outDir, r.GD.Cwidth)
	mzoom := make(map[int]float64, zoomMax-zoomMin+1)
	for z := zoomMin; z <= zoomMax; z++ {
		mzoom[z] = 156543.03 * math.Cos(44.) / math.Pow(2, float64(z))
		fmt.Printf("   pixel size at zoom %d: %.3fm\n", z, mzoom[z])
	}

	fmt.Printf(" > converting grid coordinates.. ")
	tt := time.Now()
	mmio.MakeDir(outDir)
	mtls := make(map[Tile][]int)
	lls := r.GD.CellCentroidsLatLong(epsg, true)
	fmt.Printf("%v\n", time.Since(tt))

	fmt.Printf(" > collecting %s cell values.. ", mmio.Thousands(int64(len(r.A))))
	tt = time.Now()
	rn, rx := math.MaxFloat64, -math.MaxFloat64
	for c, rac := range r.A {
		if rac < rn {
			rn = rac
		}
		if rac > rx {
			rx = rac
		}
		for z := zoomMin; z <= zoomMax; z++ {
			var t Tile
			t.FromLatLong(lls[c][0], lls[c][1], z)
			mmio.MakeDir(fmt.Sprintf("%s/%d/%d", outDir, z, t.X))
			mtls[t] = append(mtls[t], c)
		}
	}
	fmt.Printf("range: [%.3f,%.3f] (%v)\n", rn, rx, time.Since(tt))

	// lst := make([]string, 0, len(mtls)+1)
	// lst = append(lst, "z,x,y,latitude,longitude")
	// for t := range mtls {
	// 	lat, long := t.ToLatLong()
	// 	lst = append(lst, fmt.Sprintf("%d,%d,%d,%f,%f", t.Z, t.X, t.Y, lat, long))
	// }
	// mmio.LinesToAscii("tiles.csv", lst)
	// os.Exit(22)

	// create colour map
	cmap := moreland.Kindlmann()
	cmap.SetMin(rn)
	cmap.SetMax(rx)

	fmt.Printf(" > building %d tiles.. ", len(mtls))
	tt = time.Now()
	saveImg := func(a [][]float64, fp string) {
		img := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{resolution, resolution}})
		for x := 0; x < resolution; x++ {
			for y := 0; y < resolution; y++ {
				if a[x][y] != -9999. {
					// fmt.Println(x, y, rn, rx, a[x][y], fp)
					col, _ := cmap.At(a[x][y])
					img.Set(x, y, col)
				}
			}
		}
		f, _ := os.Create(fp)
		png.Encode(f, img)
	}

	gcell := func(l, h, v float64) int {
		// fmt.Println(l, h, v, (v-l)/(h-l), math.Floor((v-l)/(h-l)*float64(resolution)))
		return int(math.Floor((v - l) / (h - l) * float64(resolution)))
	}
	for t, cs := range mtls {
		n, a := make([][]float64, resolution), make([][]float64, resolution)
		for i := 0; i < resolution; i++ {
			a[i] = make([]float64, resolution)
			n[i] = make([]float64, resolution)
		}

		latUL, longUL, latLR, longLR := t.ToExtent()
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
		saveImg(a, fmt.Sprintf("%s/%d/%d/%d.png", outDir, t.Z, t.X, t.Y))
	}
	fmt.Printf("%v\n", time.Since(tt))
}
