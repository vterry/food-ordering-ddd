package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/delivery/internal/core/domain/delivery"
	"github.com/vterry/food-project/delivery/internal/core/ports"
)

type MockDeliveryRepository struct {
	mock.Mock
}

func (m *MockDeliveryRepository) Save(ctx context.Context, d *delivery.Delivery) error {
	args := m.Called(ctx, d)
	return args.Error(0)
}

func (m *MockDeliveryRepository) FindByID(ctx context.Context, id vo.ID) (*delivery.Delivery, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*delivery.Delivery), args.Error(1)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, events ...base.DomainEvent) error {
	args := m.Called(ctx, events)
	return args.Error(0)
}

func TestDeliveryService_ScheduleDelivery(t *testing.T) {
	repo := new(MockDeliveryRepository)
	service := NewDeliveryService(repo)

	cmd := ports.ScheduleDeliveryCommand{
		OrderID:      vo.NewID("o1"),
		RestaurantID: vo.NewID("r1"),
		CustomerID:   vo.NewID("c1"),
		Address:      delivery.NewAddress("s", "c", "z"),
	}

	repo.On("Save", mock.Anything, mock.AnythingOfType("*delivery.Delivery")).Return(nil)

	err := service.ScheduleDelivery(context.Background(), cmd)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeliveryService_PickUpDelivery(t *testing.T) {
	repo := new(MockDeliveryRepository)
	service := NewDeliveryService(repo)

	delID := vo.NewID("d1")
	d := delivery.NewDelivery(delID, vo.NewID("o1"), vo.NewID("r1"), vo.NewID("c1"), delivery.NewAddress("s", "c", "z"), vo.NewID(""))
	d.ClearEvents()

	courier, _ := delivery.NewCourierInfo("c1", "Courier")

	repo.On("FindByID", mock.Anything, delID).Return(d, nil)
	repo.On("Save", mock.Anything, d).Return(nil)

	err := service.PickUpDelivery(context.Background(), delID, courier)

	assert.NoError(t, err)
	assert.Equal(t, delivery.StatusPickedUp, d.Status())
	repo.AssertExpectations(t)
}
