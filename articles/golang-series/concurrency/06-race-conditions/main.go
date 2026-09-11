package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
)

// ============================================================================
// Demo 1: shared counter -- không khoá (race, xem index.md) vs atomic (an toàn)
// ============================================================================

func demoAtomicCounter() {
	fmt.Println("--- Demo 1: shared counter, sửa bằng atomic ---")
	var counter int64
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}
	wg.Wait()
	fmt.Println("Dùng atomic:", counter, "(luôn đúng = 1000)")
}

// ============================================================================
// Demo 2: closure capture biến loop -- Go 1.22+ an toàn mặc định
// ============================================================================

func demoLoopVarCapture() {
	fmt.Println("\n--- Demo 2: closure capture biến loop ---")
	fmt.Println("go version:", runtime.Version())

	var results []int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			results = append(results, i) // Go 1.22+: mỗi vòng lặp có bản i riêng
			mu.Unlock()
		}()
	}
	wg.Wait()
	sort.Ints(results)
	fmt.Println("for i := 0; i < 5 (Go 1.22+, an toàn mặc định):", results)
}

// ============================================================================
// Demo 3: biến NGOÀI vòng lặp bị chia sẻ -- Go 1.22 KHÔNG cứu được kiểu này
// (bản race thật nằm ở sharedvar.go riêng, xem index.md)
// ============================================================================

func demoSharedVarFixed() {
	fmt.Println("\n--- Demo 3: biến ngoài vòng lặp, sửa bằng truyền tham số ---")
	i := 0
	var results []int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i < 5 {
		wg.Add(1)
		go func(n int) { // truyền qua tham số -- mỗi goroutine có bản copy riêng
			defer wg.Done()
			mu.Lock()
			results = append(results, n)
			mu.Unlock()
		}(i)
		i++
	}
	wg.Wait()
	sort.Ints(results)
	fmt.Println("Truyền qua tham số (an toàn mọi phiên bản Go):", results)
}

// ============================================================================
// Demo 4: map dùng chung -- sửa bằng Mutex
// (bản crash thật "fatal error: concurrent map writes" nằm ở mapwrite.go riêng)
// ============================================================================

func demoMapFixed() {
	fmt.Println("\n--- Demo 4: map dùng chung, sửa bằng Mutex ---")
	m := make(map[int]int)
	var mu sync.Mutex
	var wg sync.WaitGroup
	for r := 0; r < 20; r++ {
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				mu.Lock()
				m[n%50] = n * n
				mu.Unlock()
			}(i)
		}
	}
	wg.Wait()
	fmt.Println("Hoàn tất, không crash. len(m) =", len(m))
}

// ============================================================================
// Demo 5: slice dùng chung -- sửa bằng Mutex
// (bản race thật gây mất phần tử nằm ở slicewrite.go riêng, xem index.md)
// ============================================================================

func demoSliceFixed() {
	fmt.Println("\n--- Demo 5: slice dùng chung, sửa bằng Mutex ---")
	var s []int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			mu.Lock()
			s = append(s, n)
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Println("Có Mutex, len(s):", len(s), "(luôn đúng = 1000)")
}

func main() {
	demoAtomicCounter()
	demoLoopVarCapture()
	demoSharedVarFixed()
	demoMapFixed()
	demoSliceFixed()
}
