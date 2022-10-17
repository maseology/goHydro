package gmet

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/batchatco/go-native-netcdf/netcdf"
)

func (g GMET) SaveGob(fp string) error {
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(g)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func LoadGob(fp string) (*GMET, error) {
	var g GMET
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	enc := gob.NewDecoder(f)
	err = enc.Decode(&g)
	if err != nil {
		return nil, err
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func LoadNC(fp string, vars []string) (*GMET, error) {
	// tt := time.Now()

	nc, err := netcdf.Open(fp)
	if err != nil {
		log.Fatalln(err)
	}
	defer nc.Close()
	fmt.Println(nc.ListVariables()) //[8:])

	sids := func() []int {
		svr, err := nc.GetVariable("station_id")
		if err != nil {
			log.Fatalln("station_id", err)
		}
		so, has := svr.Values.([]string)
		if !has {
			log.Fatalln("station_id", err)
		}
		o := make([]int, len(so))
		for i, s := range so {
			if o[i], err = strconv.Atoi(strings.Trim(s, "\x00")); err != nil {
				log.Fatalln("station_id", err)
			}
		}
		return o
	}()

	times := func() []time.Time {
		svr, err := nc.GetVariable("time")
		if err != nil {
			log.Fatalln("time", err)
		}
		tu, has := svr.Values.([]float64)
		if !has {
			log.Fatalln("time", err)
		}

		tims := make([]time.Time, len(tu))
		for i, v := range tu {
			tims[i] = time.Unix(int64(v)*60, 0)
		}

		return tims
	}()

	g := GMET{
		Nts:   len(times),
		Nsta:  len(sids),
		Ts:    times,
		Sids:  sids,
		Snams: vars,
	}
	// fmt.Printf(" %s\t%s ", time.Since(tt), " loading complete\n")
	// tt = time.Now()

	g.Sxy = func() []XYZ {
		slon, err := nc.GetVariable("lon")
		if err != nil {
			fmt.Println(" coordinates", err)
			return nil
		}
		slat, err := nc.GetVariable("lat")
		if err != nil {
			fmt.Println(" coordinates", err)
			return nil
		}
		z, err := nc.GetVariable("z")
		if err != nil {
			fmt.Println(" elevations", err)
			z = nil
		}

		o := make([]XYZ, g.Nsta)
		switch slat.Values.(type) {
		case []float64:
			sslat := slat.Values.([]float64)
			sslon := slon.Values.([]float64)
			if z != nil {
				sz := z.Values.([]float64)
				for i := 0; i < g.Nsta; i++ {
					o[i] = XYZ{sslon[i], sslat[i], sz[i]}
				}
			} else {
				for i := 0; i < g.Nsta; i++ {
					o[i] = XYZ{sslon[i], sslat[i], -9999.}
				}
			}
		case []float32:
			sslat := slat.Values.([]float32)
			sslon := slon.Values.([]float32)
			if z != nil {
				sz := z.Values.([]float32)
				for i := 0; i < g.Nsta; i++ {
					o[i] = XYZ{float64(sslon[i]), float64(sslat[i]), float64(sz[i])}
				}
			} else {
				for i := 0; i < g.Nsta; i++ {
					o[i] = XYZ{float64(sslon[i]), float64(sslat[i]), -9999.}
				}
			}
		default:
			panic("unknown type")
		}
		return o
	}()

	g.Dat = func() [][]DSet {
		getDat := func(v string) [][]float32 {
			svr, err := nc.GetVariable(v)
			if err != nil {
				log.Fatalln(v, err)
			}
			// fmt.Println(svr.Values)
			fs, has := svr.Values.([][]float32)
			if !has {
				log.Fatalln(v, err)
			}
			return fs
		}

		dd := make([][][]float32, len(vars))
		for k, s := range vars {
			dd[k] = getDat(s)
		}
		o := make([][]DSet, g.Nsta)
		for i := 0; i < g.Nsta; i++ {
			o[i] = make([]DSet, g.Nts)
			for j, t := range times {
				d := make([]float64, len(vars))
				for k := range vars {
					d[k] = float64(dd[k][j][i])
				}
				o[i][j] = DSet{
					Date: t.Format("2006-01-02 15:04:05 -0700 MST"),
					Dat:  d,
				}
			}
		}
		return o
	}()
	// fmt.Printf(" %s\t%s ", time.Since(tt), " building complete\n")

	return &g, nil
}

func LoadBin(prfx string, vars []string) (*GMET, error) { // go at the time of writing did not have the ability to read large netCDF4 files. Use FEWS/netcdf/nc4dailyToDat.py to translate files.
	tt := time.Now()

	nsta, sids := func() (int, []int) {
		f, err := os.Open(prfx + ".sta")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		nsta := fi.Size() / 4

		sids := make([]int32, nsta)
		reader := bufio.NewReader(f)
		if binary.Read(reader, binary.LittleEndian, &sids) != nil {
			fmt.Println("binary.Read failed:", err)
		}
		return int(nsta), func() []int {
			o := make([]int, len(sids))
			for i, v := range sids {
				o[i] = int(v)
			}
			return o
		}()
	}()

	nts, times := func() (int, []time.Time) {
		f, err := os.Open(prfx + ".tim")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		nts := fi.Size() / 8

		itims := make([]int64, nts)
		reader := bufio.NewReader(f)
		if binary.Read(reader, binary.LittleEndian, &itims) != nil {
			fmt.Println("binary.Read failed:", err)
		}

		tims := make([]time.Time, nts)
		for i := 0; i < int(nts); i++ {
			tims[i] = time.Unix(0, itims[i])
		}

		return int(nts), tims
	}()

	g := GMET{
		Nts:  nts,
		Nsta: nsta,
		Ts:   times,
		Sids: sids,
	}

	getDat := func(stval string) [][]float32 {
		f, err := os.Open(prfx + "." + stval)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		if fi.Size()/4 != int64(nts*nsta) {
			log.Fatal("size error with " + stval)
		}

		dat := make([]float32, nts*nsta)
		reader := bufio.NewReader(f)
		if err := binary.Read(reader, binary.LittleEndian, dat); err != nil {
			log.Fatal("binary.Read failed:", err)
		}

		out, k := make([][]float32, nts), 0
		for i := 0; i < nts; i++ {
			out[i] = make([]float32, nsta)
			for j := 0; j < nsta; j++ {
				out[i][j] = dat[k]
				k++
			}
		}

		return out
	}

	dat := func() map[string][][]float32 { // [variable][time][station]
		d := make(map[string][][]float32, len(vars))
		for _, v := range vars {
			dat := getDat(v)
			fmt.Println(" ", v, len(dat), len(dat[0]))
			d[v] = dat
		}
		return d
	}()
	// fmt.Printf("%s\n%s ", time.Since(tt), "\n python arrays loading complete")
	fmt.Printf("  > %s:  %s\n", time.Since(tt), "python arrays loading complete")
	tt = time.Now()

	fmt.Print(" ordering..")
	g.Dat = func() [][]DSet { // [station][row]
		dsets := make([][]DSet, nsta)
		for i := 0; i < nsta; i++ {
			dsets[i] = make([]DSet, nts)
			for j, t := range times {
				d := make([]float64, len(vars))
				for k, s := range vars {
					d[k] = float64(dat[s][j][i])
				}
				dsets[i][j] = DSet{
					Date: t.Format("2006-01-02"),
					Dat:  d,
				}
			}
		}
		return dsets
	}()
	// fmt.Printf("%s\n%s ", time.Since(tt), "ordering complete")
	fmt.Printf("  %s:  %s\n", time.Since(tt), "complete")

	return &g, nil
}
