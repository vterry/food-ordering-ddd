package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	apperr "github.com/vterry/food-project/common/pkg/errors"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
	"github.com/vterry/food-project/customer/internal/core/ports"
)

var (
	ErrCustomerAlreadyExists = apperr.NewConflictError("CUSTOMER_ALREADY_EXISTS", "customer with this email already exists", nil)
	ErrCustomerNotFound      = apperr.NewNotFoundError("CUSTOMER_NOT_FOUND", "customer not found", nil)
)

type CustomerService struct {
	repo      ports.CustomerRepository
	publisher ports.EventPublisher
}

func NewCustomerService(repo ports.CustomerRepository, publisher ports.EventPublisher) *CustomerService {
	return &CustomerService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *CustomerService) RegisterCustomer(ctx context.Context, cmd ports.RegisterCustomerCommand) (vo.ID, error) {
	// 1. Validar se o cliente já existe (Regra de Idempotência/Negócio)
	existing, _ := s.repo.FindByEmail(ctx, cmd.Email)
	if existing != nil {
		return vo.ID{}, ErrCustomerAlreadyExists
	}

	// 2. Converter dados para VOs
	name, err := customer.NewName(cmd.Name)
	if err != nil {
		return vo.ID{}, err
	}

	email, err := customer.NewEmail(cmd.Email)
	if err != nil {
		return vo.ID{}, err
	}

	phone, err := customer.NewPhone(cmd.Phone)
	if err != nil {
		return vo.ID{}, err
	}

	// 3. Criar Agregado
	id := vo.NewID(uuid.New().String())
	c := customer.NewCustomer(id, name, email, phone)

	// 4. Persistir
	if err := s.repo.Save(ctx, c); err != nil {
		return vo.ID{}, apperr.NewInfrastructureError("DATABASE_ERROR", "failed to save customer", err)
	}

	// 5. Publicar Eventos
	if err := s.publisher.Publish(ctx, c.Events()...); err != nil {
		// Log error but don't fail the operation (or depend on Outbox/Retry later)
		fmt.Printf("failed to publish events: %v\n", err)
	}

	c.ClearEvents()

	return id, nil
}

func (s *CustomerService) AddAddress(ctx context.Context, customerID vo.ID, cmd ports.AddAddressCommand) error {
	// 1. Buscar cliente
	c, err := s.repo.FindByID(ctx, customerID)
	if err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to find customer", err)
	}
	if c == nil {
		return ErrCustomerNotFound
	}

	// 2. Criar endereço
	addrID := vo.NewID(uuid.New().String())
	addr := customer.NewAddress(addrID, customerID, cmd.Street, cmd.City, cmd.ZipCode, cmd.IsDefault)

	// 3. Adicionar ao cliente (agrega lógica de negócio)
	c.AddAddress(addr)

	// 4. Persistir (cascata via customer)
	if err := s.repo.Save(ctx, c); err != nil {
		return fmt.Errorf("failed to update customer addresses: %w", err)
	}

	// 5. Publicar Eventos
	if err := s.publisher.Publish(ctx, c.Events()...); err != nil {
		fmt.Printf("failed to publish events: %v\n", err)
	}
	c.ClearEvents()

	return nil
}

func (s *CustomerService) GetCustomer(ctx context.Context, customerID vo.ID) (*customer.Customer, error) {
	c, err := s.repo.FindByID(ctx, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find customer: %w", err)
	}
	if c == nil {
		return nil, ErrCustomerNotFound
	}
	return c, nil
}
