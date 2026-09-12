package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response là ví dụ 1 server nhỏ dùng encoding/json + net/http -- đủ phức tạp
// để có kích thước binary thực tế, không phải "hello world" trơn.
type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{Message: "OK", Status: 200})
	})
	fmt.Println("Server demo cho bài build & binary optimization")
	_ = http.ListenAndServe // không thực sự chạy server trong demo này
}
