package main

import "fmt"

// ============================================================================
// Demo 1: Type parameter cơ bản -- Map[T, U any]
// ============================================================================

func Map[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

func demoBasicGeneric() {
	fmt.Println("--- Demo 1: Type parameter cơ bản ---")
	nums := []int{1, 2, 3, 4}
	doubled := Map(nums, func(n int) int { return n * 2 })
	fmt.Println("doubled:", doubled)

	strs := Map(nums, func(n int) string { return fmt.Sprintf("n=%d", n) })
	fmt.Println("strs:", strs)
}

// ============================================================================
// Demo 2: Constraints -- Max hoạt động cho nhiều kiểu dữ liệu
// (Xem index.md để thấy lỗi compile thật khi gọi Max với kiểu không thoả constraint)
// ============================================================================

type Ordered interface {
	~int | ~int64 | ~float64 | ~string
}

func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func demoConstraints() {
	fmt.Println("\n--- Demo 2: Constraints ---")
	fmt.Println("Max(3, 7):", Max(3, 7))
	fmt.Println("Max(3.5, 2.1):", Max(3.5, 2.1))
	fmt.Println(`Max("banana", "apple"):`, Max("banana", "apple"))
}

// ============================================================================
// Demo 3: So sánh code lặp lại vs generic
// ============================================================================

func SumInt(s []int) int {
	var total int
	for _, v := range s {
		total += v
	}
	return total
}

func SumFloat64(s []float64) float64 {
	var total float64
	for _, v := range s {
		total += v
	}
	return total
}

type Number interface {
	~int | ~int64 | ~float64
}

func Sum[T Number](s []T) T {
	var total T
	for _, v := range s {
		total += v
	}
	return total
}

func demoDuplicationVsGeneric() {
	fmt.Println("\n--- Demo 3: Code lặp lại vs generic ---")
	ints := []int{1, 2, 3}
	floats := []float64{1.5, 2.5}

	fmt.Println("SumInt:", SumInt(ints), "  Sum generic:", Sum(ints))
	fmt.Println("SumFloat64:", SumFloat64(floats), "  Sum generic:", Sum(floats))
}

// ============================================================================
// Demo 4: Khi nào generics KHÔNG cần thiết -- interface đã đủ, thậm chí linh hoạt hơn
// ============================================================================

type Stringer interface {
	String() string
}

type User struct{ Name string }

func (u User) String() string { return "User:" + u.Name }

type Product struct{ Title string }

func (p Product) String() string { return "Product:" + p.Title }

// ĐÚNG: chỉ cần gọi method chung -- interface là đủ
func PrintAll(items []Stringer) {
	for _, item := range items {
		fmt.Println(item.String())
	}
}

// KHÔNG CẦN THIẾT: generics ở đây không giải quyết được gì thêm, và còn KÉM linh
// hoạt hơn -- chỉ nhận được slice CÙNG 1 kiểu cụ thể T, không trộn được nhiều kiểu.
func PrintAllGeneric[T Stringer](items []T) {
	for _, item := range items {
		fmt.Println(item.String())
	}
}

func demoWhenGenericsNotNeeded() {
	fmt.Println("\n--- Demo 4: Khi nào generics KHÔNG cần ---")

	mixed := []Stringer{User{Name: "An"}, Product{Title: "Sách"}}
	PrintAll(mixed) // interface: nhận được slice TRỘN nhiều kiểu

	users := []User{{Name: "Bình"}}
	PrintAllGeneric(users) // generic: chỉ nhận được slice CÙNG 1 kiểu
}

func main() {
	demoBasicGeneric()
	demoConstraints()
	demoDuplicationVsGeneric()
	demoWhenGenericsNotNeeded()
}
