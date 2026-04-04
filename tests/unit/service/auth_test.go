package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/internships-backend/test-backend-marlendd/internal/model"
	"github.com/internships-backend/test-backend-marlendd/internal/service"
)

const testSecret = "test-secret"

func newAuthSvc(repo *mockUserRepo) *service.AuthService {
	return service.NewAuthService(testSecret, repo, testLog)
}

func TestRegister_Success(t *testing.T) {
	svc := newAuthSvc(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, nil
		},
		createFn: func(_ context.Context, u *model.User) error {
			u.ID = uuid.New()
			return nil
		},
	})

	user, err := svc.Register(context.Background(), "new@example.com", "password123", model.RoleUser)
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, model.RoleUser, user.Role)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotEmpty(t, user.PasswordHash)
}

func TestRegister_EmailTaken(t *testing.T) {
	svc := newAuthSvc(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{Email: "taken@example.com"}, nil
		},
	})

	_, err := svc.Register(context.Background(), "taken@example.com", "password123", model.RoleUser)
	assert.ErrorIs(t, err, model.ErrEmailTaken)
}

func TestRegister_RepoError(t *testing.T) {
	dbErr := errors.New("db error")
	svc := newAuthSvc(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, dbErr
		},
	})

	_, err := svc.Register(context.Background(), "new@example.com", "password123", model.RoleUser)
	assert.ErrorIs(t, err, dbErr)
}

func TestLogin_Success(t *testing.T) {
	// сначала регистрируем, чтобы получить реальный bcrypt-хэш
	registerSvc := newAuthSvc(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) { return nil, nil },
		createFn:     func(_ context.Context, u *model.User) error { u.ID = uuid.New(); return nil },
	})
	user, err := registerSvc.Register(context.Background(), "login@example.com", "secret", model.RoleUser)
	require.NoError(t, err)

	svc := newAuthSvc(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return user, nil
		},
	})

	token, err := svc.Login(context.Background(), "login@example.com", "secret")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestLogin_WrongPassword(t *testing.T) {
	registerSvc := newAuthSvc(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) { return nil, nil },
		createFn:     func(_ context.Context, u *model.User) error { u.ID = uuid.New(); return nil },
	})
	user, err := registerSvc.Register(context.Background(), "login@example.com", "correct", model.RoleUser)
	require.NoError(t, err)

	svc := newAuthSvc(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return user, nil
		},
	})

	_, err = svc.Login(context.Background(), "login@example.com", "wrong")
	assert.ErrorIs(t, err, model.ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	svc := newAuthSvc(&mockUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, nil
		},
	})

	_, err := svc.Login(context.Background(), "nobody@example.com", "pass")
	assert.ErrorIs(t, err, model.ErrInvalidCredentials)
}

func TestDummyLogin_Admin(t *testing.T) {
	svc := newAuthSvc(&mockUserRepo{})

	token, err := svc.DummyLogin(model.RoleAdmin)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestDummyLogin_InvalidRole(t *testing.T) {
	svc := newAuthSvc(&mockUserRepo{})

	_, err := svc.DummyLogin("superuser")
	assert.ErrorIs(t, err, model.ErrInvalidRole)
}
