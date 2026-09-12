package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
)

// ============================================================================
// Demo 1: Struct tag -- omitempty, "-", đổi tên field
// ============================================================================

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Age      int    `json:"age,omitempty"`
	Password string `json:"-"`
}

func demoStructTags() {
	fmt.Println("--- Demo 1: Struct tag ---")
	u1 := User{Name: "An", Email: "an@example.com", Age: 30, Password: "secret123"}
	b1, _ := json.Marshal(u1)
	fmt.Println("Đầy đủ field:", string(b1))

	u2 := User{Name: "Binh"} // Email và Age là zero value
	b2, _ := json.Marshal(u2)
	fmt.Println("Email/Age rỗng (omitempty):", string(b2))
}

// ============================================================================
// Demo 2: MarshalJSON/UnmarshalJSON tuỳ biến -- enum dạng string
// ============================================================================

type Status int

const (
	StatusPending Status = iota
	StatusActive
	StatusDone
)

func (s Status) String() string {
	return [...]string{"pending", "active", "done"}[s]
}

func (s Status) MarshalJSON() ([]byte, error) {
	return []byte(`"` + s.String() + `"`), nil
}

func (s *Status) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	switch str {
	case "pending":
		*s = StatusPending
	case "active":
		*s = StatusActive
	case "done":
		*s = StatusDone
	default:
		return fmt.Errorf("status không hợp lệ: %s", str)
	}
	return nil
}

type Task struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
}

func demoCustomMarshal() {
	fmt.Println("\n--- Demo 2: MarshalJSON/UnmarshalJSON tuỳ biến ---")
	t := Task{Name: "Viết bài", Status: StatusActive}
	b, _ := json.Marshal(t)
	fmt.Println("Marshal (Status thành chuỗi thay vì số):", string(b))

	var t2 Task
	json.Unmarshal([]byte(`{"name":"Đọc bài","status":"done"}`), &t2)
	fmt.Printf("Unmarshal ngược lại: %+v (Status int thật = %d)\n", t2, t2.Status)
}

// ============================================================================
// Demo 3: Streaming Decoder vs Unmarshal load hết vào bộ nhớ
// ============================================================================

type Record struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func buildJSONArray(n int) []byte {
	records := make([]Record, n)
	for i := range records {
		records[i] = Record{ID: i, Name: fmt.Sprintf("item-%d", i), Value: "some payload data here to pad size a bit"}
	}
	b, _ := json.Marshal(records)
	return b
}

func demoStreaming() {
	fmt.Println("\n--- Demo 3: Streaming Decoder vs Unmarshal load hết ---")
	data := buildJSONArray(100000)
	fmt.Println("Kích thước JSON:", len(data), "byte")

	runtime.GC()
	var before1, after1 runtime.MemStats
	runtime.ReadMemStats(&before1)
	var records1 []Record
	json.Unmarshal(data, &records1)
	runtime.ReadMemStats(&after1)
	fmt.Println("Unmarshal: HeapAlloc tăng", after1.HeapAlloc-before1.HeapAlloc, "bytes | số record:", len(records1))

	runtime.GC()
	var before2, after2 runtime.MemStats
	runtime.ReadMemStats(&before2)
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.Token() // '['
	count := 0
	for dec.More() {
		var r Record
		dec.Decode(&r)
		count++
	}
	runtime.ReadMemStats(&after2)
	fmt.Println("Decoder streaming: HeapAlloc tăng", after2.HeapAlloc-before2.HeapAlloc, "bytes | số record xử lý:", count)
}

func main() {
	demoStructTags()
	demoCustomMarshal()
	demoStreaming()
}
