package main

import (
	"testing"
	"unicode/utf8"
)

func FuzzReverse(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			return // chỉ quan tâm input UTF-8 hợp lệ
		}
		rev := Reverse(s)
		if !utf8.ValidString(rev) {
			t.Errorf("Reverse(%q) = %q -- không còn là UTF-8 hợp lệ", s, rev)
		}
	})
}
