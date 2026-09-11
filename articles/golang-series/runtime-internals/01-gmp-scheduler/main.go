package main

import (
	"fmt"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// ============================================================================
// Demo 1: GOMAXPROCS mặc định = số CPU logic
// ============================================================================

func demoGOMAXPROCS() {
	fmt.Println("--- Demo 1: NumCPU vs GOMAXPROCS mặc định ---")
	fmt.Println("runtime.NumCPU():", runtime.NumCPU())
	fmt.Println("runtime.GOMAXPROCS(0):", runtime.GOMAXPROCS(0))
}

// ============================================================================
// Demo 2: Hàng chục nghìn goroutine, chỉ vài chục OS thread
// ============================================================================

func demoMultiplexing() {
	fmt.Println("\n--- Demo 2: Multiplexing G lên M ---")
	fmt.Println("Trước khi tạo goroutine:")
	fmt.Println("  NumGoroutine:", runtime.NumGoroutine())
	fmt.Println("  threadcreate profile count:", pprof.Lookup("threadcreate").Count())

	const n = 50000
	var wg sync.WaitGroup
	ready := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			time.Sleep(500 * time.Millisecond)
		}()
	}

	time.Sleep(100 * time.Millisecond) // để mọi goroutine kịp được tạo, đang chờ ở <-ready
	fmt.Printf("\nSau khi tạo %d goroutine (đang chờ, chưa chạy việc):\n", n)
	fmt.Println("  NumGoroutine:", runtime.NumGoroutine())
	fmt.Println("  threadcreate profile count:", pprof.Lookup("threadcreate").Count())

	close(ready) // thả cho tất cả goroutine chạy time.Sleep(500ms) đồng thời
	time.Sleep(100 * time.Millisecond)
	fmt.Println("\nNgay sau khi thả cho tất cả cùng chạy:")
	fmt.Println("  NumGoroutine:", runtime.NumGoroutine())
	fmt.Println("  threadcreate profile count:", pprof.Lookup("threadcreate").Count())

	wg.Wait()
	fmt.Println("\nSau khi tất cả hoàn tất:")
	fmt.Println("  NumGoroutine:", runtime.NumGoroutine())
	fmt.Println("  threadcreate profile count:", pprof.Lookup("threadcreate").Count())
}

// ============================================================================
// Demo 3: GOMAXPROCS quyết định mức song song thật cho CPU-bound workload
// ============================================================================

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func countPrimes(lo, hi int) int {
	count := 0
	for i := lo; i < hi; i++ {
		if isPrime(i) {
			count++
		}
	}
	return count
}

func runPrimeBench(workers int) time.Duration {
	const upper = 3_000_000
	chunk := upper / workers
	start := time.Now()
	var wg sync.WaitGroup
	results := make([]int, workers)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		lo, hi := w*chunk, (w+1)*chunk
		go func(idx, lo, hi int) {
			defer wg.Done()
			results[idx] = countPrimes(lo, hi)
		}(w, lo, hi)
	}
	wg.Wait()
	return time.Since(start)
}

func demoGOMAXPROCSImpact() {
	fmt.Println("\n--- Demo 3: GOMAXPROCS ảnh hưởng CPU-bound workload thật ---")
	original := runtime.GOMAXPROCS(0)
	fmt.Println("GOMAXPROCS mặc định của máy:", original)

	runtime.GOMAXPROCS(1)
	d1 := runPrimeBench(8)
	fmt.Println("GOMAXPROCS=1, 8 worker đếm số nguyên tố dưới 3 triệu:", d1.Round(time.Millisecond))

	runtime.GOMAXPROCS(original)
	d2 := runPrimeBench(8)
	fmt.Printf("GOMAXPROCS=%d, 8 worker đếm số nguyên tố dưới 3 triệu: %v\n", original, d2.Round(time.Millisecond))
}

func main() {
	demoGOMAXPROCS()
	demoMultiplexing()
	demoGOMAXPROCSImpact()
}
