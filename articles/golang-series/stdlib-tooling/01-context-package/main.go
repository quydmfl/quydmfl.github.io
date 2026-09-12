package main

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// ============================================================================
// Demo 1: WithTimeout/WithCancel -- ctx.Done() đóng khi nào, ctx.Err() báo gì
// ============================================================================

func demoBasics() {
	fmt.Println("--- Demo 1: WithTimeout vs WithCancel ---")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	<-ctx.Done()
	fmt.Println("WithTimeout: ctx.Done() đóng sau", time.Since(start).Round(time.Millisecond), "| ctx.Err():", ctx.Err())

	ctx2, cancel2 := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel2()
	}()
	start2 := time.Now()
	<-ctx2.Done()
	fmt.Println("WithCancel : ctx.Done() đóng sau", time.Since(start2).Round(time.Millisecond), "| ctx.Err():", ctx2.Err())
}

// ============================================================================
// Demo 2: Dùng cancel() đúng cách -- goroutine tự thoát, không tích luỹ
// (bản LEAK -- bỏ qua cancel() -- nằm ở file riêng leak.go, vì go vet chặn thẳng)
// ============================================================================

func worker(ctx context.Context) {
	<-ctx.Done()
}

func demoFixedCancel() {
	fmt.Println("\n--- Demo 2: Luôn gọi cancel() -- goroutine tự thoát ---")
	runtime.GC()
	before := runtime.NumGoroutine()
	fmt.Println("NumGoroutine trước:", before)

	for i := 0; i < 1000; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		go worker(ctx)
		cancel()
	}
	time.Sleep(100 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Println("NumGoroutine sau 1000 worker CÓ cancel():", after, "(không tích luỹ, tự thoát hết)")
}

// ============================================================================
// Demo 3: WithValue -- truyền dữ liệu request-scoped qua nhiều tầng hàm
// ============================================================================

type ctxKey string

const requestIDKey ctxKey = "requestID"

func handleRequest(ctx context.Context) {
	processStep1(ctx)
}

func processStep1(ctx context.Context) {
	processStep2(ctx)
}

func processStep2(ctx context.Context) {
	reqID, ok := ctx.Value(requestIDKey).(string)
	fmt.Println("Ở tầng sâu nhất (processStep2), đọc được requestID:", reqID, "| ok:", ok)
}

func demoWithValue() {
	fmt.Println("\n--- Demo 3: WithValue truyền qua nhiều tầng hàm ---")
	ctx := context.WithValue(context.Background(), requestIDKey, "req-12345")
	handleRequest(ctx)

	v, ok := context.Background().Value(requestIDKey).(string)
	fmt.Println("Đọc key không tồn tại trên context khác: giá trị =", fmt.Sprintf("%q", v), "| ok:", ok)
}

// ============================================================================
// Demo 4: Context tree -- huỷ gốc lan truyền xuống toàn bộ con/cháu
// ============================================================================

func demoContextTree() {
	fmt.Println("\n--- Demo 4: Huỷ context gốc lan truyền xuống toàn bộ cây ---")
	root, cancelRoot := context.WithCancel(context.Background())
	child1, cancel1 := context.WithCancel(root)
	child2, cancel2 := context.WithCancel(root)
	grandchild, cancel3 := context.WithCancel(child1)
	defer cancel1()
	defer cancel2()
	defer cancel3()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var exitOrder []string

	watch := func(name string, ctx context.Context) {
		defer wg.Done()
		<-ctx.Done()
		mu.Lock()
		exitOrder = append(exitOrder, name)
		mu.Unlock()
	}

	wg.Add(3)
	go watch("child1", child1)
	go watch("child2", child2)
	go watch("grandchild (con của child1)", grandchild)

	time.Sleep(50 * time.Millisecond)
	fmt.Println("Huỷ context GỐC (root)...")
	start := time.Now()
	cancelRoot()

	wg.Wait()
	fmt.Println("Toàn bộ con + cháu thoát sau:", time.Since(start))
	fmt.Println("Thứ tự thoát:", exitOrder)
}

func main() {
	demoBasics()
	demoFixedCancel()
	demoWithValue()
	demoContextTree()
}
