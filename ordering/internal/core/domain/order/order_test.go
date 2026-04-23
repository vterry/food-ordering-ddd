package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

func TestNewOrder(t *testing.T) {
	customerID := vo.NewID("cust-1")
	restaurantID := vo.NewID("rest-1")
	orderID := vo.NewID("ord-1")
	correlationID := vo.NewID("corr-1")
	
	price := vo.NewMoneyFromFloat(10.0, "BRL")
	item := NewOrderItem(vo.NewID("item-1"), "Pizza", 2, price, "No onion")
	
	t.Run("should create valid order", func(t *testing.T) {
		o, err := NewOrder(orderID, customerID, restaurantID, []OrderItem{item}, "Street 1", correlationID)
		
		require.NoError(t, err)
		assert.Equal(t, orderID, o.ID())
		assert.Equal(t, StatusCreated, o.Status())
		assert.Equal(t, 20.0, o.TotalAmount().Amount())
		assert.Equal(t, 1, o.Version())
		assert.Len(t, o.Events(), 1)
		assert.IsType(t, OrderCreatedEvent{}, o.Events()[0])
	})
	
	t.Run("should fail with no items", func(t *testing.T) {
		_, err := NewOrder(orderID, customerID, restaurantID, []OrderItem{}, "Street 1", correlationID)
		assert.Error(t, err)
	})
}

func TestOrderTransitions(t *testing.T) {
	customerID := vo.NewID("cust-1")
	restaurantID := vo.NewID("rest-1")
	orderID := vo.NewID("ord-1")
	correlationID := vo.NewID("corr-1")
	price := vo.NewMoneyFromFloat(10.0, "BRL")
	item := NewOrderItem(vo.NewID("item-1"), "Pizza", 1, price, "")
	
	setupOrder := func() *Order {
		o, _ := NewOrder(orderID, customerID, restaurantID, []OrderItem{item}, "Street 1", correlationID)
		o.ClearEvents()
		return o
	}

	t.Run("happy path flow", func(t *testing.T) {
		o := setupOrder()
		
		err := o.MarkPaymentPending()
		require.NoError(t, err)
		assert.Equal(t, StatusAuthorizingPayment, o.Status())
		assert.Equal(t, 2, o.Version())
		
		err = o.MarkPaymentAuthorized()
		require.NoError(t, err)
		assert.Equal(t, StatusAwaitingRestaurantConfirmation, o.Status())
		
		err = o.MarkRestaurantConfirmed()
		require.NoError(t, err)
		assert.Equal(t, StatusCapturingPayment, o.Status())
		
		err = o.MarkPaymentCaptured()
		require.NoError(t, err)
		assert.Equal(t, StatusSchedulingDelivery, o.Status())
		
		err = o.MarkDeliveryScheduled()
		require.NoError(t, err)
		assert.Equal(t, StatusPreparing, o.Status())
		
		err = o.MarkReady()
		require.NoError(t, err)
		assert.Equal(t, StatusReady, o.Status())
		
		err = o.MarkOutForDelivery()
		require.NoError(t, err)
		assert.Equal(t, StatusOutForDelivery, o.Status())
		
		err = o.MarkDelivered()
		require.NoError(t, err)
		assert.Equal(t, StatusDelivered, o.Status())
	})

	t.Run("payment rejected flow", func(t *testing.T) {
		o := setupOrder()
		_ = o.MarkPaymentPending()
		
		err := o.Reject("Insufficient funds")
		require.NoError(t, err)
		assert.Equal(t, StatusRejected, o.Status())
	})

	t.Run("restaurant rejected flow", func(t *testing.T) {
		o := setupOrder()
		_ = o.MarkPaymentPending()
		_ = o.MarkPaymentAuthorized()
		
		err := o.MarkRestaurantRejected("Out of stock")
		require.NoError(t, err)
		assert.Equal(t, StatusRestaurantRejected, o.Status())
		
		err = o.MarkCancelled()
		require.NoError(t, err)
		assert.Equal(t, StatusCancelled, o.Status())
	})

	t.Run("customer cancellation flow", func(t *testing.T) {
		o := setupOrder()
		_ = o.MarkPaymentPending()
		_ = o.MarkPaymentAuthorized()
		
		err := o.Cancel()
		require.NoError(t, err)
		assert.Equal(t, StatusCancelling, o.Status())
		
		err = o.MarkCancelled()
		require.NoError(t, err)
		assert.Equal(t, StatusCancelled, o.Status())
	})

	t.Run("invalid transition should fail", func(t *testing.T) {
		o := setupOrder()
		// Current: CREATED
		err := o.MarkRestaurantConfirmed() // CREATED -> CAPTURING_PAYMENT is invalid
		assert.ErrorIs(t, err, ErrInvalidStateTransition)
	})
}
