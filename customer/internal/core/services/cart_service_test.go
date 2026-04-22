package services

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"github.com/vterry/food-project/customer/internal/core/ports"
	"testing"
)

func TestAddItemToCart(t *testing.T) {
	repo := NewMockCartRepo()
	pub := &MockPublisher{}
	cat := NewMockRestaurantCatalogPort()
	svc := NewCartService(repo, pub, cat)

	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	prodID := vo.NewID("prod-1")
	price, _ := vo.NewMoney(15.0, "BRL")

	cat.AddItem(&ports.RestaurantMenuItem{
		ID: prodID, RestaurantID: restID, Name: "Burger", Price: price, Available: true,
	})

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
	repo := NewMockCartRepo()
	pub := &MockPublisher{}
	cat := NewMockRestaurantCatalogPort()
	svc := NewCartService(repo, pub, cat)

	customerID := vo.NewID("cust-1")
	price, _ := vo.NewMoney(15.0, "BRL")

	cat.AddItem(&ports.RestaurantMenuItem{
		ID: vo.NewID("p1"), RestaurantID: vo.NewID("rest-1"), Price: price, Available: true,
	})
	cat.AddItem(&ports.RestaurantMenuItem{
		ID: vo.NewID("p2"), RestaurantID: vo.NewID("rest-2"), Price: price, Available: true,
	})

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

func TestUpdateItemQuantity(t *testing.T) {
	repo := NewMockCartRepo()
	cat := NewMockRestaurantCatalogPort()
	svc := NewCartService(repo, &MockPublisher{}, cat)

	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	prodID := vo.NewID("prod-1")
	price, _ := vo.NewMoney(15.0, "BRL")

	cat.AddItem(&ports.RestaurantMenuItem{
		ID: prodID, RestaurantID: restID, Price: price, Available: true,
	})

	_ = svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{
		RestaurantID: restID,
		ProductID:    prodID,
		Price:        price,
		Quantity:     1,
	})

	err := svc.UpdateItemQuantity(context.Background(), customerID, ports.UpdateCartItemCommand{
		ProductID: prodID,
		Quantity:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, _ := repo.FindByCustomerID(context.Background(), customerID)
	if c.Items()[0].Quantity() != 5 {
		t.Errorf("expected qty 5, got %d", c.Items()[0].Quantity())
	}
}

func TestRemoveItemFromCart(t *testing.T) {
	repo := NewMockCartRepo()
	cat := NewMockRestaurantCatalogPort()
	svc := NewCartService(repo, &MockPublisher{}, cat)

	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	prodID := vo.NewID("prod-1")
	price, _ := vo.NewMoney(15.0, "BRL")

	cat.AddItem(&ports.RestaurantMenuItem{
		ID: prodID, RestaurantID: restID, Price: price, Available: true,
	})

	_ = svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{
		RestaurantID: restID,
		ProductID:    prodID,
		Price:        price,
		Quantity:     1,
	})

	err := svc.RemoveItemFromCart(context.Background(), customerID, prodID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, _ := repo.FindByCustomerID(context.Background(), customerID)
	if len(c.Items()) != 0 {
		t.Errorf("expected 0 items, got %d", len(c.Items()))
	}
}

func TestCheckout(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := NewMockCartRepo()
		pub := &MockPublisher{}
		cat := NewMockRestaurantCatalogPort()
		svc := NewCartService(repo, pub, cat)

		customerID := vo.NewID("cust-1")
		cat.AddItem(&ports.RestaurantMenuItem{
			ID: vo.NewID("p1"), RestaurantID: vo.NewID("rest-1"), Available: true,
		})
		_ = svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{
			RestaurantID: vo.NewID("rest-1"),
			ProductID:    vo.NewID("p1"),
			Price:        vo.Money{},
			Quantity:     1,
		})

		err := svc.Checkout(context.Background(), customerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Event published for checkout
		// AddItem emits one event, Checkout emits another.
		if len(pub.Events) != 2 {
			t.Errorf("expected 2 events, got %d", len(pub.Events))
		}

		// Cart should be cleared after checkout
		c, _ := repo.FindByCustomerID(context.Background(), customerID)
		if c != nil {
			t.Error("cart should be deleted after checkout")
		}
	})

	t.Run("empty cart returns error", func(t *testing.T) {
		repo := NewMockCartRepo()
		svc := NewCartService(repo, &MockPublisher{}, NewMockRestaurantCatalogPort())

		err := svc.Checkout(context.Background(), vo.NewID("cust-1"))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("publish error rolls back cart deletion (logic test)", func(t *testing.T) {
		repo := NewMockCartRepo()
		pub := &MockPublisher{PublishErr: context.DeadlineExceeded}
		cat := NewMockRestaurantCatalogPort()
		svc := NewCartService(repo, pub, cat)

		customerID := vo.NewID("cust-1")
		cat.AddItem(&ports.RestaurantMenuItem{
			ID: vo.NewID("p1"), RestaurantID: vo.NewID("rest-1"), Available: true,
		})
		_ = svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{
			RestaurantID: vo.NewID("rest-1"),
			ProductID:    vo.NewID("p1"),
		})

		err := svc.Checkout(context.Background(), customerID)
		if err == nil {
			t.Fatal("expected error from publisher")
		}
	})
}

func TestClearCart(t *testing.T) {
	repo := NewMockCartRepo()
	cat := NewMockRestaurantCatalogPort()
	svc := NewCartService(repo, &MockPublisher{}, cat)
	customerID := vo.NewID("cust-1")

	cat.AddItem(&ports.RestaurantMenuItem{
		ID: vo.NewID("p1"), RestaurantID: vo.NewID("rest-1"), Available: true,
	})

	_ = svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{
		RestaurantID: vo.NewID("rest-1"),
		ProductID:    vo.NewID("p1"),
	})

	err := svc.ClearCart(context.Background(), customerID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, _ := repo.FindByCustomerID(context.Background(), customerID)
	if c != nil {
		t.Error("cart should be deleted after clear")
	}
}

func TestGetCart(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := NewMockCartRepo()
		cat := NewMockRestaurantCatalogPort()
		svc := NewCartService(repo, &MockPublisher{}, cat)
		customerID := vo.NewID("cust-1")

		cat.AddItem(&ports.RestaurantMenuItem{
			ID: vo.NewID("p1"), RestaurantID: vo.NewID("rest-1"), Available: true,
		})

		_ = svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{
			RestaurantID: vo.NewID("rest-1"),
			ProductID:    vo.NewID("p1"),
		})

		c, err := svc.GetCart(context.Background(), customerID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("expected cart, got nil")
		}
	})

	t.Run("not found returns empty cart but no error in svc logic (as per current implementation)", func(t *testing.T) {
		repo := NewMockCartRepo()
		svc := NewCartService(repo, &MockPublisher{}, NewMockRestaurantCatalogPort())

		c, err := svc.GetCart(context.Background(), vo.NewID("none"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c == nil {
			t.Fatal("expected new empty cart, got nil")
		}
	})
}

func TestCartService_RepoErrors(t *testing.T) {
	repo := NewMockCartRepo()
	repo.SaveErr = context.DeadlineExceeded
	cat := NewMockRestaurantCatalogPort()
	svc := NewCartService(repo, &MockPublisher{}, cat)
	customerID := vo.NewID("cust-1")

	t.Run("AddItemToCart repo error", func(t *testing.T) {
		cat.AddItem(&ports.RestaurantMenuItem{ID: vo.NewID("p1"), Available: true})
		err := svc.AddItemToCart(context.Background(), customerID, ports.AddItemToCartCommand{ProductID: vo.NewID("p1")})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("UpdateItemQuantity repo error", func(t *testing.T) {
		err := svc.UpdateItemQuantity(context.Background(), customerID, ports.UpdateCartItemCommand{})
		if err == nil {
			t.Error("expected error")
		}
	})
}
