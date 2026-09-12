package main

import "testing"

func BenchmarkWork(b *testing.B) {
	shapes := makeShapes(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		work(shapes)
	}
}
