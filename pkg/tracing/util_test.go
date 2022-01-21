package tracing

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatSQL(t *testing.T) {
	query := formatSQL("SELECT 1")
	require.Equal(t, "SELECT 1\n", query)

	query = formatSQL("SELECT ?")
	require.Equal(t, "", query)
}
