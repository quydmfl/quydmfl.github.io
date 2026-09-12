package main

import (
	"sync/atomic"
	"testing"
)

var sharedCounter int64 // dùng atomic -- an toàn khi nhiều subtest cùng ghi

func TestParallelFixed(t *testing.T) {
	cases := []string{"a", "b", "c", "d", "e"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel() // chạy song song với các subtest khác cùng cha
			atomic.AddInt64(&sharedCounter, 1)
		})
	}
}
