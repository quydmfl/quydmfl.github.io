package main

import (
	"fmt"
	"sort"
	"sync"
)

// ============================================================================
// Demo 1: Thứ tự duyệt map không đảm bảo -- cố tình, không phải bug
// (Chạy file này nhiều lần để tự thấy thứ tự đổi mỗi lần)
// ============================================================================

func demoIterationOrder() {
	fmt.Println("--- Demo 1: Thứ tự duyệt map (chạy lại nhiều lần để so sánh) ---")
	m := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}
	fmt.Print("Thứ tự ngẫu nhiên: ")
	for k := range m {
		fmt.Print(k, " ")
	}
	fmt.Println()

	// Muốn thứ tự ổn định: lấy riêng key, sort, rồi duyệt theo thứ tự đó
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Print("Thứ tự ổn định (đã sort): ")
	for _, k := range keys {
		fmt.Print(k, " ")
	}
	fmt.Println()
}

// ============================================================================
// Demo 2: Map an toàn với sync.RWMutex -- fix cho concurrent map write
// (Xem index.md để thấy "fatal error: concurrent map writes" thật khi KHÔNG khoá)
// ============================================================================

func demoSafeConcurrentMap() {
	fmt.Println("\n--- Demo 2: Map an toàn với sync.RWMutex ---")
	m := make(map[int]int)
	var mu sync.RWMutex
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			mu.Lock()
			m[n] = n
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	mu.RLock()
	fmt.Println("Hoàn tất, không crash. len(m) =", len(m))
	mu.RUnlock()
}

func main() {
	demoIterationOrder()
	demoSafeConcurrentMap()
}
