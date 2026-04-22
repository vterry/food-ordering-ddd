package payment

import (
	"errors"
	"strings"
)

var ErrInvalidCardToken = errors.New("invalid card token")

// CardToken represents a tokenized version of a payment card.
// It should never contain raw PAN data.
type CardToken struct {
	value string
}

// NewCardToken creates and validates a new CardToken.
func NewCardToken(value string) (CardToken, error) {
	if strings.TrimSpace(value) == "" {
		return CardToken{}, ErrInvalidCardToken
	}
	// Basic validation: should be a non-empty string.
	// In a real system, this might validate a specific format (e.g., UUID or provider-specific token).
	return CardToken{value: value}, nil
}

// String returns the string representation of the token.
func (t CardToken) String() string {
	return t.value
}
