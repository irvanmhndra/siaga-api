package service

import (
	"testing"
	"time"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/lib/pq"
)

func mkSchedule(layers []model.ScheduleLayer, overrides []model.ScheduleOverride) *model.Schedule {
	return &model.Schedule{ID: "sch", OrganizationID: "org", Layers: layers, Overrides: overrides}
}

func TestResolveOnCall_EmptySchedule(t *testing.T) {
	if got := ResolveOnCall(mkSchedule(nil, nil), time.Now()); got != "" {
		t.Fatalf("empty schedule: want \"\", got %q", got)
	}
}

func TestResolveOnCall_WeeklyRotation(t *testing.T) {
	start := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC) // a Monday
	layer := model.ScheduleLayer{
		Name:               "L1",
		RotationLengthDays: 7,
		StartAt:            start,
		UserIDs:            pq.StringArray{"alice", "bob"},
	}
	s := mkSchedule([]model.ScheduleLayer{layer}, nil)

	cases := []struct {
		name string
		at   time.Time
		want string
	}{
		{"at start", start, "alice"},
		{"mid first week", start.Add(3 * 24 * time.Hour), "alice"},
		{"second week", start.Add(7 * 24 * time.Hour), "bob"},
		{"third week wraps", start.Add(14 * 24 * time.Hour), "alice"},
		{"before start clamps to first", start.Add(-48 * time.Hour), "alice"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ResolveOnCall(s, c.at); got != c.want {
				t.Fatalf("want %q, got %q", c.want, got)
			}
		})
	}
}

func TestResolveOnCall_OverrideWins(t *testing.T) {
	start := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	layer := model.ScheduleLayer{RotationLengthDays: 7, StartAt: start, UserIDs: pq.StringArray{"alice"}}
	ovStart := start.Add(2 * 24 * time.Hour)
	ovEnd := start.Add(4 * 24 * time.Hour)
	ov := model.ScheduleOverride{UserID: "carol", StartAt: ovStart, EndAt: ovEnd}
	s := mkSchedule([]model.ScheduleLayer{layer}, []model.ScheduleOverride{ov})

	if got := ResolveOnCall(s, ovStart.Add(time.Hour)); got != "carol" {
		t.Fatalf("inside override: want carol, got %q", got)
	}
	if got := ResolveOnCall(s, ovEnd.Add(time.Hour)); got != "alice" {
		t.Fatalf("after override: want alice (layer), got %q", got)
	}
	if got := ResolveOnCall(s, ovEnd); got != "alice" {
		t.Fatalf("override end is exclusive: want alice, got %q", got)
	}
}

func TestResolveOnCall_HigherLayerWins(t *testing.T) {
	start := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	base := model.ScheduleLayer{Position: 0, RotationLengthDays: 7, StartAt: start, UserIDs: pq.StringArray{"alice"}}
	top := model.ScheduleLayer{Position: 1, RotationLengthDays: 7, StartAt: start, UserIDs: pq.StringArray{"bob"}}
	// Layers are provided in position order (as the repository returns them).
	s := mkSchedule([]model.ScheduleLayer{base, top}, nil)
	if got := ResolveOnCall(s, start); got != "bob" {
		t.Fatalf("highest-position layer should win: want bob, got %q", got)
	}
}

func TestResolveOnCall_EmptyLayerSkipped(t *testing.T) {
	start := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	base := model.ScheduleLayer{Position: 0, RotationLengthDays: 7, StartAt: start, UserIDs: pq.StringArray{"alice"}}
	empty := model.ScheduleLayer{Position: 1, RotationLengthDays: 7, StartAt: start, UserIDs: pq.StringArray{}}
	s := mkSchedule([]model.ScheduleLayer{base, empty}, nil)
	if got := ResolveOnCall(s, start); got != "alice" {
		t.Fatalf("empty upper layer should be skipped: want alice, got %q", got)
	}
}

func TestResolveOnCall_ZeroRotationDefaultsToWeekly(t *testing.T) {
	start := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	layer := model.ScheduleLayer{RotationLengthDays: 0, StartAt: start, UserIDs: pq.StringArray{"alice", "bob"}}
	s := mkSchedule([]model.ScheduleLayer{layer}, nil)
	if got := ResolveOnCall(s, start.Add(7*24*time.Hour)); got != "bob" {
		t.Fatalf("zero rotation should default to 7d: want bob, got %q", got)
	}
}

func TestUrgencyFromSeverity(t *testing.T) {
	cases := map[string]string{
		"critical": "high",
		"error":    "high",
		"high":     "high",
		"CRITICAL": "high",
		"warning":  "low",
		"info":     "low",
		"low":      "low",
		"":         "high",
		"weird":    "high",
	}
	for in, want := range cases {
		if got := urgencyFromSeverity(in); got != want {
			t.Errorf("urgencyFromSeverity(%q): want %q, got %q", in, want, got)
		}
	}
}
