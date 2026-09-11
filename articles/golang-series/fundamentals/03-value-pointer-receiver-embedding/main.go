package main

import "fmt"

// ============================================================================
// Demo 1: Value vs Pointer receiver -- bug thật từ nhầm lẫn
// ============================================================================

type CounterWrong struct{ n int }

func (c CounterWrong) Inc() { c.n++ } // value receiver: sửa bản copy, vô tác dụng

type CounterRight struct{ n int }

func (c *CounterRight) Inc() { c.n++ } // pointer receiver: sửa struct gốc

func demoValueVsPointerReceiver() {
	fmt.Println("--- Demo 1: Value vs Pointer receiver ---")

	wrong := CounterWrong{}
	for i := 0; i < 5; i++ {
		wrong.Inc()
	}
	fmt.Println("wrong.n =", wrong.n)

	right := CounterRight{}
	for i := 0; i < 5; i++ {
		right.Inc()
	}
	fmt.Println("right.n =", right.n)
}

// ============================================================================
// Demo 2: Method set -- pointer thoả mãn interface, value thì không
// (Xem index.md để thấy lỗi compile thật khi cố truyền value thay vì pointer)
// ============================================================================

type Incrementer interface {
	Inc()
}

func useIncrementer(inc Incrementer) {
	inc.Inc()
}

func demoMethodSet() {
	fmt.Println("\n--- Demo 2: Method set ---")
	c := CounterRight{}
	useIncrementer(&c) // phải truyền &c -- CounterRight (value) không thoả mãn Incrementer
	fmt.Println("Sau khi truyền &c:", c.n)
}

// ============================================================================
// Demo 3: Struct embedding -- field/method promotion
// ============================================================================

type Engine struct{ Power int }

func (e Engine) Start() string { return fmt.Sprintf("engine started, power=%d", e.Power) }

type Car struct {
	Engine // embedding -- không có tên field, chỉ có tên type
	Brand  string
}

func demoStructEmbedding() {
	fmt.Println("\n--- Demo 3: Struct embedding ---")
	car := Car{Engine: Engine{Power: 150}, Brand: "Toyota"}

	fmt.Println(car.Start())      // promoted method
	fmt.Println(car.Power)        // promoted field
	fmt.Println(car.Engine.Power) // vẫn truy cập được tường minh qua tên type
}

// ============================================================================
// Demo 4: Ambiguous selector khi 2 embedded struct trùng tên method
// (Xem index.md để thấy lỗi compile thật; ở đây chỉ demo cách gọi tường minh để fix)
// ============================================================================

type Alarm struct{}

func (a Alarm) Start() string { return "alarm started" }

type CarWithAlarm struct {
	Engine
	Alarm
}

func demoAmbiguousSelectorFix() {
	fmt.Println("\n--- Demo 4: Fix ambiguous selector bằng gọi tường minh ---")
	car := CarWithAlarm{Engine: Engine{Power: 150}}
	fmt.Println(car.Engine.Start()) // gọi tường minh, hết mơ hồ
	fmt.Println(car.Alarm.Start())
}

func main() {
	demoValueVsPointerReceiver()
	demoMethodSet()
	demoStructEmbedding()
	demoAmbiguousSelectorFix()
}
