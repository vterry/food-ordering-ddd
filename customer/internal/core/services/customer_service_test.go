package services

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	apperr "github.com/vterry/food-project/common/pkg/errors"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
	"github.com/vterry/food-project/customer/internal/core/ports"
	"testing"
)

// Mocks manuais para não depender de ferramentas de geração de mock agora
type MockCustomerRepo struct {
	customers map[string]*customer.Customer
}

func (m *MockCustomerRepo) Save(ctx context.Context, c *customer.Customer) error {
	m.customers[c.ID().String()] = c
	return nil
}

func (m *MockCustomerRepo) FindByID(ctx context.Context, id vo.ID) (*customer.Customer, error) {
	return m.customers[id.String()], nil
}

func (m *MockCustomerRepo) FindByEmail(ctx context.Context, email string) (*customer.Customer, error) {
	for _, c := range m.customers {
		if c.Email().String() == string(email) {
			return c, nil
		}
	}
	return nil, nil
}

type MockPublisher struct {
	Events []base.DomainEvent
}

func (m *MockPublisher) Publish(ctx context.Context, events ...base.DomainEvent) error {
	m.Events = append(m.Events, events...)
	return nil
}

func TestRegisterCustomer(t *testing.T) {
	repo := &MockCustomerRepo{customers: make(map[string]*customer.Customer)}
	pub := &MockPublisher{}
	svc := NewCustomerService(repo, pub)

	cmd := ports.RegisterCustomerCommand{
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "5511999999999",
	}

	id, err := svc.RegisterCustomer(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if id.IsEmpty() {
		t.Error("expected non-empty ID")
	}

	// Verificar se foi salvo no repo
	c, _ := repo.FindByID(context.Background(), id)
	if c == nil {
		t.Fatal("customer not found in repo")
	}

	// Verificar se evento foi publicado
	if len(pub.Events) != 1 {
		t.Errorf("expected 1 event published, got %d", len(pub.Events))
	}
}

func TestRegisterCustomer_DuplicateEmail(t *testing.T) {
	repo := &MockCustomerRepo{customers: make(map[string]*customer.Customer)}
	pub := &MockPublisher{}
	svc := NewCustomerService(repo, pub)

	cmd := ports.RegisterCustomerCommand{
		Name:  "John Doe",
		Email: "john@example.com",
		Phone: "5511999999999",
	}

	_, _ = svc.RegisterCustomer(context.Background(), cmd)
	_, err := svc.RegisterCustomer(context.Background(), cmd)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	aerr, ok := err.(*apperr.AppError)
	if !ok {
		t.Fatalf("expected *apperr.AppError, got %T", err)
	}

	if aerr.Code != "CUSTOMER_ALREADY_EXISTS" {
		t.Errorf("expected code CUSTOMER_ALREADY_EXISTS, got %s", aerr.Code)
	}

	if aerr.Type != apperr.ErrorTypeConflict {
		t.Errorf("expected type CONFLICT, got %s", aerr.Type)
	}
}
