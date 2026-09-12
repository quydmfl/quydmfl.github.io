package main

import "testing"

func directCall() int {
	return 42
}

func withDeferOnce() int {
	defer func() {}()
	return 42
}

func withDeferInLoop() int {
	sum := 0
	for i := 0; i < 8; i++ {
		defer func() { sum++ }() // defer trong loop -- không thoả open-coded, rơi về cài đặt cũ
	}
	return sum
}

func BenchmarkDirectCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		directCall()
	}
}

func BenchmarkWithDeferOnce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		withDeferOnce()
	}
}

func BenchmarkWithDeferInLoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		withDeferInLoop()
	}
}
