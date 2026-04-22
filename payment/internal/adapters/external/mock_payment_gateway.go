package external

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sony/gobreaker"
	"github.com/vterry/food-project/payment/internal/core/domain/payment"
)

// MockGateway simulates an external payment provider with built-in resilience.
type MockGateway struct {
	cb *gobreaker.CircuitBreaker
}

// NewMockGateway initializes a MockGateway with a Circuit Breaker.
func NewMockGateway() *MockGateway {
	st := gobreaker.Settings{
		Name:        "PaymentGateway",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip the circuit if more than 3 consecutive failures occur.
			return counts.ConsecutiveFailures > 3
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			fmt.Printf("Circuit Breaker %s: %s -> %s\n", name, from, to)
		},
	}

	return &MockGateway{
		cb: gobreaker.NewCircuitBreaker(st),
	}
}

// Authorize simulates a card authorization.
func (g *MockGateway) Authorize(ctx context.Context, p *payment.Payment, token payment.CardToken) error {
	_, err := g.cb.Execute(func() (interface{}, error) {
		// Simulate network latency
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return nil, fmt.Errorf("gateway timeout: %w", ctx.Err())
		}

		// Simulate random failures (10% chance)
		if rand.Float32() < 0.1 {
			return nil, errors.New("simulated gateway failure")
		}

		return nil, nil
	})

	return err
}

// Capture simulates capturing funds from an authorized payment.
func (g *MockGateway) Capture(ctx context.Context, p *payment.Payment) error {
	_, err := g.cb.Execute(func() (interface{}, error) {
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return nil, fmt.Errorf("gateway timeout: %w", ctx.Err())
		}
		return nil, nil
	})
	return err
}

// Refund simulates refunding a captured payment.
func (g *MockGateway) Refund(ctx context.Context, p *payment.Payment) error {
	_, err := g.cb.Execute(func() (interface{}, error) {
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return nil, fmt.Errorf("gateway timeout: %w", ctx.Err())
		}
		return nil, nil
	})
	return err
}

// Release simulates releasing (voiding) an authorization.
func (g *MockGateway) Release(ctx context.Context, p *payment.Payment) error {
	_, err := g.cb.Execute(func() (interface{}, error) {
		select {
		case <-time.After(50 * time.Millisecond):
		case <-ctx.Done():
			return nil, fmt.Errorf("gateway timeout: %w", ctx.Err())
		}
		return nil, nil
	})
	return err
}
