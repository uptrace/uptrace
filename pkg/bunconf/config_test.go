package bunconf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadConfigLogsSchema(t *testing.T) {
	confPath := filepath.Join(t.TempDir(), "uptrace.yml")
	confYAML := []byte(`
ch_schema:
  spans:
    ttl_delete: 30 DAY
    storage_policy: spans
  logs:
    ttl_delete: 3 DAY
    storage_policy: logs
`)
	if err := os.WriteFile(confPath, confYAML, 0o600); err != nil {
		t.Fatal(err)
	}

	conf, err := ReadConfig(confPath, "test")
	if err != nil {
		t.Fatal(err)
	}

	if got, want := conf.CHSchema.Logs.TTLDelete, "3 DAY"; got != want {
		t.Fatalf("Logs.TTLDelete = %q, want %q", got, want)
	}
	if got, want := conf.CHSchema.Logs.StoragePolicy, "logs"; got != want {
		t.Fatalf("Logs.StoragePolicy = %q, want %q", got, want)
	}
}

func TestDefaultConfigLogsSchema(t *testing.T) {
	conf := defaultConfig()
	conf.CHSchema.Spans.TTLDelete = "14 DAY"
	conf.CHSchema.Spans.StoragePolicy = "spans"

	if got, want := conf.CHSchema.Logs.TTLDelete, "30 DAY"; got != want {
		t.Fatalf("Logs.TTLDelete = %q, want %q", got, want)
	}
	if got, want := conf.CHSchema.Logs.StoragePolicy, "default"; got != want {
		t.Fatalf("Logs.StoragePolicy = %q, want %q", got, want)
	}
}
