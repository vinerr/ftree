package ftree

// Point information to be stored in an NTree leaf node.
type Point struct {
	Coords []float64
	// Arbitrary data attached to this Point.
	Data *interface{}
}
