package ports

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/delivery/internal/core/domain/delivery"
)

type ScheduleDeliveryCommand struct {
	OrderID       vo.ID
	RestaurantID  vo.ID
	CustomerID    vo.ID
	Address       delivery.Address
	CorrelationID vo.ID
}

type DeliveryUseCase interface {
	ScheduleDelivery(ctx context.Context, cmd ScheduleDeliveryCommand) error
	CancelDelivery(ctx context.Context, deliveryID vo.ID, reason string) error
	PickUpDelivery(ctx context.Context, deliveryID vo.ID, courier delivery.CourierInfo) error
	CompleteDelivery(ctx context.Context, deliveryID vo.ID) error
	RefuseDelivery(ctx context.Context, deliveryID vo.ID, reason string) error
	GetDelivery(ctx context.Context, deliveryID vo.ID) (*delivery.Delivery, error)
}
