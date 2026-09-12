package main

import (
	"encoding/json"
	"strconv"
	"testing"
)

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func manualMarshal(p Point) []byte {
	buf := make([]byte, 0, 32)
	buf = append(buf, `{"x":`...)
	buf = strconv.AppendInt(buf, int64(p.X), 10)
	buf = append(buf, `,"y":`...)
	buf = strconv.AppendInt(buf, int64(p.Y), 10)
	buf = append(buf, '}')
	return buf
}

func BenchmarkJSONMarshal(b *testing.B) {
	p := Point{X: 10, Y: 20}
	for i := 0; i < b.N; i++ {
		json.Marshal(p)
	}
}

func BenchmarkManualMarshal(b *testing.B) {
	p := Point{X: 10, Y: 20}
	for i := 0; i < b.N; i++ {
		manualMarshal(p)
	}
}

type Flat struct {
	A, B, C, D, E int
}

type Inner struct {
	X, Y, Z int
}

type Nested struct {
	Name    string
	Inners  []Inner
	Mapping map[string]Inner
	Sub     struct {
		Deep struct {
			Value int
		}
	}
}

func BenchmarkMarshalFlat(b *testing.B) {
	v := Flat{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		json.Marshal(v)
	}
}

func BenchmarkMarshalNested(b *testing.B) {
	v := Nested{
		Name: "test",
		Inners: []Inner{
			{1, 2, 3}, {4, 5, 6}, {7, 8, 9},
		},
		Mapping: map[string]Inner{
			"a": {1, 1, 1},
			"b": {2, 2, 2},
		},
	}
	v.Sub.Deep.Value = 42
	for i := 0; i < b.N; i++ {
		json.Marshal(v)
	}
}
