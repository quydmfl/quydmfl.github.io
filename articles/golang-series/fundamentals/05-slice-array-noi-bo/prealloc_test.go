package main

import "testing"

const n = 100000

// Đo thật chi phí không cấp capacity trước so với cấp trước bằng make(..., cap).
func BenchmarkAppendNoPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]int, 0)
		for j := 0; j < n; j++ {
			s = append(s, j)
		}
	}
}

func BenchmarkAppendWithPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]int, 0, n)
		for j := 0; j < n; j++ {
			s = append(s, j)
		}
	}
}
