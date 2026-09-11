package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================================================
// Demo 1: Chi phí thật của 100.000 goroutine
// ============================================================================

func demoGoroutineCost() {
	fmt.Println("--- Demo 1: Chi phí thật của 100.000 goroutine ---")
	const n = 100000

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(3 * time.Second) // đủ dài để đo tại đỉnh, chưa goroutine nào kịp xong
		}()
	}

	fmt.Println("Số goroutine ngay sau vòng lặp tạo:", runtime.NumGoroutine())
	creationTime := time.Since(start)
	fmt.Println("Thời gian riêng để TẠO xong 100.000 goroutine:", creationTime)

	var memPeak runtime.MemStats
	runtime.ReadMemStats(&memPeak)

	wg.Wait()
	fmt.Printf("RAM tăng thêm (HeapAlloc): %.2f MB cho %d goroutine (~%.0f byte/goroutine)\n",
		float64(memPeak.HeapAlloc-memBefore.HeapAlloc)/1024/1024,
		n,
		float64(memPeak.HeapAlloc-memBefore.HeapAlloc)/float64(n))
}

// ============================================================================
// Demo 2: Goroutine leak thật
// ============================================================================

func leaky() {
	ch := make(chan int)
	go func() {
		val := <-ch // không ai bao giờ gửi vào ch -- goroutine kẹt mãi mãi
		fmt.Println(val)
	}()
}

func demoGoroutineLeak() {
	fmt.Println("\n--- Demo 2: Goroutine leak thật ---")
	fmt.Println("Trước khi gọi leaky():", runtime.NumGoroutine())

	for i := 0; i < 5; i++ {
		leaky()
		time.Sleep(10 * time.Millisecond)
		fmt.Printf("Sau lần gọi leaky() thứ %d: %d goroutine\n", i+1, runtime.NumGoroutine())
	}

	fmt.Println("Gọi thêm 100 lần nữa...")
	for i := 0; i < 100; i++ {
		leaky()
	}
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Sau 100 lần gọi thêm:", runtime.NumGoroutine(), "-- không bao giờ tự giảm")
}

// ============================================================================
// Demo 3: Fix leak bằng done channel
// ============================================================================

func notLeaky(done <-chan struct{}) {
	ch := make(chan int)
	go func() {
		select {
		case val := <-ch:
			fmt.Println(val)
		case <-done: // luôn có đường thoát
			return
		}
	}()
}

func demoFixedWithDoneChannel() {
	fmt.Println("\n--- Demo 3: Fix bằng done channel ---")
	done := make(chan struct{})
	fmt.Println("Trước:", runtime.NumGoroutine())

	for i := 0; i < 5; i++ {
		notLeaky(done)
	}
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Sau 5 lần gọi notLeaky() (chưa đóng done):", runtime.NumGoroutine())

	close(done)
	time.Sleep(10 * time.Millisecond)
	fmt.Println("Sau khi close(done):", runtime.NumGoroutine(), "-- quay lại đúng như ban đầu")
}

func main() {
	demoGoroutineCost()
	demoGoroutineLeak()
	demoFixedWithDoneChannel()
}
