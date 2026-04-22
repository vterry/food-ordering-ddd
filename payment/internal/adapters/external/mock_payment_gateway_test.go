package external

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/vterry/food-project/payment/internal/core/domain/payment"
)

func TestMockGateway_CircuitBreaker(t *testing.T) {
	gateway := NewMockGateway()
	p, _ := payment.NewPayment("p1", "o1", 1000)
	token, _ := payment.NewCardToken("valid_token")

	// We need to simulate failures to trip the circuit breaker.
	// The MockGateway has a 10% random failure, but we want to force it.
	// Since the MockGateway doesn't allow injecting failures easily in this implementation,
	// I'll wrap the logic or just run it enough times if random.
	// Actually, let's just test that it works as intended with the Execute call.

	t.Run("Circuit trips after consecutive failures", func(t *testing.T) {
		// We'll call it with a cancelled context to force "failures" that the CB counts.
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // already cancelled

		for i := 0; i < 4; i++ {
			_ = gateway.Authorize(ctx, p, token)
		}

		// The 5th call should be rejected by the CB itself if it's OPEN.
		err := gateway.Authorize(context.Background(), p, token)
		if !errors.Is(err, gobreaker.ErrOpenState) {
			t.Errorf("expected gobreaker.ErrOpenState, got %v", err)
		}
	})
}

func TestMockGateway_Timeout(t *testing.T) {
	gateway := NewMockGateway()
	p, _ := payment.NewPayment("p1", "o1", 1000)
	token, _ := payment.NewCardToken("valid_token")

	t.Run("Context timeout is respected", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := gateway.Authorize(ctx, p, token)
		if err == nil {
			t.Error("expected timeout error, got nil")
		}
	})
}
