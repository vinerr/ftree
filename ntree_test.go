package ftree

import (
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup

func init() {
	// Use different values every test to try and catch edge cases.
	rand.Seed(time.Now().Unix())
}

func TestInit(t *testing.T) {
	// test all available dimensional combinations
	for n := 1; n <= MaxN; n++ {
		t.Log("Testing", n, "dimension(s).")
		center := make([]float64, n)
		bounds := make([]float64, n)
		for i := range center {
			center[i] = 0
			bounds[i] = 1
		}
		nt, err := New(center, bounds)
		if err != nil {
			t.Fatal(err)
		}
		if nt.N() != n {
			t.Fatal(n, "dimensional tree returned n=", nt.N())
		}
	}
}

func TestAdd(t *testing.T) {
	count := 500
	max := runtime.GOMAXPROCS(-1)
	// > 16 dimensions quickly becomes impractical for memory and processing reasons.
	for n := 1; n <= 16; n++ {
		t.Log("Testing", n, "dimension(s).")
		center := make([]float64, n)
		bounds := make([]float64, n)
		p1 := make([]float64, n)
		for i := range center {
			center[i] = 0.0
			bounds[i] = 1.0
			p1[i] = -1.0
		}
		nt, err := New(center, bounds)
		if err != nil {
			t.Fatal(err, n)
		}
		wg.Add(max)
		for i := 0; i < max; i++ {
			go func() {
				for j := 0; j < count/max; j++ {
					p := new(Point)
					p.Coords = make([]float64, n)
					for j := range p.Coords {
						p.Coords[j] = (rand.Float64() * 2.0) - 1.0
					}
					if err = nt.Add(p); err != nil {
						t.Error(err, n, i)
					}
				}
				wg.Done()
			}()
		}
		count = (count / max) * max

		wg.Wait()
		points, err := nt.Search(p1, bounds)
		if err != nil {
			t.Fatal(err)
		}
		if len(points) != count {
			t.Fatal("Search returned", len(points), "points instead of", count)
		}
		if nt.Count() != uint64(count) {
			t.Error("Tree estimated incorrect, sees", nt.Count(), "points instead of", count)
		}
	}
}

func benchAdd(b *testing.B, n int) {
	b.StopTimer()
	max := runtime.GOMAXPROCS(-1)
	center := make([]float64, n)
	bounds := make([]float64, n)
	for i := range center {
		center[i] = 0
		bounds[i] = 1
	}
	nt, err := New(center, bounds)
	if err != nil {
		b.Fatal(err)
	}
	wg.Add(max)
	b.StartTimer()
	for i := 0; i < max; i++ {
		go func() {
			for j := 0; j < b.N/max; j++ {
				p := new(Point)
				p.Coords = make([]float64, n)
				for j := range p.Coords {
					p.Coords[j] = (rand.Float64() * 2.0) - 1.0
				}
				if err = nt.Add(p); err != nil {
					b.Error(err)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func benchSearch(b *testing.B, n int) {
	b.StopTimer()
	max := runtime.GOMAXPROCS(-1)
	center := make([]float64, n)
	bounds := make([]float64, n)
	for i := range center {
		center[i] = 0
		bounds[i] = 1
	}
	nt, err := New(center, bounds)
	if err != nil {
		b.Fatal(err)
	}
	// 10k points gives a reasonable search space.
	for i := 0; i < 10000; i++ {
		p := new(Point)
		p.Coords = make([]float64, n)
		for j := range p.Coords {
			p.Coords[j] = (rand.Float64() * 2.0) - 1.0
		}
		if err = nt.Add(p); err != nil {
			b.Fatal(err)
		}
	}
	var swap float64
	var count uint64
	wg.Add(max)

	b.StartTimer()
	for i := 0; i < max; i++ {
		go func() {
			p1 := make([]float64, n)
			p2 := make([]float64, n)
			for j := 0; j < (b.N / max); j++ {
				// generate random search area
				for j := range p1 {
					p1[j] = (rand.Float64() * 2.0) - 1.0
					p2[j] = (rand.Float64() * 2.0) - 1.0
					if p1[j] > p2[j] {
						swap = p1[j]
						p1[j] = p2[j]
						p2[j] = swap
					}
				}
				// find points within it!
				points, err := nt.Search(p1, p2)
				if err != nil {
					b.Fatal(err)
				}
				count += uint64(len(points))
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if count == 0 && b.N > 100 {
		b.Log("Failed to find any points after", b.N, "searches.")
	}
}

func BenchmarkQuadtreeAdd(b *testing.B) {
	benchAdd(b, 2)
}

func BenchmarkQuadtreeSearch(b *testing.B) {
	benchSearch(b, 2)
}

func BenchmarkOctreeAdd(b *testing.B) {
	benchAdd(b, 3)
}

func BenchmarkOcttreeSearch(b *testing.B) {
	benchSearch(b, 3)
}
