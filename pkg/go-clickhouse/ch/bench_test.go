package ch_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkNumbers(b *testing.B) {
	ctx := context.Background()
	db := chDB()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rows, err := db.QueryContext(ctx, "SELECT number FROM system.numbers_mt LIMIT 1000000")
		if err != nil {
			b.Fatal(err)
		}

		var count int
		for rows.Next() {
			count++
		}
		require.Equal(b, 1000000, count)
	}
}
