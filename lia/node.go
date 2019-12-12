package lia

type node struct {
	z, h, n float64 // elevation, head, manning's n
	fid     []int   // face id
}
