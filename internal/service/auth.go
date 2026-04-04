package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
)

var (
	adminUUID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userUUID  = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

type Claims struct {
	UserID uuid.UUID  `json:"user_id"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	jwtSecret string
	userRepo  UserRepo
	log       *slog.Logger
}

func NewAuthService(jwtSecret string, userRepo UserRepo, log *slog.Logger) *AuthService {
	return &AuthService{jwtSecret: jwtSecret, userRepo: userRepo, log: log}
}

func (a *AuthService) DummyLogin(role model.Role) (string, error) {
	var userID uuid.UUID
	switch role {
	case model.RoleAdmin:
		userID = adminUUID
	case model.RoleUser:
		userID = userUUID
	default:
		return "", model.ErrInvalidRole
	}

	return a.issueToken(userID, role)
}

func (a *AuthService) Register(ctx context.Context, email, password string, role model.Role) (*model.User, error) {
	existing, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		a.log.Error("register: failed to check email", "error", err)
		return nil, err
	}
	if existing != nil {
		return nil, model.ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error("register: failed to hash password", "error", err)
		return nil, err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
	}

	if err := a.userRepo.Create(ctx, user); err != nil {
		a.log.Error("register: failed to create user", "error", err)
		return nil, err
	}

	return user, nil
}

func (a *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		a.log.Error("login: failed to get user", "error", err)
		return "", err
	}
	if user == nil {
		return "", model.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", model.ErrInvalidCredentials
	}

	return a.issueToken(user.ID, user.Role)
}

func (a *AuthService) issueToken(userID uuid.UUID, role model.Role) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.jwtSecret))
}
