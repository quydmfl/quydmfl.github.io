package main

import (
	"fmt"
	"math"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

type Node struct {
	Data [256]byte
	Next *Node
}

// allocGarbage tạo áp lực GC thật: cấp phát liên tục, thỉnh thoảng rời bỏ
// tham chiếu (head = nil) để tạo rác thật cho GC dọn.
func allocGarbage(n int) {
	var head *Node
	for i := 0; i < n; i++ {
		head = &Node{Next: head}
		if i%1000 == 0 {
			head = nil
		}
	}
	runtime.KeepAlive(head)
}

// ============================================================================
// Demo 1: GC pause thật -- concurrent GC chỉ dừng chương trình vài chục micro giây
// ============================================================================

func demoGCBasics() {
	fmt.Println("--- Demo 1: GC pause thật qua runtime.ReadMemStats ---")
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	allocGarbage(2_000_000)

	runtime.ReadMemStats(&after)
	numGC := after.NumGC - before.NumGC
	totalPause := after.PauseTotalNs - before.PauseTotalNs
	fmt.Println("Số lần GC chạy:", numGC)
	fmt.Println("Tổng thời gian pause:", totalPause, "ns")
	if numGC > 0 {
		fmt.Println("Trung bình mỗi lần pause:", totalPause/uint64(numGC), "ns")
	}
}

// ============================================================================
// Demo 2: GOGC quyết định tần suất GC -- đánh đổi CPU vs bộ nhớ đỉnh
// ============================================================================

func runWithGOGC(percent int, label string) {
	debug.SetGCPercent(percent)
	runtime.GC()

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()

	allocGarbage(2_000_000)

	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)

	fmt.Printf("%s: %d lần GC, tổng pause %d ns, thời gian chạy %v, HeapAlloc cuối %d bytes\n",
		label, after.NumGC-before.NumGC, after.PauseTotalNs-before.PauseTotalNs,
		elapsed.Round(time.Millisecond), after.HeapAlloc)
}

func demoGOGCCompare() {
	fmt.Println("\n--- Demo 2: GOGC quyết định tần suất GC ---")
	runWithGOGC(100, "GOGC=100 (mặc định)")
	runWithGOGC(400, "GOGC=400          ")
	runWithGOGC(-1, "GOGC=off          ")
}

// ============================================================================
// Demo 3: GOMEMLIMIT vẫn ép GC chạy dù GOGC đã tắt
// ============================================================================

func demoGOMEMLIMIT() {
	fmt.Println("\n--- Demo 3: GOMEMLIMIT ép GC chạy dù GOGC=off ---")
	debug.SetGCPercent(100)
	debug.SetMemoryLimit(math.MaxInt64)
	debug.FreeOSMemory() // trả bộ nhớ đã cấp phát dư từ Demo 2 (GOGC=off) về OS trước khi đo tiếp

	debug.SetGCPercent(-1)         // tắt hẳn GOGC-based trigger
	debug.SetMemoryLimit(30 << 20) // nhưng đặt trần cứng 30MB qua GOMEMLIMIT

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	allocGarbage(2_000_000)
	runtime.ReadMemStats(&after)

	fmt.Println("GOGC=off nhưng GOMEMLIMIT=30MB:")
	fmt.Println("  Số lần GC:", after.NumGC-before.NumGC)
	fmt.Println("  HeapAlloc cuối:", after.HeapAlloc, "bytes (giữ dưới trần 30MB)")

	// trả lại cấu hình mặc định cho demo tiếp theo
	debug.SetGCPercent(100)
	debug.SetMemoryLimit(math.MaxInt64)
}

// ============================================================================
// Demo 4: Giảm GC pressure bằng sync.Pool tái sử dụng buffer
// ============================================================================

const bufSize = 4096
const iterations = 500000

var sink []byte // buộc buf phải thực sự tồn tại, tránh bị compiler loại bỏ alloc

func processNoPool() {
	for i := 0; i < iterations; i++ {
		buf := make([]byte, bufSize) // alloc mới mỗi lần, giá trị cũ thành rác ngay
		buf[0] = byte(i)
		sink = buf
	}
}

var bufPool = sync.Pool{
	New: func() any {
		return make([]byte, bufSize)
	},
}

func processWithPool() {
	for i := 0; i < iterations; i++ {
		buf := bufPool.Get().([]byte)
		buf[0] = byte(i)
		sink = buf
		bufPool.Put(buf)
	}
}

func measurePool(label string, fn func()) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	fn()
	runtime.ReadMemStats(&after)
	fmt.Printf("%s: %d lần GC, tổng pause %d ns, TotalAlloc tăng %d bytes\n",
		label, after.NumGC-before.NumGC, after.PauseTotalNs-before.PauseTotalNs,
		after.TotalAlloc-before.TotalAlloc)
}

func demoSyncPool() {
	fmt.Println("\n--- Demo 4: Giảm GC pressure bằng sync.Pool ---")
	measurePool("Không dùng Pool (alloc mới mỗi lần)", processNoPool)
	measurePool("Dùng sync.Pool (tái sử dụng buffer)", processWithPool)
	runtime.KeepAlive(sink)
}

func main() {
	demoGCBasics()
	demoGOGCCompare()
	demoGOMEMLIMIT()
	demoSyncPool()
}
