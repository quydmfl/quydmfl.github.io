package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// ============================================================================
// Demo 1: Sentinel error -- so sánh bằng == hoặc errors.Is
// ============================================================================

var ErrNotFound = errors.New("not found")

func find(id int) error {
	if id != 1 {
		return ErrNotFound
	}
	return nil
}

func demoSentinelError() {
	fmt.Println("--- Demo 1: Sentinel error ---")
	err := find(42)
	fmt.Println("So sánh trực tiếp ==:", err == ErrNotFound)
	fmt.Println("errors.Is:", errors.Is(err, ErrNotFound))
	fmt.Println("Giới hạn: không biết đang tìm ID nào, chỉ biết 'not found' chung chung")
}

// ============================================================================
// Demo 2: Typed error + error wrapping xuyên nhiều lớp gọi hàm
// ============================================================================

type NotFoundError struct {
	Resource string
	ID       int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s #%d: %v", e.Resource, e.ID, ErrNotFound)
}

func (e *NotFoundError) Unwrap() error { return ErrNotFound }

func queryDB(id int) error {
	return &NotFoundError{Resource: "user", ID: id}
}

func getUser(id int) error {
	if err := queryDB(id); err != nil {
		return fmt.Errorf("getUser: %w", err) // %w giữ nguyên chuỗi Unwrap
	}
	return nil
}

func handleRequest(id int) error {
	if err := getUser(id); err != nil {
		return fmt.Errorf("handleRequest: %w", err)
	}
	return nil
}

// Phiên bản dùng %v thay vì %w -- để đối chiếu
func getUserNoWrap(id int) error {
	if err := queryDB(id); err != nil {
		return fmt.Errorf("getUser: %v", err)
	}
	return nil
}

func demoErrorWrapping() {
	fmt.Println("\n--- Demo 2: Error wrapping qua 3 lớp gọi hàm (%w) ---")
	err := handleRequest(42)
	fmt.Println("Thông báo lỗi đầy đủ:", err)
	fmt.Println("errors.Is(err, ErrNotFound):", errors.Is(err, ErrNotFound))

	var nfErr *NotFoundError
	ok := errors.As(err, &nfErr)
	fmt.Println("errors.As lấy được NotFoundError:", ok)
	if ok {
		fmt.Println("  Resource:", nfErr.Resource, ", ID:", nfErr.ID)
	}

	fmt.Println("\n--- Đối chiếu: dùng verb v thay vì w khi wrap lỗi (fmt.Errorf) ---")
	errNoWrap := getUserNoWrap(42)
	fmt.Println("Thông báo lỗi (trông giống hệt):", errNoWrap)
	fmt.Println("errors.Is(errNoWrap, ErrNotFound):", errors.Is(errNoWrap, ErrNotFound))
	var nfErr2 *NotFoundError
	fmt.Println("errors.As lấy được NotFoundError:", errors.As(errNoWrap, &nfErr2))
}

// ============================================================================
// Demo 3: panic/recover ở biên -- HTTP middleware
// ============================================================================

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Printf("  [middleware] chặn được panic: %v -- server KHÔNG sập\n", rec)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func buggyHandler(w http.ResponseWriter, r *http.Request) {
	var m map[string]int // nil map
	m["x"] = 1           // panic: assignment to entry in nil map
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func demoPanicRecover() {
	fmt.Println("\n--- Demo 3: panic/recover ở biên (HTTP middleware) ---")
	mux := http.NewServeMux()
	mux.HandleFunc("/buggy", buggyHandler)
	mux.HandleFunc("/ok", okHandler)
	handler := recoverMiddleware(mux)

	fmt.Println("Request tới /buggy (handler panic):")
	req1 := httptest.NewRequest("GET", "/buggy", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	fmt.Println("  status code trả về:", rec1.Code)

	fmt.Println("Request tới /ok NGAY SAU ĐÓ (server vẫn sống):")
	req2 := httptest.NewRequest("GET", "/ok", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	fmt.Println("  status code:", rec2.Code, ", body:", rec2.Body.String())
}

func main() {
	demoSentinelError()
	demoErrorWrapping()
	demoPanicRecover()
}
