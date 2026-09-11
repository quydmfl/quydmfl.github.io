package main

import "fmt"

// ============================================================================
// Demo 1: Slice là view (pointer + len + cap), không copy dữ liệu
// ============================================================================

func demoSliceIsView() {
	fmt.Println("--- Demo 1: Slice là view, không copy dữ liệu ---")
	a := []int{1, 2, 3, 4, 5}
	b := a[1:3] // slicing tạo view, chia sẻ chung underlying array với a

	fmt.Printf("a: %v, len=%d, cap=%d\n", a, len(a), cap(a))
	fmt.Printf("b: %v, len=%d, cap=%d\n", b, len(b), cap(b))
	fmt.Println("Địa chỉ b[0] có trùng địa chỉ a[1]?", &b[0] == &a[1])

	b[0] = 999
	fmt.Println("\nSau khi sửa b[0] = 999:")
	fmt.Println("a:", a, " <- bị ảnh hưởng dù chỉ sửa qua b")
	fmt.Println("b:", b)
}

// ============================================================================
// Demo 2: Bug aliasing thật qua append -- khi còn dư capacity
// ============================================================================

func processA(data []int) []int {
	// Tưởng đang làm việc độc lập, nhưng nếu data còn dư capacity, append sẽ
	// ghi đè ngay lên underlying array gốc mà nơi gọi vẫn đang giữ tham chiếu.
	return append(data, 999)
}

func demoAppendAliasingBug() {
	fmt.Println("\n--- Demo 2: Bug aliasing qua append ---")
	original := make([]int, 3, 5) // len=3, cap=5 -- cố tình còn dư capacity
	original[0], original[1], original[2] = 1, 2, 3

	sub := original[:2] // sub chia sẻ underlying array, cap(sub) vẫn là 5
	fmt.Println("Trước:", original, " cap(sub) =", cap(sub))

	result := processA(sub)

	fmt.Println("result:  ", result)
	fmt.Println("original:", original, " <- phần tử index 2 bị ghi đè âm thầm!")
}

// ============================================================================
// Demo 3: Fix bằng copy tường minh
// ============================================================================

func processASafe(data []int) []int {
	safe := make([]int, len(data), len(data)) // cap == len, không dư chỗ để chia sẻ
	copy(safe, data)
	return append(safe, 999)
}

func demoAppendFixed() {
	fmt.Println("\n--- Demo 3: Fix bằng copy tường minh ---")
	original := make([]int, 3, 5)
	original[0], original[1], original[2] = 1, 2, 3
	sub := original[:2]

	result := processASafe(sub)
	fmt.Println("result:  ", result)
	fmt.Println("original:", original, " <- KHÔNG bị ảnh hưởng")
}

// ============================================================================
// Demo 4: Cơ chế grow capacity -- đo thật
// ============================================================================

func demoCapacityGrowth() {
	fmt.Println("\n--- Demo 4: Cơ chế grow capacity ---")
	s := make([]int, 0)
	prevCap := cap(s)
	for i := 0; i < 20; i++ {
		s = append(s, i)
		if cap(s) != prevCap {
			fmt.Printf("len=%2d  cap=%2d  (tăng từ %d)\n", len(s), cap(s), prevCap)
			prevCap = cap(s)
		}
	}
}

func main() {
	demoSliceIsView()
	demoAppendAliasingBug()
	demoAppendFixed()
	demoCapacityGrowth()
}
