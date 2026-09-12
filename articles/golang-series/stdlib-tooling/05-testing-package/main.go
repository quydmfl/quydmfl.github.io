package main

import "fmt"

// Add và Reverse là 2 hàm nhỏ dùng làm đối tượng test xuyên suốt bài viết --
// xem table_test.go, fuzz_test.go, parallel_test.go, bn_test.go.

func Add(a, b int) int {
	return a + b
}

// Reverse đảo ngược 1 chuỗi UTF-8 theo RUNE (ký tự), không phải byte -- xử lý
// đúng cả ký tự multi-byte (tiếng Việt có dấu, tiếng Hebrew...). Bản đầu tiên
// viết theo byte có bug thật, xem phần "fuzzing" trong index.md.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func main() {
	fmt.Println("Bài này tập trung vào `go test` -- xem các file _test.go và index.md.")
	fmt.Println("Add(2, 3) =", Add(2, 3))
	fmt.Println(`Reverse("xin chào") =`, Reverse("xin chào"))
}
