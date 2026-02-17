package valueobjects

import "errors"

var (
	ErrStreetIsEmpty  = errors.New("street name cannot be empty")
	ErrCityIsEmpty    = errors.New("city and state cannot be empty")
	ErrZipCodeIsEmpty = errors.New("zipcode cannot be empty")
)

type Address struct {
	street       string
	number       string
	complement   string
	neighborhood string
	city         string
	state        string
	zipCode      string
}

func NewAddress(street, number, complement, neighborhood, city, state, zipCode string) (Address, error) {
	addr := Address{
		street:       street,
		number:       number,
		complement:   complement,
		neighborhood: neighborhood,
		city:         city,
		state:        state,
		zipCode:      zipCode,
	}
	if err := addr.Validate(); err != nil {
		return Address{}, err
	}
	return addr, nil
}

func (a *Address) Validate() error {
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

func (a *Address) Street() string {
	return a.street
}
func (a *Address) Number() string {
	return a.number
}

func (a *Address) Complement() string {
	return a.complement
}

func (a *Address) Neighborhood() string {
	return a.neighborhood
}

func (a *Address) City() string {
	return a.city
}

func (a *Address) State() string {
	return a.state
}
func (a *Address) ZipCode() string {
	return a.zipCode
}
