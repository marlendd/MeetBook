package conference

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Client — интерфейс внешнего сервиса конференций.
type Client interface {
	CreateLink(ctx context.Context, bookingID uuid.UUID) (string, error)
}

// MockClient имитирует внешний сервис конференций.
// В реальном проекте здесь был бы HTTP-клиент к внешнему API.
type MockClient struct {
	// FailureRate задаёт вероятность сбоя (0.0–1.0), удобно для тестирования.
	FailureRate float64
	rng         *rand.Rand
}

func NewMockClient(failureRate float64) *MockClient {
	return &MockClient{
		FailureRate: failureRate,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
	}
}

func (m *MockClient) CreateLink(_ context.Context, bookingID uuid.UUID) (string, error) {
	if m.rng.Float64() < m.FailureRate {
		return "", fmt.Errorf("conference service unavailable")
	}
	return fmt.Sprintf("https://meet.example.com/%s", bookingID), nil
}
