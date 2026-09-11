package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// ============================================================================
// Demo 1: Implicit satisfaction -- UpperWriter tự động thoả mãn io.Writer
// mà không hề khai báo "implements", không cần biết io.Writer tồn tại lúc định nghĩa.
// ============================================================================

type UpperWriter struct {
	Out *strings.Builder
}

func (w *UpperWriter) Write(p []byte) (int, error) {
	upper := strings.ToUpper(string(p))
	return w.Out.WriteString(upper)
}

func demoImplicitSatisfaction() {
	fmt.Println("--- Demo 1: Implicit satisfaction ---")
	var sb strings.Builder
	uw := &UpperWriter{Out: &sb}

	var _ io.Writer = uw // kiểm tra tường minh tại compile-time
	fmt.Fprintf(uw, "hello from go\n")

	fmt.Println("Kết quả:", sb.String())
}

// ============================================================================
// Demo 2: Nil interface trap -- bug kinh điển nhất Go
// ============================================================================

type CustomError struct {
	msg string
}

func (e *CustomError) Error() string {
	return e.msg
}

func mayFail(shouldFail bool) *CustomError {
	if shouldFail {
		return &CustomError{msg: "đã lỗi"}
	}
	return nil // pointer nil, kiểu *CustomError
}

// Phiên bản SAI: trả trực tiếp biến pointer đã typed, ép kiểu ngầm định sang error.
func doSomethingBuggy() error {
	var err *CustomError = mayFail(false)
	return err // interface (type=*CustomError, value=nil) -- KHÔNG bằng nil
}

// Phiên bản ĐÚNG: kiểm tra tường minh trước khi return.
func doSomethingFixed() error {
	customErr := mayFail(false)
	if customErr != nil {
		return customErr
	}
	return nil // nil interface THẬT (type=nil, value=nil)
}

func demoNilInterfaceTrap() {
	fmt.Println("\n--- Demo 2: Nil interface trap ---")

	buggyErr := doSomethingBuggy()
	fmt.Printf("Phiên bản SAI:  err == nil? %-5v  (kiểu %T)\n", buggyErr == nil, buggyErr)

	fixedErr := doSomethingFixed()
	fmt.Printf("Phiên bản ĐÚNG: err == nil? %-5v  (kiểu %T)\n", fixedErr == nil, fixedErr)
}

// ============================================================================
// Demo 3: Interface composition -- io.Reader + io.Writer = io.ReadWriter
// ============================================================================

func useReadWriter(rw io.ReadWriter) string {
	rw.Write([]byte("ghi qua interface ghép"))
	buf := make([]byte, 64)
	n, _ := rw.Read(buf)
	return string(buf[:n])
}

func demoInterfaceComposition() {
	fmt.Println("\n--- Demo 3: Interface composition ---")
	// bytes.Buffer tự động thoả mãn CẢ io.Reader lẫn io.Writer, nên cũng tự động
	// thoả mãn io.ReadWriter -- không cần khai báo gì thêm.
	var buf bytes.Buffer
	result := useReadWriter(&buf)
	fmt.Println("Đọc lại:", result)
}

// ============================================================================
// Demo 4: Accept interfaces, return structs
// ============================================================================

type Logger struct {
	prefix string
	out    io.Writer
}

// Trả về concrete struct (*Logger), không phải interface -- người gọi có đủ
// thông tin, dùng được field/method riêng của Logger nếu cần.
func NewLogger(prefix string, out io.Writer) *Logger {
	return &Logger{prefix: prefix, out: out}
}

func (l *Logger) Log(msg string) {
	fmt.Fprintf(l.out, "[%s] %s\n", l.prefix, msg)
}

// Chỉ cần khả năng "ghi" -- nhận io.Writer, không ép buộc kiểu cụ thể nào.
func writeReport(w io.Writer, content string) {
	fmt.Fprintln(w, "=== BÁO CÁO ===")
	fmt.Fprintln(w, content)
}

func demoAcceptInterfaceReturnStruct() {
	fmt.Println("\n--- Demo 4: Accept interfaces, return structs ---")
	logger := NewLogger("APP", os.Stdout)
	logger.Log("khởi động thành công")

	writeReport(os.Stdout, "mọi thứ ổn")
}

func main() {
	demoImplicitSatisfaction()
	demoNilInterfaceTrap()
	demoInterfaceComposition()
	demoAcceptInterfaceReturnStruct()
}
