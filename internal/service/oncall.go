package service

import (
	"time"

	"github.com/irvanmhndra/siaga-api/internal/model"
)

// ResolveOnCall computes the user ID on-call for a schedule at time t.
//
// Precedence: an active override wins outright. Otherwise the highest-position
// layer that yields someone wins (upper layers cover lower ones, PagerDuty
// style). Within a layer, members rotate every rotation_length_days starting
// from the layer's start_at. Returns "" when nobody is on-call.
func ResolveOnCall(s *model.Schedule, t time.Time) string {
	for _, o := range s.Overrides {
		if !t.Before(o.StartAt) && t.Before(o.EndAt) {
			return o.UserID
		}
	}

	oncall := ""
	for _, l := range s.Layers {
		if u := rotationUser(l, t); u != "" {
			oncall = u
		}
	}
	return oncall
}

func rotationUser(l model.ScheduleLayer, t time.Time) string {
	n := len(l.UserIDs)
	if n == 0 {
		return ""
	}
	days := l.RotationLengthDays
	if days <= 0 {
		days = 7
	}
	period := time.Duration(days) * 24 * time.Hour
	elapsed := max(t.Sub(l.StartAt), 0)
	idx := int(elapsed/period) % n
	return l.UserIDs[idx]
}
