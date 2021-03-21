package routing

import "github.com/maseology/mmaths"

type Path struct {
	Segments []*mmaths.Node
	Verts    [][][3]float64
}

func BuildPaths(nodes []*mmaths.Node, coords [][3]float64) *Path {
	jns := mmaths.Junctions(nodes)
	isjn := make(map[*mmaths.Node]bool, len(jns))
	for _, jn := range jns {
		isjn[jn] = true
	}

	queue := make([]*mmaths.Node, 0)
	for _, ln := range mmaths.Leaves(nodes) {
		queue = append(queue, ln) // push sinks/leaves/headwaters
	}

	for {
		if len(queue) == 0 {
			break
		}

		// pop
		x := queue[0]
		queue = queue[1:]

		// push
		for _, dn := range x.DS {
			if _, ok := isjn[dn]; ok {

			} else {
				queue = append(queue, dn)
			}
		}
	}
	return &Path{}
}
