//go:build !old

package main

import (
	"sync"
	"testing"
)

const bufSize = 4096
const iterations = 5000

var sink []byte

var bufPool = sync.Pool{
	New: func() any { return make([]byte, bufSize) },
}

// BenchmarkBufferReuse (bản MỚI, mặc định) -- tái sử dụng buffer qua sync.Pool
// thay vì cấp phát mới mỗi lần lặp. Chạy: go test -bench . -benchmem -count=10 > new.txt
func BenchmarkBufferReuse(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for i := 0; i < iterations; i++ {
			buf := bufPool.Get().([]byte)
			buf[0] = byte(i)
			sink = buf
			bufPool.Put(buf)
		}
	}
}
