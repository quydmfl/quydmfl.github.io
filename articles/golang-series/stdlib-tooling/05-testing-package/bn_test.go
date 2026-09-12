package main

import (
	"testing"
	"time"
)

func expensiveSetup() {
	time.Sleep(50 * time.Millisecond) // giả lập setup tốn kém, chỉ nên chạy 1 lần
}

func BenchmarkWithoutResetTimer(b *testing.B) {
	expensiveSetup() // tính luôn vào thời gian đo -- SAI
	for i := 0; i < b.N; i++ {
		_ = i * i
	}
}

func BenchmarkWithResetTimer(b *testing.B) {
	expensiveSetup()
	b.ResetTimer() // loại bỏ thời gian setup khỏi kết quả đo
	for i := 0; i < b.N; i++ {
		_ = i * i
	}
}
