package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const readN = 500000 // 500KB, đọc từng byte một

// testFilePath tạo (nếu chưa có) 1 file văn bản đủ lớn trong thư mục tạm của hệ
// điều hành để benchmark đọc -- không cần commit file dữ liệu lớn vào repo.
func testFilePath(tb testing.TB) string {
	tb.Helper()
	path := filepath.Join(os.TempDir(), "io-bufio-bench-testfile.txt")
	if info, err := os.Stat(path); err == nil && info.Size() > readN {
		return path
	}
	f, err := os.Create(path)
	if err != nil {
		tb.Fatal(err)
	}
	defer f.Close()
	line := strings.Repeat("dữ liệu mẫu cho benchmark đọc file ", 4) + "\n"
	for i := 0; i < 50000; i++ {
		if _, err := f.WriteString(line); err != nil {
			tb.Fatal(err)
		}
	}
	return path
}

func BenchmarkReadByteNoBuffer(b *testing.B) {
	path := testFilePath(b)
	for n := 0; n < b.N; n++ {
		f, err := os.Open(path)
		if err != nil {
			b.Fatal(err)
		}
		buf := make([]byte, 1)
		count := 0
		for count < readN {
			_, err := f.Read(buf)
			if err == io.EOF {
				break
			}
			count++
		}
		f.Close()
	}
}

func BenchmarkReadByteBuffered(b *testing.B) {
	path := testFilePath(b)
	for n := 0; n < b.N; n++ {
		f, err := os.Open(path)
		if err != nil {
			b.Fatal(err)
		}
		r := bufio.NewReader(f)
		count := 0
		for count < readN {
			_, err := r.ReadByte()
			if err == io.EOF {
				break
			}
			count++
		}
		f.Close()
	}
}
