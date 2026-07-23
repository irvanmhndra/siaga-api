package postgres

import (
	"context"

	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/jmoiron/sqlx"
)

type notificationRepo struct{ db *sqlx.DB }

func NewNotificationRepository(db *sqlx.DB) *notificationRepo { return &notificationRepo{db} }

func (r *notificationRepo) Create(ctx context.Context, n *model.Notification) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications (id, organization_id, incident_id, user_id, channel, status, message)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		n.ID, n.OrganizationID, n.IncidentID, n.UserID, n.Channel, n.Status, n.Message,
	)
	return err
}

func (r *notificationRepo) ListByUser(ctx context.Context, userID string, unreadOnly bool, limit int) ([]model.Notification, error) {
	notifications := []model.Notification{}
	q := `SELECT * FROM notifications WHERE user_id = $1 AND channel = 'inapp'`
	if unreadOnly {
		q += " AND read_at IS NULL"
	}
	q += " ORDER BY created_at DESC LIMIT $2"
	err := r.db.SelectContext(ctx, &notifications, q, userID, limit)
	return notifications, err
}

func (r *notificationRepo) CountUnread(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND channel = 'inapp' AND read_at IS NULL`, userID)
	return n, err
}

func (r *notificationRepo) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read_at = NOW() WHERE id = $1 AND user_id = $2 AND read_at IS NULL`, id, userID)
	return err
}

func (r *notificationRepo) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL`, userID)
	return err
}
