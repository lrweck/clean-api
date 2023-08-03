package rest

import "testing"

var Size = 5

func BenchmarkMapSimple(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		m := NewSafeMap[int, int](Size)
		b.ResetTimer()
		b.ReportAllocs()

		i := 0
		for pb.Next() {
			i++
			m.GetOrSet2(i%Size, i)
		}
	})

}

func BenchmarkOptimized(b *testing.B) {

	b.RunParallel(func(pb *testing.PB) {
		m := NewSafeMap[int, int](Size)
		b.ResetTimer()
		b.ReportAllocs()

		i := 0
		for pb.Next() {
			i++
			m.GetOrSet1(i%Size, i)
		}
	})

}
