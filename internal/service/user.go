package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users}
}

type ProfileRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Phone       *string `json:"phone"`
	Timezone    string  `json:"timezone" validate:"omitempty"`
	NotifyEmail *bool   `json:"notify_email"`
	NotifySMS   *bool   `json:"notify_sms"`
}

// InviteRequest adds a teammate to the current organization with a starting
// password they can change later.
type InviteRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (s *UserService) ListByOrg(ctx context.Context, orgID string) ([]model.User, error) {
	return s.users.ListByOrg(ctx, orgID)
}

// Invite creates a new member inside the given organization.
func (s *UserService) Invite(ctx context.Context, orgID string, req InviteRequest) (*model.User, error) {
	existing, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.Conflict("email already registered")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		Name:           req.Name,
		Email:          req.Email,
		PasswordHash:   string(hash),
		Timezone:       "Asia/Jakarta",
		NotifyEmail:    true,
		NotifySMS:      false,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, req ProfileRequest) (*model.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, apperror.NotFound("user not found")
	}
	u.Name = req.Name
	u.Phone = req.Phone
	if req.Timezone != "" {
		u.Timezone = req.Timezone
	}
	if req.NotifyEmail != nil {
		u.NotifyEmail = *req.NotifyEmail
	}
	if req.NotifySMS != nil {
		u.NotifySMS = *req.NotifySMS
	}
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
