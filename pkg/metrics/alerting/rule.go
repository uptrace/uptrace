package alerting

import (
	"context"
	"fmt"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/uptrace/uptrace/pkg/metrics/upql"
	"golang.org/x/exp/maps"
)

type Alert struct {
	ID    uint64     `json:"id,string"`
	State AlertState `json:"state,omitzero"`
	Attrs upql.Attrs `json:"attrs,omitzero"`

	LastSeenAt time.Time `json:"lastSeenAt"`
	FiredAt    time.Time `json:"firedAt"`
	ResolvedAt time.Time `json:"resolvedAt"`
	LastSentAt time.Time `json:"lastSentAt"`
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

	alertMap map[uint64]*Alert
}

type RuleConfig struct {
	Name        string
	Metrics     []upql.Metric
	Query       string
	For         time.Duration
	Labels      map[string]string
	Annotations map[string]string
}

func (r *RuleConfig) ID() int64 {
	return int64(xxhash.Sum64String(r.Name + "-" + r.Query))
}

func NewRule(conf *RuleConfig, alerts []Alert) *Rule {
	r := &Rule{
		conf:     conf,
		alertMap: make(map[uint64]*Alert, len(alerts)),
	}
	for i := range alerts {
		alert := &alerts[i]
		r.alertMap[alert.ID] = alert
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
	timeseries, err := engine.Eval(ctx, r.conf.Metrics, r.conf.Query, tm.Add(-r.conf.For), tm)
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

		alert, ok := r.alertMap[hash]
		if ok {
			alert.LastSeenAt = tm
			delete(unused, alert.ID)
		} else {
			alert = &Alert{
				ID:         hash,
				State:      StateActive,
				Attrs:      ts.Attrs,
				LastSeenAt: tm,
			}
			r.alertMap[hash] = alert
		}

		if r.checkTimeseries(ts, alert, tm) {
			alerts = append(alerts, *alert)
		}
	}

	if len(unused) == 0 {
		return alerts, nil
	}

	empty := new(upql.Timeseries)
	for id, alert := range unused {
		if time.Since(alert.LastSeenAt) > 24*time.Hour {
			delete(r.alertMap, id)
			continue
		}

		if r.checkTimeseries(empty, alert, tm) {
			alerts = append(alerts, *alert)
		}
	}

	return alerts, nil
}

func (r *Rule) checkTimeseries(ts *upql.Timeseries, alert *Alert, tm time.Time) bool {
	var dur time.Duration

	for i := len(ts.Value) - 1; i >= 0; i-- {
		if ts.Value[i] <= 0 {
			break
		}
		dur += time.Minute
	}

	switch {
	case dur == 0:
		if alert.State != StateActive {
			alert.State = StateActive
			alert.ResolvedAt = tm
			return true
		}
	case dur >= r.conf.For:
		if alert.State != StateFiring {
			alert.State = StateFiring
			alert.FiredAt = tm
			alert.ResolvedAt = time.Time{}
			return true
		}
	}

	return false
}
