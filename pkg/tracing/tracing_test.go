package tracing

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/uptrace/pkg/testbed"
)

func TestPing(t *testing.T) {
	ctx, app := testbed.StartApp(t)
	defer app.Stop()

	err := app.CH().Ping(ctx)
	require.NoError(t, err)
}
