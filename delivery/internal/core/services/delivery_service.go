package services

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/delivery/internal/core/domain/delivery"
	"github.com/vterry/food-project/delivery/internal/core/ports"
)

type DeliveryService struct {
	repo ports.DeliveryRepository
}

func NewDeliveryService(repo ports.DeliveryRepository) *DeliveryService {
	return &DeliveryService{
		repo: repo,
	}
}

func (s *DeliveryService) ScheduleDelivery(ctx context.Context, cmd ports.ScheduleDeliveryCommand) error {
	// ID generation should ideally be handled by a factory or service
	id := vo.NewID("") // This will be handled by the repository/DB in a real scenario, or generated here
	
	d := delivery.NewDelivery(id, cmd.OrderID, cmd.RestaurantID, cmd.CustomerID, cmd.Address, cmd.CorrelationID)
	
	return s.repo.Save(ctx, d)
}

func (s *DeliveryService) CancelDelivery(ctx context.Context, deliveryID vo.ID, reason string) error {
	d, err := s.repo.FindByID(ctx, deliveryID)
	if err != nil {
		return err
	}
	
	if err := d.Cancel(reason); err != nil {
		return err
	}
	
	return s.repo.Save(ctx, d)
}

func (s *DeliveryService) PickUpDelivery(ctx context.Context, deliveryID vo.ID, courier delivery.CourierInfo) error {
	d, err := s.repo.FindByID(ctx, deliveryID)
	if err != nil {
		return err
	}
	
	if err := d.PickUp(courier); err != nil {
		return err
	}
	
	return s.repo.Save(ctx, d)
}

func (s *DeliveryService) CompleteDelivery(ctx context.Context, deliveryID vo.ID) error {
	d, err := s.repo.FindByID(ctx, deliveryID)
	if err != nil {
		return err
	}
	
	if err := d.Complete(); err != nil {
		return err
	}
	
	return s.repo.Save(ctx, d)
}

func (s *DeliveryService) RefuseDelivery(ctx context.Context, deliveryID vo.ID, reason string) error {
	d, err := s.repo.FindByID(ctx, deliveryID)
	if err != nil {
		return err
	}
	
	if err := d.Refuse(reason); err != nil {
		return err
	}
	
	return s.repo.Save(ctx, d)
}

func (s *DeliveryService) GetDelivery(ctx context.Context, deliveryID vo.ID) (*delivery.Delivery, error) {
	return s.repo.FindByID(ctx, deliveryID)
}
