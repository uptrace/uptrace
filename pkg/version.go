package pkg

import "github.com/uptrace/uptrace/pkg/internal/version"

func Version() string {
	return version.Version
}
