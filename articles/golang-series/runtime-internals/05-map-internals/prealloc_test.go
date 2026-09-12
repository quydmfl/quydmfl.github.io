package main

import "testing"

func BenchmarkMapNoPrealloc(b *testing.B) {
	for n := 0; n < b.N; n++ {
		m := make(map[int]int)
		for i := 0; i < 10000; i++ {
			m[i] = i
		}
	}
}

func BenchmarkMapPrealloc(b *testing.B) {
	for n := 0; n < b.N; n++ {
		m := make(map[int]int, 10000)
		for i := 0; i < 10000; i++ {
			m[i] = i
		}
	}
}
