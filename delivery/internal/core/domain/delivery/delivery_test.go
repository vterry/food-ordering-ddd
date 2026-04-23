package delivery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

func TestNewDelivery(t *testing.T) {
	id := vo.NewID("del-123")
	orderID := vo.NewID("ord-123")
	restID := vo.NewID("rest-123")
	custID := vo.NewID("cust-123")
	addr := NewAddress("Street 1", "City", "12345")

	d := NewDelivery(id, orderID, restID, custID, addr, vo.NewID(""))

	assert.Equal(t, id, d.ID())
	assert.Equal(t, orderID, d.OrderID())
	assert.Equal(t, StatusScheduled, d.Status())
	assert.Len(t, d.Events(), 1)
	assert.IsType(t, DeliveryScheduled{}, d.Events()[0])
}

func TestDelivery_PickUp(t *testing.T) {
	d := NewDelivery(vo.NewID("d1"), vo.NewID("o1"), vo.NewID("r1"), vo.NewID("c1"), NewAddress("s", "c", "z"), vo.NewID(""))
	courier, _ := NewCourierInfo("c-1", "John Doe")

	t.Run("successful pickup", func(t *testing.T) {
		err := d.PickUp(courier)
		assert.NoError(t, err)
		assert.Equal(t, StatusPickedUp, d.Status())
		assert.Equal(t, &courier, d.Courier())
		assert.IsType(t, DeliveryPickedUp{}, d.Events()[1])
	})

	t.Run("fail if not scheduled", func(t *testing.T) {
		err := d.PickUp(courier)
		assert.ErrorIs(t, err, ErrInvalidStatusTransition)
	})
}

func TestDelivery_Complete(t *testing.T) {
	d := NewDelivery(vo.NewID("d1"), vo.NewID("o1"), vo.NewID("r1"), vo.NewID("c1"), NewAddress("s", "c", "z"), vo.NewID(""))
	courier, _ := NewCourierInfo("c-1", "John Doe")

	t.Run("fail if not picked up", func(t *testing.T) {
		err := d.Complete()
		assert.ErrorIs(t, err, ErrDeliveryNotPickedUp)
	})

	t.Run("successful completion", func(t *testing.T) {
		d.PickUp(courier)
		err := d.Complete()
		assert.NoError(t, err)
		assert.Equal(t, StatusDelivered, d.Status())
		assert.IsType(t, DeliveryCompleted{}, d.Events()[2])
	})
}

func TestDelivery_Refuse(t *testing.T) {
	d := NewDelivery(vo.NewID("d1"), vo.NewID("o1"), vo.NewID("r1"), vo.NewID("c1"), NewAddress("s", "c", "z"), vo.NewID(""))
	courier, _ := NewCourierInfo("c-1", "John Doe")

	t.Run("fail if not picked up", func(t *testing.T) {
		err := d.Refuse("client not home")
		assert.ErrorIs(t, err, ErrDeliveryNotPickedUp)
	})

	t.Run("successful refuse", func(t *testing.T) {
		d.PickUp(courier)
		err := d.Refuse("client not home")
		assert.NoError(t, err)
		assert.Equal(t, StatusRefused, d.Status())
		assert.IsType(t, DeliveryRefused{}, d.Events()[2])
	})
}

func TestDelivery_Cancel(t *testing.T) {
	t.Run("successful cancel from scheduled", func(t *testing.T) {
		d := NewDelivery(vo.NewID("d1"), vo.NewID("o1"), vo.NewID("r1"), vo.NewID("c1"), NewAddress("s", "c", "z"), vo.NewID(""))
		err := d.Cancel("out of stock")
		assert.NoError(t, err)
		assert.Equal(t, StatusCancelled, d.Status())
	})

	t.Run("fail if delivered", func(t *testing.T) {
		d := NewDelivery(vo.NewID("d1"), vo.NewID("o1"), vo.NewID("r1"), vo.NewID("c1"), NewAddress("s", "c", "z"), vo.NewID(""))
		courier, _ := NewCourierInfo("c-1", "John Doe")
		d.PickUp(courier)
		d.Complete()

		err := d.Cancel("error")
		assert.ErrorIs(t, err, ErrInvalidStatusTransition)
	})
}
