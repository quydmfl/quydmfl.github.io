package main

import (
	"fmt"
	"os"
	"runtime/trace"
	"sync"
	"time"
)

func cpuBurn(d time.Duration) {
	end := time.Now().Add(d)
	x := 0
	for time.Now().Before(end) {
		x++
	}
	_ = x
}

// demoTrace tạo 20 goroutine CPU-bound (mỗi goroutine "đốt" CPU 50ms) và ghi
// lại toàn bộ sự kiện scheduler vào trace.out -- nối lại kịch bản GOMAXPROCS
// ở Runtime Internals bài 01, lần này quan sát qua go tool trace thay vì
// GODEBUG=schedtrace.
func demoTrace() {
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		panic(err)
	}
	defer trace.Stop()

	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cpuBurn(50 * time.Millisecond)
		}()
	}
	wg.Wait()
	fmt.Println("Đã ghi trace.out, tổng thời gian:", time.Since(start).Round(time.Millisecond))
}

func main() {
	demoTrace()
}
