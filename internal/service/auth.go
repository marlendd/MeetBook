package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
}

func NewAuthService(jwtSecret string) *AuthService {
	return &AuthService{jwtSecret: jwtSecret}
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
