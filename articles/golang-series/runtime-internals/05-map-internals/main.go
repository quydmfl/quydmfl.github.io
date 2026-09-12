package main

import (
	"fmt"
	"time"
)

// ============================================================================
// Demo 1: Thứ tự duyệt map bị random hoá cố ý
// ============================================================================

func demoRandomOrder() {
	fmt.Println("--- Demo 1: Thứ tự duyệt map bị random hoá cố ý ---")
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
	for run := 0; run < 3; run++ {
		var order []string
		for k := range m {
			order = append(order, k)
		}
		fmt.Println("Lần", run+1, ":", order)
	}
}

// ============================================================================
// Demo 2: Chi phí mỗi lần insert khi map tự lớn dần (growth incremental)
// ============================================================================

func demoGrowthLatency() {
	fmt.Println("\n--- Demo 2: Chi phí insert khi map tự lớn dần ---")
	m := make(map[int]int) // không preallocate, để tự grow
	checkpoints := map[int]bool{
		8: true, 16: true, 32: true, 64: true, 128: true,
		256: true, 512: true, 1024: true, 2048: true, 4096: true, 8192: true,
	}

	lastTime := time.Now()
	for i := 0; i < 10000; i++ {
		start := time.Now()
		m[i] = i
		elapsed := time.Since(start)
		if checkpoints[i+1] {
			fmt.Printf("Sau %5d phần tử: insert thứ %d mất %v (từ mốc trước: %v)\n",
				i+1, i+1, elapsed, time.Since(lastTime))
			lastTime = time.Now()
		}
	}
}

func main() {
	demoRandomOrder()
	demoGrowthLatency()
}
