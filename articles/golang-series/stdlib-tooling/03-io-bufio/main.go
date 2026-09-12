package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"
)

// ============================================================================
// Demo 1: io.Copy -- không cần biết trước kích thước dữ liệu
// ============================================================================

func demoIOCopy() {
	fmt.Println("--- Demo 1: io.Copy ---")
	var buf bytes.Buffer
	src := strings.NewReader(strings.Repeat("x", 100000))
	n, err := io.Copy(&buf, src)
	fmt.Println("Đã copy", n, "byte, lỗi:", err, "| buf.Len():", buf.Len())
}

// ============================================================================
// Demo 2: io.MultiReader -- nối nhiều Reader thành 1 luồng liên tục
// ============================================================================

func demoMultiReader() {
	fmt.Println("\n--- Demo 2: io.MultiReader ---")
	r1 := strings.NewReader("Phần 1. ")
	r2 := strings.NewReader("Phần 2. ")
	r3 := strings.NewReader("Phần 3.")
	combined := io.MultiReader(r1, r2, r3)
	data, _ := io.ReadAll(combined)
	fmt.Println("Kết quả:", string(data))
}

// ============================================================================
// Demo 3: io.TeeReader -- vừa đọc vừa ghi song song sang nơi khác (hash)
// ============================================================================

func demoTeeReader() {
	fmt.Println("\n--- Demo 3: io.TeeReader ---")
	src := strings.NewReader("dữ liệu cần vừa đọc vừa hash")
	hasher := sha256.New()
	tee := io.TeeReader(src, hasher)
	content, _ := io.ReadAll(tee)
	fmt.Println("Đọc được:", string(content))
	fmt.Printf("SHA256 tính được trong lúc đọc: %x\n", hasher.Sum(nil))
}

func main() {
	demoIOCopy()
	demoMultiReader()
	demoTeeReader()
}
