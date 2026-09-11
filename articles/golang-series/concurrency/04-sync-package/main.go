package main

import (
	"fmt"
	"sync"
)

// ============================================================================
// Demo 1: Mutex -- bảo vệ shared state đơn giản
// ============================================================================

func demoMutex() {
	fmt.Println("--- Demo 1: Mutex ---")

	counterUnsafe := 0
	var wg1 sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			counterUnsafe++ // đọc-cộng-ghi, không nguyên tử
		}()
	}
	wg1.Wait()
	fmt.Println("Không khoá, 1000 lần tăng, kết quả:", counterUnsafe, "(thường < 1000)")

	counterSafe := 0
	var mu sync.Mutex
	var wg2 sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()
			mu.Lock()
			counterSafe++
			mu.Unlock()
		}()
	}
	wg2.Wait()
	fmt.Println("Có Mutex, 1000 lần tăng, kết quả:", counterSafe, "(luôn đúng = 1000)")
}

// ============================================================================
// Demo 2: WaitGroup -- đúng cách và bẫy Add trong goroutine
// (Xem index.md để thấy cảnh báo DATA RACE thật từ -race)
// ============================================================================

func demoWaitGroupCorrect() {
	fmt.Println("\n--- Demo 2: WaitGroup dùng đúng cách ---")
	var wg sync.WaitGroup
	var mu sync.Mutex
	completed := 0

	for i := 0; i < 5; i++ {
		wg.Add(1) // ĐÚNG: Add gọi TRƯỚC go func(), trong main goroutine
		go func(n int) {
			defer wg.Done()
			mu.Lock()
			completed++
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	fmt.Println("completed =", completed, "(luôn đúng = 5)")
}

// ============================================================================
// Demo 3: Once -- đảm bảo 1 đoạn code chỉ chạy đúng 1 lần
// ============================================================================

func demoOnce() {
	fmt.Println("\n--- Demo 3: sync.Once ---")
	var once sync.Once
	var wg sync.WaitGroup
	var mu sync.Mutex
	initCount := 0

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			once.Do(func() {
				mu.Lock()
				initCount++
				mu.Unlock()
				fmt.Println("init chạy, gọi bởi goroutine", n)
			})
		}(i)
	}
	wg.Wait()
	fmt.Println("Số lần init thực sự chạy:", initCount, "(dù 10 goroutine cùng gọi once.Do)")
}

func main() {
	demoMutex()
	demoWaitGroupCorrect()
	demoOnce()
}
