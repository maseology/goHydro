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
	"github.com/maseology/mmio"
)

func (g GMET) SaveGob(fp string) error {
	f, err := os.Create(fp)
	defer f.Close()
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(g)
	if err != nil {
		return err
	}
	return nil
}

func LoadGob(fp string) (*GMET, error) {
	var g GMET
	f, err := os.Open(fp)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	enc := gob.NewDecoder(f)
	err = enc.Decode(&g)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func LoadNC(fp string, vars []string) (*GMET, error) {
	tt := mmio.NewTimer()

	nc, err := netcdf.Open(fp)
	if err != nil {
		log.Fatalln(err)
	}
	defer nc.Close()
	// fmt.Println(nc.ListVariables())

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
		Nts:  len(times),
		Nsta: len(sids),
		Ts:   times,
		Sids: sids,
	}
	tt.Lap("\n loading complete")

	g.Dat = func() [][]dset {
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
		tx := getDat(vars[0])
		tn := getDat(vars[1])
		rf := getDat(vars[2])
		sf := getDat(vars[3])
		sm := getDat(vars[4])
		pa := getDat(vars[5])

		o := make([][]dset, g.Nsta)
		for i := 0; i < g.Nsta; i++ {
			o[i] = make([]dset, g.Nts)
			for j, t := range times {
				o[i][j] = dset{
					Date: t.Format("2006-01-02"),
					Tx:   float64(tx[j][i]),
					Tn:   float64(tn[j][i]),
					Rf:   float64(rf[j][i]),
					Sf:   float64(sf[j][i]),
					Sm:   float64(sm[j][i]),
					Pa:   float64(pa[j][i]),
				}
			}
		}
		return o
	}()
	tt.Lap("ordering complete")

	return &g, nil
}

func LoadBin(prfx string, vars []string) (*GMET, error) { // go at the time of writing did not have the ability to read large netCDF4 files. Use FEWS/netcdf/nc4dailyToDat.py to translate files.
	tt := mmio.NewTimer()

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

	d := func() map[string][][]float32 { // [variable][time][station]
		d := make(map[string][][]float32, len(vars))
		for _, v := range vars {
			dat := getDat(v)
			fmt.Println(v, len(dat), len(dat[0]))
			d[v] = dat
		}
		return d
	}()
	tt.Lap("\n python arrays loading complete")

	g.Dat = func() [][]dset { // [station][row]
		dsets := make([][]dset, nsta)
		for i := 0; i < nsta; i++ {
			dsets[i] = make([]dset, nts)
			for j, t := range times {
				dsets[i][j] = dset{
					Date: t.Format("2006-01-02"),
					Tx:   float64(d[vars[0]][j][i]),
					Tn:   float64(d[vars[1]][j][i]),
					Rf:   float64(d[vars[2]][j][i]),
					Sf:   float64(d[vars[3]][j][i]),
					Sm:   float64(d[vars[4]][j][i]),
					Pa:   float64(d[vars[5]][j][i]),
				}
			}
		}
		return dsets
	}()
	tt.Lap("ordering complete")

	return &g, nil
}
