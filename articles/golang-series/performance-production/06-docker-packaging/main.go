package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response là ứng dụng nhỏ dùng để build image trong bài -- đủ để có 1 static
// binary thật (liên hệ bài 02), không phải "hello world" trơn.
type Response struct {
	Message string `json:"message"`
	Time    string `json:"time"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{Message: "OK", Time: time.Now().Format(time.RFC3339)})
	})
	fmt.Println("Server chạy tại :8080")
	http.ListenAndServe(":8080", nil)
}
