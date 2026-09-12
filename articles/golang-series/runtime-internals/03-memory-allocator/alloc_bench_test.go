package main

import "testing"

type SmallStruct struct {
	A, B int64
}

type LargeStruct struct {
	Data [40000]byte
}

func BenchmarkAllocSmall(b *testing.B) {
	var sink *SmallStruct
	for i := 0; i < b.N; i++ {
		sink = &SmallStruct{A: int64(i)}
	}
	_ = sink
}

func BenchmarkAllocLarge(b *testing.B) {
	var sink *LargeStruct
	for i := 0; i < b.N; i++ {
		sink = &LargeStruct{}
	}
	_ = sink
}
