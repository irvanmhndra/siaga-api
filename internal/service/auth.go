package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/irvanmhndra/siaga-api/internal/model"
	"github.com/irvanmhndra/siaga-api/internal/repository"
	"github.com/irvanmhndra/siaga-api/pkg/apperror"
	"github.com/irvanmhndra/siaga-api/pkg/sessiontoken"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	orgRepo     repository.OrganizationRepository
	userRepo    repository.UserRepository
	sessionRepo repository.UserSessionRepository
}

func NewAuthService(orgRepo repository.OrganizationRepository, userRepo repository.UserRepository, sessionRepo repository.UserSessionRepository) *AuthService {
	return &AuthService{orgRepo, userRepo, sessionRepo}
}

type RegisterRequest struct {
	Name             string `json:"name" validate:"required,min=2"`
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=6"`
	OrganizationName string `json:"organization_name" validate:"omitempty,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         *model.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
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

	// First user of a fresh sign-up owns a new organization (the tenant).
	orgName := req.OrganizationName
	if orgName == "" {
		orgName = req.Name + "'s Organization"
	}
	org := &model.Organization{
		ID:   uuid.NewString(),
		Name: orgName,
	}
	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	user := &model.User{
		ID:             uuid.NewString(),
		OrganizationID: org.ID,
		Name:           req.Name,
		Email:          req.Email,
		PasswordHash:   string(hash),
		Timezone:       "Asia/Jakarta",
		NotifyEmail:    true,
		NotifySMS:      false,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return s.createSession(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}
	return s.createSession(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.DeleteByTokenHash(ctx, sessiontoken.Hash(token))
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	session, err := s.sessionRepo.FindByRefreshTokenHash(ctx, sessiontoken.Hash(refreshToken))
	if err != nil {
		return nil, err
	}
	if session == nil || time.Now().After(session.RefreshExpiresAt) {
		return nil, apperror.TokenExpired()
	}
	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, apperror.Unauthorized("user not found")
	}
	// Replace the old session with the new one atomically
	next, resp := newSession(user)
	rotated, err := s.sessionRepo.Rotate(ctx, session.ID, next)
	if err != nil {
		return nil, err
	}
	if !rotated {
		// Another request exchanged this refresh token first.
		return nil, apperror.TokenExpired()
	}
	return resp, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*model.User, error) {
	session, err := s.sessionRepo.FindByTokenHash(ctx, sessiontoken.Hash(token))
	if err != nil {
		return nil, err
	}
	if session == nil || time.Now().After(session.ExpiresAt) {
		return nil, apperror.TokenExpired()
	}
	return s.userRepo.FindByID(ctx, session.UserID)
}

func (s *AuthService) Me(ctx context.Context, userID string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

const (
	accessTokenTTL  = 2 * time.Hour
	refreshTokenTTL = 30 * 24 * time.Hour
)

func (s *AuthService) createSession(ctx context.Context, user *model.User) (*AuthResponse, error) {
	session, resp := newSession(user)
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return resp, nil
}

// newSession issues fresh tokens: the response carries the raw tokens for the
// client, the session row only their digests.
func newSession(user *model.User) (*model.UserSession, *AuthResponse) {
	token := sessiontoken.New()
	refresh := sessiontoken.New()
	now := time.Now()
	session := &model.UserSession{
		ID:               uuid.NewString(),
		UserID:           user.ID,
		TokenHash:        sessiontoken.Hash(token),
		RefreshTokenHash: sessiontoken.Hash(refresh),
		ExpiresAt:        now.Add(accessTokenTTL),
		RefreshExpiresAt: now.Add(refreshTokenTTL),
	}
	return session, &AuthResponse{AccessToken: token, RefreshToken: refresh, User: user}
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
