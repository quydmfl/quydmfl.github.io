package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// slowHandler giả lập 1 request tốn thời gian xử lý (2 giây) -- đủ dài để
// quan sát rõ hành vi shutdown.
func slowHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handler: bắt đầu xử lý, sẽ mất 2 giây...")
	time.Sleep(2 * time.Second)
	fmt.Println("Handler: xử lý xong, trả response")
	fmt.Fprintln(w, "Done")
}

func main() {
	srv := &http.Server{Addr: ":18083", Handler: http.HandlerFunc(slowHandler)}

	// signal.NotifyContext: context tự Done() khi nhận SIGTERM/SIGINT
	// (liên hệ Stdlib bài 01 -- context package)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	go func() {
		fmt.Println("Server đang chạy tại :18083")
		srv.ListenAndServe()
	}()
	time.Sleep(200 * time.Millisecond)

	// Client gọi request trong goroutine riêng, giống 1 request thật đang xử lý dở
	go func() {
		resp, err := http.Get("http://localhost:18083/")
		if err != nil {
			fmt.Println("Client: lỗi:", err)
			return
		}
		defer resp.Body.Close()
		fmt.Println("Client: nhận response thành công")
	}()

	// Tự gửi SIGTERM cho chính process sau 500ms để mô phỏng orchestrator
	// (Kubernetes, systemd...) yêu cầu dừng service giữa lúc đang có request dở.
	time.Sleep(500 * time.Millisecond)
	fmt.Println(">>> Tự gửi SIGTERM cho chính mình (mô phỏng orchestrator dừng service) <<<")
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGTERM)

	<-ctx.Done()
	fmt.Println("Đã nhận tín hiệu shutdown, đang chờ request xử lý xong...")

	// http.Server.Shutdown: ngừng nhận connection mới, CHỜ request đang xử lý
	// hoàn tất trong deadline cho trước -- không cắt ngang giữa chừng.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	srv.Shutdown(shutdownCtx)
	fmt.Println("Server đã dừng sạch sau:", time.Since(start).Round(time.Millisecond))
}
