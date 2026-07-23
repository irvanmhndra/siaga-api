package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/pkg/apperror"
	"github.com/lib/pq"
)

type ScheduleService struct {
	schedules repository.ScheduleRepository
}

func NewScheduleService(schedules repository.ScheduleRepository) *ScheduleService {
	return &ScheduleService{schedules}
}

type LayerRequest struct {
	Name               string    `json:"name" validate:"required"`
	RotationType       string    `json:"rotation_type" validate:"omitempty,oneof=daily weekly custom"`
	RotationLengthDays int       `json:"rotation_length_days" validate:"omitempty,min=1"`
	HandoffTime        string    `json:"handoff_time"`
	StartAt            time.Time `json:"start_at" validate:"required"`
	UserIDs            []string  `json:"user_ids" validate:"required,min=1"`
}

type ScheduleRequest struct {
	Name     string         `json:"name" validate:"required,min=2,max=100"`
	Timezone string         `json:"timezone" validate:"omitempty"`
	Layers   []LayerRequest `json:"layers" validate:"dive"`
}

type OverrideRequest struct {
	UserID  string    `json:"user_id" validate:"required"`
	StartAt time.Time `json:"start_at" validate:"required"`
	EndAt   time.Time `json:"end_at" validate:"required"`
}

func (s *ScheduleService) Create(ctx context.Context, orgID string, req ScheduleRequest) (*model.Schedule, error) {
	tz := req.Timezone
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	sch := &model.Schedule{ID: uuid.NewString(), OrganizationID: orgID, Name: req.Name, Timezone: tz}
	if err := s.schedules.Create(ctx, sch); err != nil {
		return nil, err
	}
	if err := s.schedules.ReplaceLayers(ctx, sch.ID, toLayers(req.Layers)); err != nil {
		return nil, err
	}
	return s.Get(ctx, sch.ID, orgID)
}

func (s *ScheduleService) List(ctx context.Context, orgID string) ([]model.Schedule, error) {
	return s.schedules.ListByOrg(ctx, orgID)
}

func (s *ScheduleService) Get(ctx context.Context, id, orgID string) (*model.Schedule, error) {
	sch, err := s.schedules.FindByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if sch == nil {
		return nil, apperror.NotFound("schedule not found")
	}
	return sch, nil
}

func (s *ScheduleService) Update(ctx context.Context, id, orgID string, req ScheduleRequest) (*model.Schedule, error) {
	sch, err := s.Get(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	sch.Name = req.Name
	if req.Timezone != "" {
		sch.Timezone = req.Timezone
	}
	if err := s.schedules.Update(ctx, sch); err != nil {
		return nil, err
	}
	if req.Layers != nil {
		if err := s.schedules.ReplaceLayers(ctx, sch.ID, toLayers(req.Layers)); err != nil {
			return nil, err
		}
	}
	return s.Get(ctx, id, orgID)
}

func (s *ScheduleService) Delete(ctx context.Context, id, orgID string) error {
	if _, err := s.Get(ctx, id, orgID); err != nil {
		return err
	}
	return s.schedules.Delete(ctx, id, orgID)
}

func (s *ScheduleService) AddOverride(ctx context.Context, scheduleID, orgID string, req OverrideRequest) (*model.ScheduleOverride, error) {
	if _, err := s.Get(ctx, scheduleID, orgID); err != nil {
		return nil, err
	}
	if !req.EndAt.After(req.StartAt) {
		return nil, apperror.BadRequest("end_at must be after start_at")
	}
	o := &model.ScheduleOverride{
		ID:         uuid.NewString(),
		ScheduleID: scheduleID,
		UserID:     req.UserID,
		StartAt:    req.StartAt,
		EndAt:      req.EndAt,
	}
	if err := s.schedules.CreateOverride(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (s *ScheduleService) DeleteOverride(ctx context.Context, scheduleID, orgID, overrideID string) error {
	if _, err := s.Get(ctx, scheduleID, orgID); err != nil {
		return err
	}
	return s.schedules.DeleteOverride(ctx, overrideID, scheduleID)
}

// OnCallNow returns the user ID currently on-call for a schedule.
func (s *ScheduleService) OnCallNow(ctx context.Context, id, orgID string) (string, error) {
	sch, err := s.Get(ctx, id, orgID)
	if err != nil {
		return "", err
	}
	return ResolveOnCall(sch, time.Now()), nil
}

func toLayers(reqs []LayerRequest) []model.ScheduleLayer {
	layers := make([]model.ScheduleLayer, 0, len(reqs))
	for _, r := range reqs {
		rt := r.RotationType
		if rt == "" {
			rt = "weekly"
		}
		days := r.RotationLengthDays
		if days <= 0 {
			days = 7
		}
		handoff := r.HandoffTime
		if handoff == "" {
			handoff = "09:00"
		}
		layers = append(layers, model.ScheduleLayer{
			Name:               r.Name,
			RotationType:       rt,
			RotationLengthDays: days,
			HandoffTime:        handoff,
			StartAt:            r.StartAt,
			UserIDs:            pq.StringArray(r.UserIDs),
		})
	}
	return layers
}
