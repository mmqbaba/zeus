package log

import "testing"

func BenchmarkLogInfo(b *testing.B) {
	b.Run("report", func(b *testing.B) {
		defer logger.Sync()
		for i := 0; i < b.N; i++ {
			print()
		}
	})
}
