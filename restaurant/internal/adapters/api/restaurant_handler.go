package api

import (
	"context"
	"time"

	"github.com/vterry/food-project/common/pkg/domain/vo"
	apperr "github.com/vterry/food-project/common/pkg/errors"
	apigen "github.com/vterry/food-project/restaurant/internal/adapters/api/generated"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
	"github.com/vterry/food-project/restaurant/internal/core/ports"
)

type RestaurantHandler struct {
	restaurantService ports.RestaurantUseCase
	ticketService     ports.TicketUseCase
	queryService      ports.RestaurantQueryUseCase
}

func NewRestaurantHandler(
	rs ports.RestaurantUseCase,
	ts ports.TicketUseCase,
	qs ports.RestaurantQueryUseCase,
) *RestaurantHandler {
	return &RestaurantHandler{
		restaurantService: rs,
		ticketService:     ts,
		queryService:      qs,
	}
}

// Admin / Restaurant Operations

func (h *RestaurantHandler) CreateRestaurant(ctx context.Context, request apigen.CreateRestaurantRequestObject) (apigen.CreateRestaurantResponseObject, error) {
	var hours []restaurant.OperatingPeriod
	if request.Body.OperatingHours != nil {
		for _, oh := range *request.Body.OperatingHours {
			hours = append(hours, restaurant.OperatingPeriod{
				DayOfWeek: time.Weekday(oh.DayOfWeek),
				Open:      oh.Open,
				Close:     oh.Close,
			})
		}
	}

	cmd := ports.CreateRestaurantCommand{
		Name: request.Body.Name,
		Address: restaurant.Address{
			Street:  request.Body.Address.Street,
			City:    request.Body.Address.City,
			ZipCode: request.Body.Address.ZipCode,
		},
		Hours: hours,
	}

	id, err := h.restaurantService.CreateRestaurant(ctx, cmd)
	if err != nil {
		return nil, h.handleError(err)
	}

	return apigen.CreateRestaurant201JSONResponse{Id: ptr(id.String())}, nil
}

func (h *RestaurantHandler) GetMenu(ctx context.Context, request apigen.GetMenuRequestObject) (apigen.GetMenuResponseObject, error) {
	restaurantID := vo.NewID(request.RestaurantId)

	infos, err := h.queryService.ListRestaurantMenuItems(ctx, restaurantID)
	if err != nil {
		return nil, h.handleError(err)
	}

	if len(infos) == 0 {
		return nil, h.handleError(apperr.NewNotFoundError("RESTAURANT_NOT_FOUND", "restaurant or menu items not found", nil))
	}

	items := make([]apigen.MenuItem, 0)
	for _, info := range infos {
		price := float32(info.Price.Amount())
		items = append(items, apigen.MenuItem{
			Id:          ptr(info.ID.String()),
			Name:        ptr(info.Name),
			Description: ptr(info.Description),
			Price:       &price,
			Category:    ptr(info.Category),
			Available:   ptr(info.IsAvailable),
		})
	}

	return apigen.GetMenu200JSONResponse{
		RestaurantId:   ptr(restaurantID.String()),
		RestaurantName: ptr(infos[0].RestaurantName),
		Items:          &items,
	}, nil
}

func (h *RestaurantHandler) CreateMenu(ctx context.Context, request apigen.CreateMenuRequestObject) (apigen.CreateMenuResponseObject, error) {
	id, err := h.restaurantService.CreateMenu(ctx, vo.NewID(request.RestaurantId), request.Body.Name)
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.CreateMenu201JSONResponse{Id: ptr(id.String())}, nil
}

func (h *RestaurantHandler) ActivateMenu(ctx context.Context, request apigen.ActivateMenuRequestObject) (apigen.ActivateMenuResponseObject, error) {
	err := h.restaurantService.ActivateMenu(ctx, vo.NewID(request.MenuId))
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.ActivateMenu204Response{}, nil
}

func (h *RestaurantHandler) AddMenuItem(ctx context.Context, request apigen.AddMenuItemRequestObject) (apigen.AddMenuItemResponseObject, error) {
	currency := "BRL"
	if request.Body.Currency != nil {
		currency = *request.Body.Currency
	}
	description := ""
	if request.Body.Description != nil {
		description = *request.Body.Description
	}

	price, _ := vo.NewMoney(float64(request.Body.Price), currency)
	cmd := ports.AddMenuItemCommand{
		MenuID:      vo.NewID(request.MenuId),
		Name:        request.Body.Name,
		Description: description,
		Price:       price,
		Category:    request.Body.Category,
	}

	id, err := h.restaurantService.AddItemToMenu(ctx, cmd)
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.AddMenuItem201JSONResponse{Id: ptr(id.String())}, nil
}

func (h *RestaurantHandler) UpdateItemAvailability(ctx context.Context, request apigen.UpdateItemAvailabilityRequestObject) (apigen.UpdateItemAvailabilityResponseObject, error) {
	err := h.restaurantService.UpdateItemAvailability(ctx, vo.NewID(request.MenuId), vo.NewID(request.ItemId), request.Body.Available)
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.UpdateItemAvailability204Response{}, nil
}

// Ticket Operations

func (h *RestaurantHandler) GetTicket(ctx context.Context, request apigen.GetTicketRequestObject) (apigen.GetTicketResponseObject, error) {
	t, err := h.ticketService.GetTicket(ctx, vo.NewID(request.Id))
	if err != nil {
		return nil, h.handleError(err)
	}

	items := make([]apigen.TicketItem, 0)
	for _, item := range t.Items() {
		items = append(items, apigen.TicketItem{
			ProductId: ptr(item.ProductID.String()),
			Name:      ptr(item.Name),
			Quantity:  ptr(item.Quantity),
		})
	}

	return apigen.GetTicket200JSONResponse{
		Id:           ptr(t.ID().String()),
		OrderId:      ptr(t.OrderID().String()),
		RestaurantId: ptr(t.RestaurantID().String()),
		Status:       ptr(string(t.Status())),
		Items:        &items,
	}, nil
}

func (h *RestaurantHandler) ConfirmTicket(ctx context.Context, request apigen.ConfirmTicketRequestObject) (apigen.ConfirmTicketResponseObject, error) {
	err := h.ticketService.ConfirmTicket(ctx, vo.NewID(request.Id))
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.ConfirmTicket204Response{}, nil
}

func (h *RestaurantHandler) StartPreparingTicket(ctx context.Context, request apigen.StartPreparingTicketRequestObject) (apigen.StartPreparingTicketResponseObject, error) {
	err := h.ticketService.StartPreparingTicket(ctx, vo.NewID(request.Id))
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.StartPreparingTicket204Response{}, nil
}

func (h *RestaurantHandler) MarkTicketAsReady(ctx context.Context, request apigen.MarkTicketAsReadyRequestObject) (apigen.MarkTicketAsReadyResponseObject, error) {
	err := h.ticketService.MarkTicketAsReady(ctx, vo.NewID(request.Id))
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.MarkTicketAsReady204Response{}, nil
}

func (h *RestaurantHandler) RejectTicket(ctx context.Context, request apigen.RejectTicketRequestObject) (apigen.RejectTicketResponseObject, error) {
	err := h.ticketService.RejectTicket(ctx, vo.NewID(request.Id), request.Body.Reason)
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.RejectTicket204Response{}, nil
}

func (h *RestaurantHandler) handleError(err error) error {
	return err
}

func ptr[T any](v T) *T {
	return &v
}
