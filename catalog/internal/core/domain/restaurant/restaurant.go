package restaurant

import (
	"errors"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

var (
	ErrAlreadyOpened = errors.New("restaurant is already opened")
	ErrAlreadyClosed = errors.New("restaurant is already closed")
	ErrMenuIdIsNil   = errors.New("active menu id is nil")
	ErrNoActiveMenu  = errors.New("missing an active menu")
)

type Restaurant struct {
	valueobjects.RestaurantID
	name         string
	address      valueobjects.Address
	status       enums.RestaurantStatus
	activeMenuID valueobjects.MenuID
	events       []common.DomainEvent
}

func NewRestaurant(name string, address valueobjects.Address) (*Restaurant, error) {

	if err := ValidateNewRestaurant(name, address); err != nil {
		return nil, err
	}

	r := &Restaurant{
		RestaurantID: valueobjects.NewRestaurantID(),
		name:         name,
		address:      address,
		status:       enums.RestaurantClosed,
	}
	r.AddEvent(NewRestaurantCreated(*r))
	return r, nil
}

func (r *Restaurant) Open() error {

	if r.status == enums.RestaurantOpened {
		return common.NewBusinessRuleErr(ErrAlreadyOpened)
	}

	if r.activeMenuID.IsZero() {
		return common.NewBusinessRuleErr(ErrNoActiveMenu)
	}

	r.status = enums.RestaurantOpened

	event := NewRestaurantOpened(r.RestaurantID)
	r.AddEvent(event)

	return nil
}

func (r *Restaurant) Close() error {

	if r.status == enums.RestaurantClosed {
		return common.NewBusinessRuleErr(ErrAlreadyClosed)
	}

	r.status = enums.RestaurantClosed

	event := NewRestaurantClosed(r.RestaurantID)
	r.AddEvent(event)

	return nil
}

func (r *Restaurant) UpdateMenu(activeMenuID valueobjects.MenuID) error {
	if activeMenuID.IsZero() {
		return common.NewBusinessRuleErr(ErrMenuIdIsNil)
	}
	r.activeMenuID = activeMenuID
	r.AddEvent(NewRestaurantMenuUpdated(r.RestaurantID, activeMenuID))
	return nil
}

func (r *Restaurant) CanAcceptOrder() bool {
	return r.status == enums.RestaurantOpened && !r.activeMenuID.IsZero()
}

func (r *Restaurant) Name() string                      { return r.name }
func (r *Restaurant) Address() valueobjects.Address     { return r.address }
func (r *Restaurant) Status() enums.RestaurantStatus    { return r.status }
func (r *Restaurant) ActiveMenuID() valueobjects.MenuID { return r.activeMenuID }

func Restore(
	restaurantId valueobjects.RestaurantID,
	name string,
	address valueobjects.Address,
	status enums.RestaurantStatus,
	activeMenuID valueobjects.MenuID,
) *Restaurant {
	return &Restaurant{
		RestaurantID: restaurantId,
		name:         name,
		address:      address,
		status:       status,
		activeMenuID: activeMenuID,
		events:       []common.DomainEvent{},
	}
}

func (r *Restaurant) AddEvent(events ...common.DomainEvent) {
	r.events = append(r.events, events...)
}

func (r *Restaurant) PullEvent() []common.DomainEvent {
	events := r.events
	r.events = nil
	return events
}
