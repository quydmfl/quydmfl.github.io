package main

import (
	"strconv"
	"testing"
)

func BenchmarkMapIntKey(b *testing.B) {
	m := make(map[int]int, 10000)
	for i := 0; i < 10000; i++ {
		m[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i%10000]
	}
}

func BenchmarkMapStringKey(b *testing.B) {
	m := make(map[string]int, 10000)
	keys := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		keys[i] = strconv.Itoa(i)
		m[keys[i]] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[keys[i%10000]]
	}
}
