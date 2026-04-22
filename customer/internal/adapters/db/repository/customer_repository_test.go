package repository

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
	"testing"
)

func TestSQLCustomerRepository_Integration(t *testing.T) {
	skipIfShort(t)
	db, teardown := setupMySQLContainer(t)
	defer teardown()

	repo := NewSQLCustomerRepository(db)
	ctx := context.Background()

	t.Run("Save and FindByID", func(t *testing.T) {
		id := vo.NewID("cust-1")
		name, _ := customer.NewName("John Doe")
		email, _ := customer.NewEmail("john@example.com")
		phone, _ := customer.NewPhone("5511999999999")
		c := customer.NewCustomer(id, name, email, phone)

		err := repo.Save(ctx, c)
		if err != nil {
			t.Fatalf("failed to save customer: %v", err)
		}

		got, err := repo.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("failed to find customer: %v", err)
		}
		if got == nil {
			t.Fatal("expected customer, got nil")
		}
		if got.Name().String() != "John Doe" {
			t.Errorf("expected name John Doe, got %s", got.Name().String())
		}
	})

	t.Run("Save with Addresses", func(t *testing.T) {
		id := vo.NewID("cust-2")
		name, _ := customer.NewName("Jane Due")
		email, _ := customer.NewEmail("jane@example.com")
		phone, _ := customer.NewPhone("5511888888888")
		c := customer.NewCustomer(id, name, email, phone)

		addr1 := customer.NewAddress(vo.NewID("addr-1"), id, "Street 1", "City", "12345", true)
		c.AddAddress(addr1)

		err := repo.Save(ctx, c)
		if err != nil {
			t.Fatalf("failed to save customer with addresses: %v", err)
		}

		got, err := repo.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("failed to find customer: %v", err)
		}
		if len(got.Addresses()) != 1 {
			t.Errorf("expected 1 address, got %d", len(got.Addresses()))
		}
		if got.Addresses()[0].Street() != "Street 1" {
			t.Errorf("expected street Street 1, got %s", got.Addresses()[0].Street())
		}
	})

	t.Run("Update Customer", func(t *testing.T) {
		id := vo.NewID("cust-1")
		c, _ := repo.FindByID(ctx, id)
		
		newName, _ := customer.NewName("John Updated")
		c.ChangeName(newName)

		err := repo.Save(ctx, c)
		if err != nil {
			t.Fatalf("failed to update customer: %v", err)
		}

		got, _ := repo.FindByID(ctx, id)
		if got.Name().String() != "John Updated" {
			t.Errorf("expected name John Updated, got %s", got.Name().String())
		}
	})

	t.Run("FindByEmail", func(t *testing.T) {
		got, err := repo.FindByEmail(ctx, "jane@example.com")
		if err != nil {
			t.Fatalf("failed to find by email: %v", err)
		}
		if got == nil || got.Email().String() != "jane@example.com" {
			t.Fatal("failed to find correct customer by email")
		}
	})

	t.Run("FindByID Not Found", func(t *testing.T) {
		got, err := repo.FindByID(ctx, vo.NewID("none"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Error("expected nil for non-existent id")
		}
	})
}
