package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // side-effect import: tự đăng ký /debug/pprof/* vào DefaultServeMux
	"os/exec"
	"strings"
	"time"
)

// ============================================================================
// Handler 1: CPU hot path -- thuật toán kiểm tra số nguyên tố cố tình chậm
// ============================================================================

func isPrimeSlow(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i < n; i++ { // chậm: lẽ ra chỉ cần chạy tới sqrt(n)
		if n%i == 0 {
			return false
		}
	}
	return true
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	count := 0
	for i := 2; i < 8000; i++ {
		if isPrimeSlow(i) {
			count++
		}
	}
	fmt.Fprintf(w, "Số nguyên tố: %d\n", count)
}

// ============================================================================
// Handler 2: cấp phát thừa mỗi lần gọi -- hot path cho heap profile
// ============================================================================

type Payload struct {
	Data [1024]byte
}

func allocHandler(w http.ResponseWriter, r *http.Request) {
	var items []*Payload
	for i := 0; i < 5000; i++ {
		items = append(items, &Payload{})
	}
	fmt.Fprintf(w, "Đã cấp phát %d item\n", len(items))
}

// ============================================================================
// Handler 3: context leak thật -- liên hệ bài 01 (context package)
// ============================================================================

func leakHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel // THỰC TẾ: "che" cảnh báo lostcancel của go vet nhưng vẫn leak thật -- xem index.md
	go func() { <-ctx.Done() }()
	fmt.Fprintf(w, "leaked\n")
}

// ============================================================================
// Hạ tầng demo: fire load + gọi go tool pprof như 1 subprocess
// ============================================================================

func fireLoad(url string, n int, duration time.Duration) {
	deadline := time.Now().Add(duration)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		for i := 0; i < n; i++ {
			resp, err := client.Get(url)
			if err == nil {
				resp.Body.Close()
			}
		}
	}
}

func runPprof(args ...string) string {
	cmd := exec.Command("go", append([]string{"tool", "pprof"}, args...)...)
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func printPprofOutput(out string) {
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) != "" && !strings.Contains(line, "Fetching") && !strings.Contains(line, "Saved profile") {
			fmt.Println(line)
		}
	}
}

func main() {
	http.HandleFunc("/work", slowHandler)
	http.HandleFunc("/alloc", allocHandler)
	http.HandleFunc("/leak", leakHandler)
	go http.ListenAndServe("127.0.0.1:16060", nil)
	time.Sleep(200 * time.Millisecond)

	fmt.Println("--- CPU profile: fire load vào /work trong 3s, chụp CPU profile ---")
	go fireLoad("http://127.0.0.1:16060/work", 1, 3*time.Second)
	printPprofOutput(runPprof("-top", "-nodecount", "6", "http://127.0.0.1:16060/debug/pprof/profile?seconds=3"))

	fmt.Println("\n--- Heap profile: fire 100 request vào /alloc, chụp heap profile ---")
	fireLoad("http://127.0.0.1:16060/alloc", 100, 1*time.Millisecond)
	printPprofOutput(runPprof("-top", "-alloc_space", "-nodecount", "5", "http://127.0.0.1:16060/debug/pprof/heap"))

	fmt.Println("\n--- Goroutine profile: fire 30 request vào /leak, xem goroutine tích luỹ ---")
	fireLoad("http://127.0.0.1:16060/leak", 30, 1*time.Millisecond)
	resp, _ := http.Get("http://127.0.0.1:16060/debug/pprof/goroutine?debug=1")
	buf := make([]byte, 256)
	n, _ := resp.Body.Read(buf)
	resp.Body.Close()
	fmt.Println(string(buf[:n]))
}
