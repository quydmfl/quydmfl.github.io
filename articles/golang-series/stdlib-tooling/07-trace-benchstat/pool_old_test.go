//go:build old

package main

import "testing"

const bufSize = 4096
const iterations = 5000

var sink []byte

// BenchmarkBufferReuse (bản CŨ) -- cấp phát buffer mới mỗi lần lặp, không
// tái sử dụng. Chạy: go test -tags old -bench . -benchmem -count=10 > old.txt
func BenchmarkBufferReuse(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < iterations; i++ {
			buf := make([]byte, bufSize)
			buf[0] = byte(i)
			sink = buf
		}
	}
}
