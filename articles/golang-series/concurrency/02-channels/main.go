package main

import (
	"fmt"
	"time"
)

// ============================================================================
// Demo 1: Unbuffered channel -- gửi/nhận đồng bộ hoá lẫn nhau
// ============================================================================

func demoUnbufferedSync() {
	fmt.Println("--- Demo 1: Unbuffered channel đồng bộ hoá ---")
	ch := make(chan string)

	go func() {
		fmt.Println(time.Now().Format("15:04:05.000"), "goroutine: chuẩn bị gửi")
		time.Sleep(300 * time.Millisecond)
		ch <- "xong việc"
		fmt.Println(time.Now().Format("15:04:05.000"), "goroutine: đã gửi xong")
	}()

	fmt.Println(time.Now().Format("15:04:05.000"), "main: gọi <-ch, sẽ BLOCK cho tới khi có người gửi")
	msg := <-ch
	fmt.Println(time.Now().Format("15:04:05.000"), "main: nhận được:", msg)
}

// ============================================================================
// Demo 2: Buffered channel -- gửi không block cho tới khi đầy
// ============================================================================

func demoBufferedChannel() {
	fmt.Println("\n--- Demo 2: Buffered channel ---")
	ch := make(chan int, 3)

	for i := 1; i <= 3; i++ {
		ch <- i
		fmt.Printf("Gửi %d thành công, không block (buffer: %d/%d)\n", i, len(ch), cap(ch))
	}

	fmt.Println("Buffer đã đầy. Nhận trước 1 phần tử để có chỗ gửi tiếp (tránh deadlock):")
	v := <-ch
	fmt.Println("Đã nhận:", v)

	ch <- 4
	fmt.Printf("Gửi 4 thành công (buffer: %d/%d)\n", len(ch), cap(ch))
}

// ============================================================================
// Demo 3: Đóng channel -- nhận sau khi đóng vẫn OK, gửi sau khi đóng panic
// (Xem index.md để thấy lỗi deadlock/panic thật, không nằm trong main.go này)
// ============================================================================

func demoClosingChannel() {
	fmt.Println("\n--- Demo 3: Đóng channel ---")
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	close(ch)

	v1, ok1 := <-ch
	fmt.Println("Nhận lần 1:", v1, ", ok:", ok1)

	v2, ok2 := <-ch
	fmt.Println("Nhận lần 2:", v2, ", ok:", ok2)

	v3, ok3 := <-ch
	fmt.Println("Nhận lần 3 (buffer đã hết):", v3, ", ok:", ok3)

	v4, ok4 := <-ch
	fmt.Println("Nhận lần 4 (vẫn OK, không panic):", v4, ", ok:", ok4)
}

// ============================================================================
// Demo 4: Pipeline qua nhiều channel
// ============================================================================

func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out) // đóng khi đã gửi hết, để tầng sau biết dừng
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in { // range tự động dừng khi in bị đóng
			out <- n * n
		}
	}()
	return out
}

func demoPipeline() {
	fmt.Println("\n--- Demo 4: Pipeline qua nhiều channel ---")
	nums := generate(1, 2, 3, 4, 5)
	squared := square(nums)

	for v := range squared {
		fmt.Println(v)
	}
	fmt.Println("Pipeline kết thúc sạch, không leak, không deadlock")
}

func main() {
	demoUnbufferedSync()
	demoBufferedChannel()
	demoClosingChannel()
	demoPipeline()
}
