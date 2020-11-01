package met

import (
	"fmt"
	"time"
)

// Coll is an alias hold temporal data
// type Coll = map[time.Time]map[int]float64

// Coll holds met data
type Coll struct {
	T []time.Time   // [date ID]
	D [][][]float64 // D [date ID][location ID][type ID] or [cell ID][date ID][type ID]
}

// Print prints the data in tabular form
func (c *Coll) Print(wbl []string) {
	const brk = 20
	h := fmt.Sprintf("%30s", "Date")
	for _, hh := range wbl {
		h += fmt.Sprintf("%16s", hh)
	}
	fmt.Println(h)
	for i, dt := range c.T {
		if i >= brk && i < len(c.T)-brk {
			if i == brk {
				fmt.Println("  ....")
			}
			continue
		}
		s := fmt.Sprintf("%30v", dt)
		for j := range wbl {
			s += fmt.Sprintf("%16.3f", c.D[i][0][j])
		}
		fmt.Println(s)
	}
	fmt.Println()
}

// Get returns a column of values
func (c *Coll) Get(loc, col int) ([]time.Time, []float64) {
	d, o := make([]time.Time, len(c.T)), make([]float64, len(c.T))
	for i, dt := range c.T {
		d[i] = dt
		o[i] = c.D[i][loc][col]
	}
	return d, o
}
