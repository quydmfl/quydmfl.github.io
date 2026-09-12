package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Demo 1: log/slog -- cùng 1 log call, output khác nhau theo handler
// ============================================================================

func demoSlogHandlers() {
	fmt.Println("--- Demo 1: slog Text handler vs JSON handler ---")
	textLogger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	textLogger.Info("user logged in", "user_id", 42, "ip", "10.0.0.5")
	jsonLogger.Info("user logged in", "user_id", 42, "ip", "10.0.0.5")
}

// ============================================================================
// Demo 2: Correlation -- request ID qua context, giữ nguyên qua nhiều tầng
// và nhiều request chạy đồng thời (liên hệ Stdlib bài 01 -- context package)
// ============================================================================

type ctxKey string

const requestIDKey ctxKey = "request_id"

func loggerFromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return base.With("request_id", reqID)
	}
	return base
}

func handleRequest(ctx context.Context, base *slog.Logger, delay time.Duration) {
	log := loggerFromContext(ctx, base)
	log.Info("bắt đầu xử lý")
	time.Sleep(delay)
	log.Info("xử lý xong")
}

func demoCorrelation() {
	fmt.Println("\n--- Demo 2: correlation qua request_id, nhiều request đồng thời ---")
	base := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var wg sync.WaitGroup
	requests := []struct {
		id    string
		delay time.Duration
	}{
		{"req-A", 30 * time.Millisecond},
		{"req-B", 10 * time.Millisecond},
		{"req-C", 20 * time.Millisecond},
	}
	for _, r := range requests {
		wg.Add(1)
		go func(id string, delay time.Duration) {
			defer wg.Done()
			ctx := context.WithValue(context.Background(), requestIDKey, id)
			handleRequest(ctx, base, delay)
		}(r.id, r.delay)
	}
	wg.Wait()
}

// ============================================================================
// Demo 3: metrics cơ bản qua expvar -- đếm request/error, expose /debug/vars
// ============================================================================

var (
	requestCount int64
	errorCount   int64
)

func init() {
	expvar.Publish("request_count", expvar.Func(func() any {
		return atomic.LoadInt64(&requestCount)
	}))
	expvar.Publish("error_count", expvar.Func(func() any {
		return atomic.LoadInt64(&errorCount)
	}))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&requestCount, 1)
	if r.URL.Query().Get("fail") == "1" {
		atomic.AddInt64(&errorCount, 1)
		http.Error(w, "simulated error", http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "OK")
}

func demoExpvarMetrics() {
	fmt.Println("\n--- Demo 3: metrics qua expvar ---")
	http.HandleFunc("/", metricsHandler)
	go http.ListenAndServe(":19092", nil)
	time.Sleep(200 * time.Millisecond)

	client := &http.Client{}
	for i := 0; i < 8; i++ {
		client.Get("http://localhost:19092/")
	}
	for i := 0; i < 2; i++ {
		client.Get("http://localhost:19092/?fail=1")
	}

	resp, _ := client.Get("http://localhost:19092/debug/vars")
	var raw map[string]json.RawMessage
	json.NewDecoder(resp.Body).Decode(&raw)
	fmt.Println("request_count:", string(raw["request_count"]))
	fmt.Println("error_count:", string(raw["error_count"]))
}

func main() {
	demoSlogHandlers()
	demoCorrelation()
	demoExpvarMetrics()
}
