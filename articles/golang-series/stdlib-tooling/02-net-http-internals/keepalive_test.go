package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkWithKeepAlive(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer srv.Close()

	client := &http.Client{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(srv.URL)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

func BenchmarkWithoutKeepAlive(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))
	defer srv.Close()

	client := &http.Client{
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(srv.URL)
		if err != nil {
			b.Fatal(err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}
