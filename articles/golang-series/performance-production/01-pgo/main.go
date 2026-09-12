package main

import "fmt"

// Shape/Rectangle/Circle/Triangle mô phỏng tình huống PGO thường được nhắc tới
// nhiều nhất: gọi qua interface trong vòng lặp nóng (hot loop), nơi 1 concrete
// type áp đảo về tần suất (ở đây Rectangle chiếm 95%, Circle 5%).

type Shape interface {
	Area() float64
}

type Rectangle struct{ W, H float64 }

func (r Rectangle) Area() float64 { return r.W * r.H }

type Circle struct{ R float64 }

func (c Circle) Area() float64 { return 3.14159265 * c.R * c.R }

type Triangle struct{ B, H float64 }

func (t Triangle) Area() float64 { return 0.5 * t.B * t.H }

func sumAreas(shapes []Shape) float64 {
	total := 0.0
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}

func makeShapes(n int) []Shape {
	shapes := make([]Shape, n)
	for i := range shapes {
		if i%20 == 0 {
			shapes[i] = Circle{R: float64(i % 10)} // 5%
		} else {
			shapes[i] = Rectangle{W: float64(i % 10), H: float64(i % 7)} // 95%
		}
	}
	return shapes
}

func work(shapes []Shape) float64 {
	total := 0.0
	for iter := 0; iter < 2000; iter++ {
		total += sumAreas(shapes)
	}
	return total
}

func main() {
	shapes := makeShapes(10000)
	fmt.Println(work(shapes))
}
