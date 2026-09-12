package main

import (
	"fmt"
	"runtime"
	"sync"
)

// ============================================================================
// Demo 2: Chi phí stack ban đầu mỗi goroutine (~2KB)
// ============================================================================

func demoStackCost() {
	fmt.Println("\n--- Demo 2: Chi phí stack ban đầu mỗi goroutine ---")
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	const n = 10000
	ready := make(chan struct{})
	done := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			<-done
		}()
	}
	close(ready)

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	diff := after.StackInuse - before.StackInuse
	fmt.Println("StackInuse trước:", before.StackInuse, "bytes")
	fmt.Println("StackInuse sau khi tạo", n, "goroutine:", after.StackInuse, "bytes")
	fmt.Println("Chênh lệch:", diff, "bytes, trung bình mỗi goroutine:", diff/n, "bytes")

	close(done)
	wg.Wait()
}

// ============================================================================
// Demo 1: Stack tự lớn dần theo độ sâu đệ quy
// ============================================================================

func reportStack(depth int) {
	// Khai báo runtime.MemStats trong 1 hàm RIÊNG (không nằm trong recurseGrowth)
	// để struct lớn của nó không cộng dồn vào kích thước frame của mỗi tầng đệ quy.
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("Độ sâu %6d: StackInuse = %10d bytes\n", depth, ms.StackInuse)
}

func recurseGrowth(depth, maxDepth int) {
	var localArray [256]byte // tiêu tốn thêm không gian mỗi tầng
	localArray[0] = byte(depth)

	if depth%20000 == 0 {
		reportStack(depth)
	}
	if depth >= maxDepth {
		return
	}
	recurseGrowth(depth+1, maxDepth)
}

func demoStackGrowth() {
	fmt.Println("--- Demo 1: Stack tự lớn dần theo độ sâu đệ quy ---")
	recurseGrowth(0, 100000)
	fmt.Println("Hoàn tất 100000 tầng đệ quy, không crash")
}

// ============================================================================
// Demo 3: Tính đúng đắn của biến cục bộ trên stack sau nhiều lần grow
// ============================================================================

func recurseIntegrity(depth, maxDepth int) int {
	localValue := depth // biến cục bộ thật, không escape lên heap
	if depth >= maxDepth {
		return localValue
	}
	return localValue + recurseIntegrity(depth+1, maxDepth)
}

func demoStackIntegrity() {
	fmt.Println("\n--- Demo 3: Tính đúng đắn dữ liệu stack sau nhiều lần grow ---")
	const maxDepth = 200000
	sum := recurseIntegrity(0, maxDepth)
	expected := maxDepth * (maxDepth + 1) / 2
	fmt.Println("Đệ quy", maxDepth, "tầng, tổng tính được:", sum)
	fmt.Println("Công thức đóng n(n+1)/2 =", expected)
	if sum == expected {
		fmt.Println("KHỚP -- mọi biến cục bộ trên stack vẫn đúng sau nhiều lần stack grow")
	} else {
		fmt.Println("SAI LỆCH -- chênh:", expected-sum)
	}
}

func main() {
	demoStackGrowth()
	demoStackCost()
	demoStackIntegrity()
}
