package valueobjects

import "errors"

var (
	ErrStreetIsEmpty  = errors.New("street name cannot be empty")
	ErrCityIsEmpty    = errors.New("city and state cannot be empty")
	ErrZipCodeIsEmpty = errors.New("zipcode cannot be empty")
)

type DeliveryAddress struct {
	street       string
	number       string
	neighborhood string
	city         string
	state        string
	zipCode      string
	complement   string
}

func NewDeliveryAddress(street, number, complement, neighborhood, city, state, zipCode string) (DeliveryAddress, error) {
	addr := DeliveryAddress{
		street:       street,
		number:       number,
		complement:   complement,
		neighborhood: neighborhood,
		city:         city,
		state:        state,
		zipCode:      zipCode,
	}

	if err := addr.Validate(); err != nil {
		return DeliveryAddress{}, err
	}
	return addr, nil
}

func (a *DeliveryAddress) Validate() error {
	if a.street == "" {
		return ErrStreetIsEmpty
	}
	if a.city == "" || a.state == "" {
		return ErrCityIsEmpty
	}
	if a.zipCode == "" {
		return ErrZipCodeIsEmpty
	}
	return nil
}

func (a *DeliveryAddress) Street() string {
	return a.street
}
func (a *DeliveryAddress) Number() string {
	return a.number
}

func (a *DeliveryAddress) Complement() string {
	return a.complement
}

func (a *DeliveryAddress) Neighborhood() string {
	return a.neighborhood
}

func (a *DeliveryAddress) City() string {
	return a.city
}

func (a *DeliveryAddress) State() string {
	return a.state
}
func (a *DeliveryAddress) ZipCode() string {
	return a.zipCode
}
