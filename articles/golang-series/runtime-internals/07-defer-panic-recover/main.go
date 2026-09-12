package main

import "fmt"

// ============================================================================
// Demo 1: recover() gọi trực tiếp trong defer -- cách dùng đúng, an toàn
// (bản gọi GIÁN TIẾP qua 1 hàm khác -- không bắt được panic -- nằm ở file riêng)
// ============================================================================

func withDirectRecover() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("defer trực tiếp bắt được:", r)
		}
	}()
	panic("panic từ withDirectRecover")
}

func demoDirectRecover() {
	fmt.Println("--- Demo 1: recover() gọi trực tiếp trong defer ---")
	withDirectRecover()
	fmt.Println("Chương trình tiếp tục chạy bình thường sau withDirectRecover")
}

// ============================================================================
// Demo 2: panic unwind -- chạy ngược defer chain qua nhiều tầng gọi hàm
// ============================================================================

func level3() {
	defer fmt.Println("level3: defer chạy (unwind đi qua đây)")
	panic("panic từ level3")
}

func level2() {
	defer fmt.Println("level2: defer chạy (unwind đi qua đây)")
	level3()
	fmt.Println("level2: dòng này KHÔNG chạy vì panic đã xảy ra ở level3")
}

func level1() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("level1: recover() bắt được:", r)
		}
	}()
	defer fmt.Println("level1: defer chạy (unwind đi qua đây)")
	level2()
	fmt.Println("level1: dòng này cũng KHÔNG chạy")
}

func demoPanicUnwind() {
	fmt.Println("\n--- Demo 2: panic unwind qua nhiều tầng gọi hàm ---")
	level1()
	fmt.Println("Chương trình tiếp tục sau khi level1 xử lý xong panic")
}

func main() {
	demoDirectRecover()
	demoPanicUnwind()
}
