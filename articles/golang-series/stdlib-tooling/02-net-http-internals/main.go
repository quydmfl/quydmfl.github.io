package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"time"
)

// ============================================================================
// Demo 1: http.Client mặc định KHÔNG có timeout
// ============================================================================

func demoClientTimeout() {
	fmt.Println("--- Demo 1: http.Client mặc định không có timeout ---")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second) // giả lập server chậm
		w.Write([]byte("OK"))
	}))
	defer srv.Close()

	start := time.Now()
	resp, err := http.Get(srv.URL)
	if err != nil {
		fmt.Println("Lỗi:", err)
	} else {
		resp.Body.Close()
		fmt.Println("DefaultClient thành công sau:", time.Since(start).Round(time.Millisecond), "(chờ đúng bằng server, không tự bỏ cuộc)")
	}

	client := &http.Client{Timeout: 300 * time.Millisecond}
	start2 := time.Now()
	_, err2 := client.Get(srv.URL)
	fmt.Println("Client{Timeout: 300ms} kết thúc sau:", time.Since(start2).Round(time.Millisecond), "| lỗi:", err2)
}

// ============================================================================
// Demo 2: Mỗi *http.Request mang context riêng, huỷ khi client ngắt kết nối
// ============================================================================

func demoRequestContext() {
	fmt.Println("\n--- Demo 2: r.Context() huỷ khi client ngắt kết nối ---")
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			fmt.Println("Server: phát hiện client đã ngắt kết nối, dừng xử lý sớm. Err:", r.Context().Err())
			close(done)
		case <-time.After(5 * time.Second):
			w.Write([]byte("OK"))
		}
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 200 * time.Millisecond}
	start := time.Now()
	_, err := client.Get(srv.URL)
	fmt.Println("Client: request kết thúc sau", time.Since(start).Round(time.Millisecond), "| lỗi:", err)

	<-done
}

// ============================================================================
// Demo 3: http.Server.ReadTimeout ngắt kết nối với client gửi quá chậm
// ============================================================================

func demoServerTimeout() {
	fmt.Println("\n--- Demo 3: http.Server.ReadTimeout ngắt client chậm ---")
	srv := &http.Server{
		Addr:        "127.0.0.1:18765",
		ReadTimeout: 300 * time.Millisecond,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			fmt.Println("Handler: đọc được", len(body), "byte trước khi lỗi:", err)
		}),
	}
	go srv.ListenAndServe()
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", "127.0.0.1:18765")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	body := "field=" + string(make([]byte, 5000))
	header := fmt.Sprintf("POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: %d\r\n\r\n", len(body))
	conn.Write([]byte(header))

	start := time.Now()
	for i := 0; i < len(body); i++ {
		if _, err := conn.Write([]byte{body[i]}); err != nil {
			fmt.Println("Client: ghi lỗi sau", time.Since(start).Round(time.Millisecond), "-- server đã đóng do ReadTimeout:", err)
			break
		}
		time.Sleep(5 * time.Millisecond) // gửi cực chậm, 1 byte / 5ms
	}

	srv.Close()
}

func main() {
	demoClientTimeout()
	demoRequestContext()
	demoServerTimeout()
}
