package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"
)

type Chunk struct {
	Data [1 << 20]byte // 1MB mỗi chunk
}

// simulateCache giữ 1 cửa sổ trượt các chunk gần đây -- chunk rớt khỏi cửa sổ
// trở thành rác thật, mô phỏng 1 cache/buffer pool thực tế (không phải leak
// thật như context leak ở Stdlib bài 01 -- ở đây rác sinh ra do thiết kế,
// không phải do quên giải phóng).
func simulateCache(window int) {
	var recent []*Chunk
	for i := 0; ; i++ {
		c := &Chunk{}
		for j := 0; j < len(c.Data); j += 4096 {
			c.Data[j] = byte(i) // ghi thật vào từng trang nhớ -- buộc commit thật,
			// không chỉ cấp phát virtual memory (xem index.md để hiểu vì sao bước
			// này bắt buộc phải có để tái hiện đúng OOM-kill thật)
		}
		recent = append(recent, c)
		if len(recent) > window {
			recent = recent[1:] // rớt chunk cũ nhất -- chunk đó thành rác thật
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
	if os.Getenv("SHOW_LIMITS") != "" {
		fmt.Println("NumCPU:", runtime.NumCPU())
		fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))
		fmt.Println("Memory limit (runtime/debug.SetMemoryLimit(-1)):", debug.SetMemoryLimit(-1))
		return
	}
	simulateCache(40) // cửa sổ trượt 40 chunk -- working set thật ~40MB
}
