package ftree

import "errors"

// Iter runs function f on every node in the tree.
func (nt *NTree) Iter(f func(n *NTree)) {
	nt.mutex.Lock()
	defer nt.mutex.Unlock()
	f(nt)
	if nt.children != nil {
		for i := range nt.children {
			f(nt.children[i])
		}
	}
}

// Search finds all Points falling within the bounding box between p1 and p2.
// Note that it is assumed for every dimension i, p1[i] <= p2[i].
// This is an inclusive search, so Points whose coordinates are equal to
// the supplied bounds in a given dimension will match.
//
// Returns nil, error if the length of p1, p2 don't match
// nt.N().
func (nt *NTree) Search(p1, p2 []float64) (points []*Point, err error) {
	nt.mutex.RLock()
	defer nt.mutex.RUnlock()
	if len(p1) != len(p2) || len(p2) != nt.N() {
		return nil, errors.New("Bounding points have different dimensions than tree.")
	}
	if nt.children == nil && nt.p != nil {
		// check local Point for leaf node
		for i := range nt.p.Coords {
			if nt.p.Coords[i] < p1[i] || nt.p.Coords[i] > p2[i] {
				return nil, nil
			}
		}
		return []*Point{nt.p}, nil
	}
	if nt.children != nil {
		// check children for matching bounds, and collect matching points from them.
		points = make([]*Point, 0)
		for _, child := range nt.children {
			for i := range p1 {
				s1 := child.center[i] - child.bounds[i]
				s2 := child.center[i] + child.bounds[i]
				// skip this child if outside this dimension's bounds
				if (s1 < p1[i] && s2 < p1[i]) || (s1 > p2[i] && s2 > p2[i]) {
					continue
				}
			}
			p, err := child.Search(p1, p2)
			if err != nil {
				return nil, err
			}
			points = append(points, p...)
		}
	}
	return points, nil
}
