package services

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"github.com/vterry/food-project/customer/internal/core/ports"
	"testing"
)

type MockCartRepo struct {
	carts map[string]*cart.Cart
}

func (m *MockCartRepo) Save(ctx context.Context, c *cart.Cart) error {
	m.carts[c.CustomerID().String()] = c
	return nil
}

func (m *MockCartRepo) FindByCustomerID(ctx context.Context, customerID vo.ID) (*cart.Cart, error) {
	return m.carts[customerID.String()], nil
}

func (m *MockCartRepo) Delete(ctx context.Context, customerID vo.ID) error {
	delete(m.carts, customerID.String())
	return nil
}

func TestAddItemToCart(t *testing.T) {
	repo := &MockCartRepo{carts: make(map[string]*cart.Cart)}
	pub := &MockPublisher{}
	svc := NewCartService(repo, pub)

	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	prodID := vo.NewID("prod-1")
	price, _ := vo.NewMoney(15.0, "BRL")

	cmd := ports.AddItemToCartCommand{
		RestaurantID: restID,
		ProductID:    prodID,
		Name:         "Burger",
		Price:        price,
		Quantity:     1,
	}

	err := svc.AddItemToCart(context.Background(), customerID, cmd)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verificar persistência
	c, _ := repo.FindByCustomerID(context.Background(), customerID)
	if c == nil {
		t.Fatal("cart not found in repo")
	}
	if len(c.Items()) != 1 {
		t.Errorf("expected 1 item, got %d", len(c.Items()))
	}

	// Verificar evento
	if len(pub.Events) != 1 {
		t.Errorf("expected 1 event published, got %d", len(pub.Events))
	}
}

func TestAddItemToCart_DifferentRestaurant(t *testing.T) {
	repo := &MockCartRepo{carts: make(map[string]*cart.Cart)}
	pub := &MockPublisher{}
	svc := NewCartService(repo, pub)

	customerID := vo.NewID("cust-1")
	price, _ := vo.NewMoney(15.0, "BRL")

	_ = svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{
		RestaurantID: vo.NewID("rest-1"),
		ProductID:    vo.NewID("p1"),
		Price:        price,
		Quantity:     1,
	})

	err := svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{
		RestaurantID: vo.NewID("rest-2"),
		ProductID:    vo.NewID("p2"),
		Price:        price,
		Quantity:     1,
	})

	if err != cart.ErrDifferentRestaurant {
		t.Errorf("expected ErrDifferentRestaurant, got %v", err)
	}
}
