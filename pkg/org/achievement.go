package org

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"github.com/uptrace/uptrace/pkg/bunutil"
	"go.uber.org/zap"
)

const (
	AchievConfigureTracing    = "configure-tracing"
	AchievConfigureMetrics    = "configure-metrics"
	AchievInstallCollector    = "install-collector"
	AchievCreateMetricMonitor = "create-metric-monitor"
)

type Achievement struct {
	bun.BaseModel `bun:"achievements,alias:a"`

	ID        uint64 `json:"id" bun:",pk,autoincrement"`
	UserID    uint64 `json:"-" bun:",nullzero"`
	ProjectID uint32 `json:"-" bun:",nullzero"`
	Name      string `json:"name"`
}

func (a *Achievement) key() string {
	return fmt.Sprintf("%d-%d-%s", a.UserID, a.ProjectID, a.Name)
}

func SelectAchievements(
	ctx context.Context, pg *bun.DB, userID uint64, projectID uint32,
) ([]*Achievement, error) {
	achievements := make([]*Achievement, 0)
	if err := pg.NewSelect().
		DistinctOn("(name)").
		Model(&achievements).
		Where("user_id IS NULL OR user_id = ?", userID).
		Where("project_id IS NULL OR project_id = ?", projectID).
		OrderExpr("name, user_id DESC NULLS LAST").
		Scan(ctx); err != nil {
		return nil, err
	}
	return achievements, nil
}

var achievOnce bunutil.OnceMap

func CreateAchievementOnce(
	ctx context.Context,
	logger *otelzap.Logger,
	pg *bun.DB,
	achievement *Achievement,
) {
	achievOnce.Do(achievement.key(), func() {
		if err := UpsertAchievement(ctx, pg, achievement); err != nil {
			logger.Error("UpsertAchievement failed", zap.Error(err))
		}
	})
}

func UpsertAchievement(
	ctx context.Context, pg *bun.DB, achievement *Achievement,
) error {
	if _, err := pg.NewInsert().
		Model(achievement).
		On("CONFLICT (coalesce(user_id, 0), coalesce(project_id, 0), name) DO NOTHING").
		Exec(ctx); err != nil {
		return err
	}
	return nil
}
