package main

import "testing"

func BenchmarkUnbufferedChan(b *testing.B) {
	ch := make(chan int)
	done := make(chan struct{})
	go func() {
		for v := range ch {
			_ = v
		}
		close(done)
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
	}
	close(ch)
	<-done
}

func BenchmarkBufferedChan(b *testing.B) {
	ch := make(chan int, 128)
	done := make(chan struct{})
	go func() {
		for v := range ch {
			_ = v
		}
		close(done)
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
	}
	close(ch)
	<-done
}
