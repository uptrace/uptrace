package testbed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/uptrace/pkg/bunapp"
)

func StartApp(t *testing.T) (context.Context, *bunapp.App) {
	ctx, app, err := bunapp.Start(context.Background(), "../../config/uptrace.yml", "test")
	require.NoError(t, err)
	return ctx, app
}
