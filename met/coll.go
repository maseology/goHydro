package met

import "time"

// Coll is an alias hold temporal data
// type Coll = map[time.Time]map[int]float64

// Coll holds met data
type Coll struct {
	T []time.Time
	D [][][]float64
}
