package main

import "fmt"

// ============================================================================
// Demo: pollorder ngẫu nhiên hoá -- khi nhiều case cùng sẵn sàng, select
// không luôn ưu tiên case đầu tiên hay bất kỳ thứ tự cố định nào.
// ============================================================================

func demoSelectFairness() {
	fmt.Println("--- Demo: pollorder ngẫu nhiên hoá trong select ---")
	counts := make(map[string]int)
	const trials = 100000
	for i := 0; i < trials; i++ {
		ch1 := make(chan int, 1)
		ch2 := make(chan int, 1)
		ch3 := make(chan int, 1)
		ch1 <- 1
		ch2 <- 1
		ch3 <- 1

		select {
		case <-ch1:
			counts["ch1"]++
		case <-ch2:
			counts["ch2"]++
		case <-ch3:
			counts["ch3"]++
		}
	}
	fmt.Println("Cả 3 channel LUÔN sẵn sàng cùng lúc, chạy", trials, "lần select:")
	fmt.Println("  ch1 được chọn:", counts["ch1"], "lần")
	fmt.Println("  ch2 được chọn:", counts["ch2"], "lần")
	fmt.Println("  ch3 được chọn:", counts["ch3"], "lần")
}

func main() {
	demoSelectFairness()
}
