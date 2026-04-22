package services

import (
	"context"
	"errors"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	apperr "github.com/vterry/food-project/common/pkg/errors"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
	"github.com/vterry/food-project/customer/internal/core/ports"
	"testing"
)

func TestRegisterCustomer(t *testing.T) {
	repo := NewMockCustomerRepo()
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

func TestRegisterCustomer_Errors(t *testing.T) {
	t.Run("duplicate email returns conflict", func(t *testing.T) {
		repo := NewMockCustomerRepo()
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

		if aerr.Slug != "CUSTOMER_ALREADY_EXISTS" {
			t.Errorf("expected slug CUSTOMER_ALREADY_EXISTS, got %s", aerr.Slug)
		}
	})

	t.Run("invalid email format returns error", func(t *testing.T) {
		repo := NewMockCustomerRepo()
		pub := &MockPublisher{}
		svc := NewCustomerService(repo, pub)

		cmd := ports.RegisterCustomerCommand{
			Name:  "John",
			Email: "invalid-email",
			Phone: "5511999999999",
		}

		_, err := svc.RegisterCustomer(context.Background(), cmd)
		if err != customer.ErrInvalidEmail {
			t.Errorf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("repository save error", func(t *testing.T) {
		repo := NewMockCustomerRepo()
		repo.SaveErr = errors.New("db down")
		pub := &MockPublisher{}
		svc := NewCustomerService(repo, pub)

		cmd := ports.RegisterCustomerCommand{
			Name:  "John",
			Email: "john@ex.com",
			Phone: "5511999999999",
		}

		_, err := svc.RegisterCustomer(context.Background(), cmd)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		aerr, ok := err.(*apperr.AppError)
		if !ok || aerr.Slug != "DATABASE_ERROR" {
			t.Errorf("expected DATABASE_ERROR, got %v", err)
		}
	})
}

func TestAddAddress(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := NewMockCustomerRepo()
		pub := &MockPublisher{}
		svc := NewCustomerService(repo, pub)

		custID := vo.NewID("123")
		name, _ := customer.NewName("John")
		email, _ := customer.NewEmail("j@j.com")
		phone, _ := customer.NewPhone("123456789")
		_ = repo.Save(context.Background(), customer.NewCustomer(custID, name, email, phone))

		cmd := ports.AddAddressCommand{
			Street:  "Main St",
			City:    "NY",
			ZipCode: "10001",
		}

		err := svc.AddAddress(context.Background(), custID, cmd)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		c, _ := repo.FindByID(context.Background(), custID)
		if len(c.Addresses()) != 1 {
			t.Errorf("expected 1 address, got %d", len(c.Addresses()))
		}
	})

	t.Run("customer not found", func(t *testing.T) {
		repo := NewMockCustomerRepo()
		pub := &MockPublisher{}
		svc := NewCustomerService(repo, pub)

		err := svc.AddAddress(context.Background(), vo.NewID("none"), ports.AddAddressCommand{})
		if err != ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound, got %v", err)
		}
	})
}

func TestGetCustomer(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := NewMockCustomerRepo()
		svc := NewCustomerService(repo, &MockPublisher{})

		custID := vo.NewID("123")
		name, _ := customer.NewName("John")
		email, _ := customer.NewEmail("j@j.com")
		phone, _ := customer.NewPhone("123456789")
		_ = repo.Save(context.Background(), customer.NewCustomer(custID, name, email, phone))

		got, err := svc.GetCustomer(context.Background(), custID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID().String() != custID.String() {
			t.Errorf("expected id %v, got %v", custID, got.ID())
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := NewMockCustomerRepo()
		svc := NewCustomerService(repo, &MockPublisher{})

		_, err := svc.GetCustomer(context.Background(), vo.NewID("none"))
		if err != ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound, got %v", err)
		}
	})
}

