package service

import (
	"context"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
)

type NotificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo}
}

func (s *NotificationService) List(ctx context.Context, userID string, unreadOnly bool) ([]model.Notification, error) {
	return s.repo.ListByUser(ctx, userID, unreadOnly, 50)
}

func (s *NotificationService) CountUnread(ctx context.Context, userID string) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}
