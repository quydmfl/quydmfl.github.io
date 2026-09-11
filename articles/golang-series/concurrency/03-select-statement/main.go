package main

import (
	"fmt"
	"time"
)

// ============================================================================
// Demo 1: select cơ bản -- chờ nhiều channel, xử lý cái nào sẵn sàng trước
// ============================================================================

func demoBasicSelect() {
	fmt.Println("--- Demo 1: select cơ bản ---")
	fast := make(chan string)
	slow := make(chan string)

	go func() {
		time.Sleep(100 * time.Millisecond)
		fast <- "fast xong"
	}()
	go func() {
		time.Sleep(400 * time.Millisecond)
		slow <- "slow xong"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg := <-fast:
			fmt.Println(time.Now().Format("15:04:05.000"), "nhận từ fast:", msg)
		case msg := <-slow:
			fmt.Println(time.Now().Format("15:04:05.000"), "nhận từ slow:", msg)
		}
	}
}

// ============================================================================
// Demo 2: Timeout pattern với time.After
// ============================================================================

func slowCall() <-chan string {
	ch := make(chan string)
	go func() {
		time.Sleep(3 * time.Second) // giả lập 1 lời gọi rất chậm
		ch <- "kết quả (sẽ không ai nhận vì đã timeout)"
	}()
	return ch
}

func demoTimeoutPattern() {
	fmt.Println("\n--- Demo 2: Timeout pattern ---")
	start := time.Now()
	select {
	case res := <-slowCall():
		fmt.Println("nhận được:", res)
	case <-time.After(1 * time.Second):
		fmt.Println("TIMEOUT sau", time.Since(start).Round(time.Millisecond))
	}
}

// ============================================================================
// Demo 3: default case cho non-blocking operation
// ============================================================================

func demoDefaultCase() {
	fmt.Println("\n--- Demo 3: default case (non-blocking) ---")
	ch := make(chan int)
	select {
	case v := <-ch:
		fmt.Println("có dữ liệu:", v)
	default:
		fmt.Println("chưa có dữ liệu, không chờ")
	}

	ch2 := make(chan int, 1)
	ch2 <- 42
	select {
	case v := <-ch2:
		fmt.Println("có dữ liệu:", v)
	default:
		fmt.Println("chưa có dữ liệu, không chờ")
	}
}

// ============================================================================
// Demo 4: time.NewTicker -- cách đúng cho công việc lặp lại định kỳ
// (thay vì gọi time.After lặp lại trong vòng lặp, tạo timer mới mỗi lần)
// ============================================================================

func demoTicker() {
	fmt.Println("\n--- Demo 4: time.NewTicker cho công việc định kỳ ---")
	ticker := time.NewTicker(200 * time.Millisecond) // tạo đúng 1 lần, tái sử dụng
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(900 * time.Millisecond)
		done <- true
	}()

	count := 0
	for {
		select {
		case <-done:
			fmt.Println("Hoàn tất, đã tick", count, "lần")
			return
		case t := <-ticker.C:
			count++
			fmt.Println("tick", count, "lúc", t.Format("15:04:05.000"))
		}
	}
}

func main() {
	demoBasicSelect()
	demoTimeoutPattern()
	demoDefaultCase()
	demoTicker()
}
