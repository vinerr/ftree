# ftree
An n-dimensional tree implementation in Go, patterned after Quad and Octrees.

# ntree - an n-dimensional tree structure in Go.

This library is an abstraction of the design behind
[Quadtrees](https://en.wikipedia.org/wiki/Quadtree) and
[Octrees](https://en.wikipedia.org/wiki/Octree), useful data structures in 2d
and 3d graphics. Those cases are n=2 and n=3 specific cases for this library,
which should handle any case up to 63 dimensions.

Note that each additional dimension added makes the tree structure exponentially
slower, so cases over about 16 dimensions are not recommended.

## Install

This library is fully installable via go get:

  go get -u -v github.com/nergdron/ntree

## Tests and Benchmarks

There are useful unit tests and benchmarks included. Run them with the
standard go test command:

  go test -bench=.*

You can also verify thread safety by running the same tests with more
active CPUs:

  go test -bench=.* -cpu 4

You should find that Add() becomes slightly slower per op with more concurrency
due to lock contention, but Search() becomes linearly faster due to parallelism.

## License

NTree is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kdtree is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with kdtree.  If not, see
[http://www.gnu.org/licenses/](http://www.gnu.org/licenses/).

