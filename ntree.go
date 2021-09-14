// NTrees are an N-dimensional subdividing spacial representation.
// Common dimension-specific types are the 2-dimensional (quadtree) and
// 3-dimensional (octree) variants. This library supports an arbitrary number of
// dimensions, implemented in the same manner as those specific cases.

package ftree

import (
	"errors"
	"strconv"
	"sync"

	"github.com/cznic/mathutil"
)

// Maximum number of dimensions handled by this lib. Currently restricted by
// number of usable bits in Go's int type.
const MaxN = 63

// NTree is a bounding box in N-dimensional space, along with optional data
// and children.
type NTree struct {
	// The bounding n-dimensional box for this ntree. It should always
	// be true that origin[i] += bounds[i] contains p.coords[i].
	center, bounds []float64
	// Optional piece of data to associate with this node. Location may be
	// imprecise on leaf nodes.
	p *Point
	// Slice for child storage, should be 2^n if initialized.
	children []*NTree
	// Used to coordinate write operations, concurrency.
	mutex sync.RWMutex
	// keep track of child point counts under each node, useful for
	// histograms, density predictions, etc.
	count uint64
}

// New creates an ntree root node, using N dimensional slices for
// the center coordinates and relative bounds of the tree space.
// Bounds slice values must be positive, as they define
// a range of center[i] +- bounds[i] for each dimension.
//
// Returns an error if center and bounds don't have the same cardinality,
// or a bounds dimension is <= 0.
func New(center, bounds []float64) (nt *NTree, err error) {
	if len(center) > MaxN {
		return nil, errors.New("64 bit ints limit this library to <= 63 dimensions")
	}
	if len(center) != len(bounds) {
		return nil, errors.New("center and bounds have mismatched lengths")
	}
	if len(center) == 0 {
		return nil, errors.New("Can't have 0-dimensional ntree")
	}
	for i := range bounds {
		if bounds[i] <= 0 {
			return nil, errors.New("Dimension " + strconv.FormatInt(int64(i), 10) +
				" has bounding size <= 0.")
		}
	}
	nt = new(NTree)
	nt.center = center
	nt.bounds = bounds
	return nt, nil
}

// N returns the number of dimensions (N) for this NTree.
func (nt *NTree) N() int {
	return len(nt.center)
}

// Center returns the center coordinates for this NTree node.
func (nt *NTree) Center() []float64 {
	nt.mutex.RLock()
	defer nt.mutex.RUnlock()
	return nt.center
}

// Bounds returns the positive bounding dimensions from center for this NTree
// node. This node covers the entire space of Center() +- Bounds().
func (nt *NTree) Bounds() []float64 {
	nt.mutex.RLock()
	defer nt.mutex.RUnlock()
	return nt.bounds
}

// BoundPoints returns the min and max points for this NTree node.
// This is a shortcut instead of doing the Center() +- Bounds() math manually.
func (nt *NTree) BoundPoints() (min, max []float64) {
	nt.mutex.RLock()
	defer nt.mutex.RUnlock()
	min = make([]float64, len(nt.center))
	max = make([]float64, len(nt.center))
	for i := range nt.center {
		min[i] = nt.center[i] - nt.bounds[i]
		max[i] = nt.center[i] + nt.bounds[i]
	}
	return min, max
}

// Point returns the optional data chunk associated with this NTree nodes.
// This will return null if there's no Point on this node, which should happen
// on any non-leaf node.
func (nt *NTree) Point() *Point {
	nt.mutex.RLock()
	defer nt.mutex.RUnlock()
	return nt.p
}

// Count returns this node's estimate of how many Points lie within it.
func (nt *NTree) Count() uint64 {
	nt.mutex.RLock()
	defer nt.mutex.RUnlock()
	return nt.count
}

// Contains checks if point p is within the bounds of the ntree.
// Returns an error if len(p) != nt.N().
func (nt *NTree) Contains(p *Point) (bool, error) {
	nt.mutex.RLock()
	defer nt.mutex.RUnlock()
	if p == nil {
		return false, errors.New("Point is nil")
	}
	if len(p.Coords) != nt.N() {
		return false, errors.New("Point is " + strconv.FormatUint(uint64(len(p.Coords)), 10) +
			" dimensional, NTree is " + strconv.FormatUint(uint64(nt.N()), 10))
	}
	// fmt.Println(p)
	for i := range p.Coords {
		if (nt.center[i]-nt.bounds[i]) > p.Coords[i] || (nt.center[i]+nt.bounds[i]) < p.Coords[i] {
			return false, nil
		}
	}
	return true, nil
}

// Bitwise operations on array indices are used to keep track of what subset of
// space each child occupies, as described here:
// http://www.brandonpelfrey.com/blog/coding-a-simple-octree/
func hasBit(n int, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}

func setBit(n int, pos uint) int {
	n |= (1 << pos)
	return n
}

// Add inserts a new Point into the NTree. Returns an error on any failure,
// or nil.
func (nt *NTree) Add(p *Point) error {
	in, err := nt.Contains(p)
	if err != nil {
		return err
	}
	if !in {
		return errors.New("Point doesn't fall within bounds of NTree.")
	}
	nt.mutex.Lock()
	if nt.p == nil && nt.children == nil {
		defer nt.mutex.Unlock()
		// simplest case, add to current node
		nt.p = p
		nt.count++
		return nil
	}
	if nt.children != nil {
		defer nt.mutex.Unlock()
		// recurse into children by generating child bounding bitmask
		var target int
		for j := range nt.center {
			if p.Coords[j] > nt.center[j] {
				target = setBit(target, uint(j))
			}
		}
		err = nt.children[target].Add(p)
		if err == nil {
			nt.count++
		}
		return err
	}
	if nt.p != nil && nt.children == nil {
		// create children, re-add current node's Point data, then add Point p.
		size := mathutil.ModPowUint64(2, uint64(nt.N()), mathutil.MaxInt)
		nt.children = make([]*NTree, size)
		// create new child nodes with correct bounds
		for i := range nt.children {
			// determine child dimensions
			center := make([]float64, nt.N())
			bounds := make([]float64, nt.N())
			for j := range center {
				// use bitmask of child index to determine dimension range for child.
				// positive bit means positive range, otherwise negative range.
				if hasBit(i, uint(j)) {
					bounds[j] = nt.bounds[j] / 2.0
					center[j] = nt.center[j] + bounds[j]
				} else {
					bounds[j] = nt.bounds[j] / 2.0
					center[j] = nt.center[j] - bounds[j]
				}
			}
			if nt.children[i], err = New(center, bounds); err != nil {
				return err
			}
		}
		// remove current Point data and re-add so it cascades into child nodes.
		// need to bounce the mutex for this, prossible race condition?
		curP := nt.p
		nt.p = nil
		nt.count--
		nt.mutex.Unlock()
		if err = nt.Add(curP); err != nil {
			return err
		}
		// now add new Point
		return nt.Add(p)
	}
	return nil
}
