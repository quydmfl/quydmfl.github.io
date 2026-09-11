package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// work mô phỏng 1 job "chậm" -- gọi API, query DB, đọc file lớn...
func work(n int) int {
	time.Sleep(50 * time.Millisecond)
	return n * n
}

// trackPeakGoroutines lấy mẫu runtime.NumGoroutine() liên tục cho tới khi
// nhận tín hiệu trên stop, trả về con trỏ tới giá trị đỉnh (cập nhật bằng CAS
// vì goroutine đo và goroutine main chạy song song).
func trackPeakGoroutines(stop <-chan struct{}) *int64 {
	var peak int64
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				n := int64(runtime.NumGoroutine())
				for {
					old := atomic.LoadInt64(&peak)
					if n <= old || atomic.CompareAndSwapInt64(&peak, old, n) {
						break
					}
				}
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()
	return &peak
}

// ============================================================================
// Demo 1: Không giới hạn -- 1 goroutine cho mỗi job
// ============================================================================

func unlimitedGoroutines(jobs []int) []int {
	results := make([]int, len(jobs))
	var wg sync.WaitGroup
	for i, j := range jobs {
		wg.Add(1)
		go func(idx, n int) {
			defer wg.Done()
			results[idx] = work(n)
		}(i, j)
	}
	wg.Wait()
	return results
}

// ============================================================================
// Demo 2: Worker pool -- số worker cố định, fan-out job / fan-in result
// ============================================================================

func WorkerPool(jobs []int, numWorkers int) []int {
	jobCh := make(chan int, len(jobs))
	resultCh := make(chan int, len(jobs))
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh { // fan-out: nhiều worker cùng đọc 1 channel job
				resultCh <- work(j)
			}
		}()
	}

	for _, j := range jobs {
		jobCh <- j
	}
	close(jobCh) // báo hết job, worker sẽ thoát vòng for-range khi rút cạn jobCh

	go func() {
		wg.Wait()      // chờ toàn bộ worker xong
		close(resultCh) // an toàn để đóng vì chắc chắn không ai còn gửi vào resultCh
	}()

	var results []int
	for r := range resultCh { // fan-in: gom kết quả từ mọi worker về 1 nơi
		results = append(results, r)
	}
	return results
}

func runDemo(label string, numWorkers int, jobs []int) {
	stop := make(chan struct{})
	peak := trackPeakGoroutines(stop)
	start := time.Now()

	var results []int
	if numWorkers <= 0 {
		results = unlimitedGoroutines(jobs)
	} else {
		results = WorkerPool(jobs, numWorkers)
	}

	elapsed := time.Since(start)
	close(stop)
	time.Sleep(5 * time.Millisecond) // để tracker kịp lấy mẫu sau khi việc chính xong

	fmt.Println(label)
	fmt.Printf("Thời gian: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("Peak NumGoroutine: %d\n", atomic.LoadInt64(peak))
	fmt.Printf("Số kết quả: %d\n\n", len(results))
}

func main() {
	jobs := make([]int, 200)
	for i := range jobs {
		jobs[i] = i
	}

	fmt.Println("--- Demo 1: Không giới hạn (1 goroutine / job) ---")
	runDemo("200 job, mỗi job 50ms, không giới hạn goroutine", 0, jobs)

	fmt.Println("--- Demo 2: Worker pool, 10 worker cố định ---")
	runDemo("200 job, mỗi job 50ms, worker pool 10 worker", 10, jobs)

	fmt.Println("--- Demo 3: Worker pool, 50 worker ---")
	runDemo("200 job, mỗi job 50ms, worker pool 50 worker", 50, jobs)
}
