package services

import (
	"context"
	"fmt"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	apperr "github.com/vterry/food-project/common/pkg/errors"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"github.com/vterry/food-project/customer/internal/core/ports"
)

type CartService struct {
	repo          ports.CartRepository
	publisher     ports.EventPublisher
	catalogClient ports.RestaurantCatalogPort
}

func NewCartService(repo ports.CartRepository, publisher ports.EventPublisher, catalogClient ports.RestaurantCatalogPort) *CartService {
	return &CartService{
		repo:          repo,
		publisher:     publisher,
		catalogClient: catalogClient,
	}
}

func (s *CartService) GetCart(ctx context.Context, customerID vo.ID) (*cart.Cart, error) {
	c, err := s.repo.FindByCustomerID(ctx, customerID)
	if err != nil {
		return nil, apperr.NewInfrastructureError("DATABASE_ERROR", "failed to get cart", err)
	}

	// Se não existe, cria um novo transiente
	if c == nil {
		c = cart.NewCart(vo.NewID("cart-"+customerID.String()), customerID)
	}

	return c, nil
}
func (s *CartService) AddItemToCart(ctx context.Context, customerID vo.ID, cmd ports.AddItemToCartCommand) error {
	// 1. Validar item no catálogo externo (Restaurant Service)
	menuItem, err := s.catalogClient.GetMenuItem(ctx, cmd.RestaurantID, cmd.ProductID)
	if err != nil {
		return fmt.Errorf("failed to validate menu item: %w", err)
	}
	if menuItem == nil || !menuItem.Available {
		return apperr.NewDomainError("ITEM_UNAVAILABLE", "item is not available in restaurant catalog", nil)
	}

	// 2. Buscar carrinho
	c, err := s.GetCart(ctx, customerID)
	if err != nil {
		return err
	}

	// 2. Criar Item
	item := cart.NewCartItem(cmd.ProductID, cmd.Name, cmd.Price, cmd.Quantity, cmd.Observation)

	// 3. Adicionar (dispara invariantes e eventos)
	if err := c.AddItem(cmd.RestaurantID, item); err != nil {
		return err
	}

	// 4. Salvar
	if err := s.repo.Save(ctx, c); err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to save cart", err)
	}

	// 5. Publicar eventos
	if err := s.publisher.Publish(ctx, c.Events()...); err != nil {
		fmt.Printf("failed to publish events: %v\n", err)
	}
	c.ClearEvents()

	return nil
}

func (s *CartService) UpdateItemQuantity(ctx context.Context, customerID vo.ID, cmd ports.UpdateCartItemCommand) error {
	c, err := s.GetCart(ctx, customerID)
	if err != nil {
		return err
	}

	c.UpdateItemQuantity(cmd.ProductID, cmd.Quantity)

	if err := s.repo.Save(ctx, c); err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to save cart", err)
	}

	return nil
}

func (s *CartService) RemoveItemFromCart(ctx context.Context, customerID vo.ID, productID vo.ID) error {
	c, err := s.GetCart(ctx, customerID)
	if err != nil {
		return err
	}

	c.RemoveItem(productID)

	if err := s.repo.Save(ctx, c); err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to save cart", err)
	}

	return nil
}

func (s *CartService) Checkout(ctx context.Context, customerID vo.ID) error {
	// 1. Buscar carrinho
	c, err := s.repo.FindByCustomerID(ctx, customerID)
	if err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to find cart for checkout", err)
	}
	if c == nil || len(c.Items()) == 0 {
		return apperr.NewDomainError("CART_EMPTY", "cannot checkout an empty cart", nil)
	}

	// 2. Executar checkout no agregado (gera evento)
	if err := c.Checkout(); err != nil {
		return apperr.NewDomainError("CHECKOUT_FAILED", err.Error(), err)
	}

	// 3. Publicar evento de integração (CheckoutRequested)
	if err := s.publisher.Publish(ctx, c.Events()...); err != nil {
		return apperr.NewInfrastructureError("MESSAGING_ERROR", "failed to publish checkout event", err)
	}

	// 4. Limpar carrinho após checkout de sucesso
	if err := s.ClearCart(ctx, customerID); err != nil {
		// Log error but don't fail as the event was already sent
		fmt.Printf("failed to clear cart after checkout: %v\n", err)
	}

	return nil
}

func (s *CartService) ClearCart(ctx context.Context, customerID vo.ID) error {
	if err := s.repo.Delete(ctx, customerID); err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to clear cart", err)
	}
	return nil
}
