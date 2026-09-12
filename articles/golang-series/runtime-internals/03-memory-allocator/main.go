package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
)

// ============================================================================
// Demo 1: Size class -- alloc bị làm tròn lên kích thước cố định gần nhất
// ============================================================================

type Exact16 struct {
	A, B int64 // đúng 16 byte, khớp thẳng 1 size class
}

type Odd17 struct {
	Data [17]byte // 17 byte -- không khớp size class nào, phải làm tròn lên
}

type Odd40 struct {
	Data [40]byte // 40 byte -- cũng phải làm tròn lên
}

func measureAvgAlloc(n int, alloc func() any) uint64 {
	keep := make([]any, n) // cấp phát mảng giữ kết quả TRƯỚC, ngoài phạm vi đo

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < n; i++ {
		keep[i] = alloc()
	}
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(keep)
	return (after.HeapAlloc - before.HeapAlloc) / uint64(n)
}

func demoSizeClass() {
	fmt.Println("--- Demo 1: Size class -- alloc bị làm tròn lên ---")
	const n = 1_000_000

	avg16 := measureAvgAlloc(n, func() any { return &Exact16{} })
	fmt.Println("Exact16 (16 byte logic): trung bình", avg16, "bytes/alloc (khớp thẳng, không lãng phí)")

	avg17 := measureAvgAlloc(n, func() any { return &Odd17{} })
	fmt.Println("Odd17   (17 byte logic): trung bình", avg17, "bytes/alloc (làm tròn lên size class 24)")

	avg40 := measureAvgAlloc(n, func() any { return &Odd40{} })
	fmt.Println("Odd40   (40 byte logic): trung bình", avg40, "bytes/alloc (làm tròn lên size class 48)")
}

// ============================================================================
// Demo 2: Alloc lớn (>32KB) đi thẳng qua mheap, làm tròn theo trang bộ nhớ
// ============================================================================

type BigObject struct {
	Data [40000]byte // 40000 byte > 32KB -- vượt ngưỡng size class, đi thẳng mheap
}

func demoLargeAlloc() {
	fmt.Println("\n--- Demo 2: Alloc lớn (>32KB) đi thẳng qua mheap ---")
	const n = 1000
	keep := make([]*BigObject, n)

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < n; i++ {
		keep[i] = &BigObject{}
	}
	runtime.ReadMemStats(&after)
	runtime.KeepAlive(keep)

	avg := (after.HeapAlloc - before.HeapAlloc) / uint64(n)
	fmt.Println("BigObject (40000 byte logic): trung bình", avg, "bytes/alloc (làm tròn theo trang bộ nhớ, không theo size class)")
}

// ============================================================================
// Demo 3: Escape analysis -- ai quyết định biến nằm stack hay heap
// ============================================================================

// escapesToHeap: trả về con trỏ tới biến cục bộ -- biến "thoát" khỏi hàm,
// buộc phải cấp phát trên heap (nếu không, con trỏ trỏ vào vùng nhớ đã thu hồi)
func escapesToHeap() *int {
	x := 42
	return &x
}

// staysOnStack: biến cục bộ chỉ dùng nội bộ, không có con trỏ nào thoát ra ngoài
func staysOnStack() int {
	y := 42
	return y
}

func demoEscapeAnalysis() {
	fmt.Println("\n--- Demo 3: Escape analysis ---")
	p := escapesToHeap()
	v := staysOnStack()
	fmt.Println("escapesToHeap() ->", *p, "| staysOnStack() ->", v)
	fmt.Println("Chạy 'go build -gcflags=\"-m\" .' để xem quyết định thật của compiler")
}

// ============================================================================
// Demo 4: pprof allocation profile thật
// ============================================================================

type Payload struct {
	Data [128]byte
}

func allocateLots() {
	var keep []*Payload
	for i := 0; i < 500000; i++ {
		keep = append(keep, &Payload{})
	}
	runtime.KeepAlive(keep)
}

func demoHeapProfile() {
	fmt.Println("\n--- Demo 4: pprof heap profile thật ---")
	allocateLots()

	f, err := os.Create("mem.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	runtime.GC()
	if err := pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}
	fmt.Println("Đã ghi mem.prof -- xem bằng: go tool pprof -top -alloc_space -nodecount 8 mem.prof")
}

func main() {
	demoSizeClass()
	demoLargeAlloc()
	demoEscapeAnalysis()
	demoHeapProfile()
}
