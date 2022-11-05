package alerting

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"golang.org/x/exp/maps"
)

type Alert struct {
	ID          uint64                      `json:"id,string"`
	ProjectID   uint32                      `json:"projectId"`
	State       AlertState                  `json:"state,omitzero"`
	Attrs       upql.Attrs                  `json:"attrs,omitzero"`
	Annotations map[string]string           `json:"-"`
	Metrics     map[string]*upql.Timeseries `json:"-"`

	LastSeenAt time.Time `json:"lastSeenAt"`
	FiredAt    time.Time `json:"firedAt"`
	ResolvedAt time.Time `json:"resolvedAt"`
	LastSentAt time.Time `json:"lastSentAt"`
}

type alertKey struct {
	ID        uint64
	ProjectID uint32
}

func (a *Alert) key() alertKey {
	return alertKey{
		ID:        a.ID,
		ProjectID: a.ProjectID,
	}
}

type AlertState int

const (
	StateActive AlertState = iota
	StateFiring
)

func (s AlertState) String() string {
	switch s {
	case StateActive:
		return "active"
	case StateFiring:
		return "firing"
	}
	panic(fmt.Errorf("unknown alert state: %d", s))
}

type Rule struct {
	conf *RuleConfig

	alertMap map[alertKey]*Alert
}

type RuleConfig struct {
	Name        string
	Projects    []uint32
	Metrics     []upql.Metric
	Query       string
	For         time.Duration
	Labels      map[string]string
	Annotations map[string]string
}

func (r *RuleConfig) ID() int64 {
	return int64(xxhash.Sum64String(r.Name))
}

func NewRule(conf *RuleConfig, alerts []Alert) *Rule {
	r := &Rule{
		conf:     conf,
		alertMap: make(map[alertKey]*Alert, len(alerts)),
	}
	for i := range alerts {
		alert := &alerts[i]
		r.alertMap[alert.key()] = alert
	}
	return r
}

func (r *Rule) Config() *RuleConfig {
	return r.conf
}

func (r *Rule) Alerts() []Alert {
	alerts := make([]Alert, 0, len(r.alertMap))
	for _, alert := range r.alertMap {
		alerts = append(alerts, *alert)
	}
	return alerts
}

func (r *Rule) Eval(ctx context.Context, engine Engine, tm time.Time) ([]Alert, error) {
	dur := r.conf.For + time.Minute // for delta func
	timeseries, vars, err := engine.Eval(
		ctx,
		r.conf.Projects,
		r.conf.Metrics,
		r.conf.Query,
		tm.Add(-dur),
		tm,
	)
	if err != nil {
		return nil, err
	}

	unused := maps.Clone(r.alertMap)
	var alerts []Alert

	var buf []byte
	for i := range timeseries {
		ts := &timeseries[i]

		buf = ts.Attrs.Bytes(buf[:0])
		hash := xxhash.Sum64(buf)
		key := alertKey{
			ID:        hash,
			ProjectID: ts.ProjectID,
		}

		alert, ok := r.alertMap[key]
		if ok {
			delete(unused, key)
		} else {
			alert = &Alert{
				ID:        hash,
				ProjectID: ts.ProjectID,
				State:     StateActive,
				Attrs:     ts.Attrs,
			}
			r.alertMap[key] = alert
		}

		alert.Annotations = ts.Annotations
		alert.Metrics = make(map[string]*upql.Timeseries)
		alert.LastSeenAt = tm

		if r.checkTimeseries(ts, alert, tm) {
			for metricName, timeseries := range vars {
				metricName := strings.TrimPrefix(metricName, "_")

				for i := range timeseries {
					ts2 := &timeseries[i]

					buf = ts2.Attrs.Bytes(buf[:0])
					if xxhash.Sum64(buf) == hash {
						alert.Metrics[metricName] = ts2
					}
				}
			}

			alerts = append(alerts, *alert)
		}
	}

	if len(unused) == 0 {
		return alerts, nil
	}

	// TODO: fix checking disappeared timeseries
	for key, alert := range unused {
		if time.Since(alert.LastSeenAt) > 24*time.Hour {
			delete(r.alertMap, key)
			continue
		}
	}

	return alerts, nil
}

func (r *Rule) checkTimeseries(ts *upql.Timeseries, alert *Alert, tm time.Time) bool {
	var dur time.Duration

	for i := len(ts.Value) - 1; i >= 0; i-- {
		if ts.Value[i] == 0 {
			break
		}
		dur += time.Minute
	}

	switch {
	case dur == 0:
		// TODO: should we keep sending this alert for some time?
		if alert.State != StateActive {
			alert.State = StateActive
			alert.ResolvedAt = tm
			return true
		}
	case dur >= r.conf.For:
		alert.State = StateFiring
		alert.FiredAt = tm
		alert.ResolvedAt = time.Time{}
		return true
	}

	return false
}
