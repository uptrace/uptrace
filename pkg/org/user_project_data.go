package org

import "github.com/uptrace/bun"

type UserProjectData struct {
	bun.BaseModel `bun:"user_project_data,alias:d"`

	UserID    uint64 `json:"-"`
	ProjectID uint32 `json:"-"`

	NotifyOnMetrics         bool `json:"notifyOnMetrics"`
	NotifyOnNewErrors       bool `json:"notifyOnNewErrors"`
	NotifyOnRecurringErrors bool `json:"notifyOnRecurringErrors"`
}
