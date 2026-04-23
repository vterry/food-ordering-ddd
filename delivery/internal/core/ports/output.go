package ports

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/delivery/internal/core/domain/delivery"
)

type DeliveryRepository interface {
	Save(ctx context.Context, d *delivery.Delivery) error
	FindByID(ctx context.Context, id vo.ID) (*delivery.Delivery, error)
}
