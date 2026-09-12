package main

import "testing"

func TestAdd(t *testing.T) {
	cases := []struct {
		name string
		a, b int
		want int
	}{
		{"hai số dương", 2, 3, 5},
		{"số âm", -1, -1, -2},
		{"cộng 0", 5, 0, 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Add(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Add(%d, %d) = %d, muốn %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
