package customer

import (
	"errors"
	"testing"
)

func TestNewName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  error
	}{
		{
			name:     "valid name",
			input:    "John Doe",
			expected: "John Doe",
			wantErr:  nil,
		},
		{
			name:     "empty name",
			input:    "",
			expected: "",
			wantErr:  ErrEmptyName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewName(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.String() != tt.expected {
				t.Errorf("NewName() = %v, expected %v", got.String(), tt.expected)
			}
		})
	}
}

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  error
	}{
		{
			name:     "valid email",
			input:    "john@example.com",
			expected: "john@example.com",
			wantErr:  nil,
		},
		{
			name:     "missing at",
			input:    "johnexample.com",
			expected: "",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "missing domain",
			input:    "john@",
			expected: "",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "empty email",
			input:    "",
			expected: "",
			wantErr:  ErrInvalidEmail,
		},
		{
			name:     "invalid format",
			input:    "john@com",
			expected: "",
			wantErr:  ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmail(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.String() != tt.expected {
				t.Errorf("NewEmail() = %v, expected %v", got.String(), tt.expected)
			}
		})
	}
}

func TestNewPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  error
	}{
		{
			name:     "valid phone br",
			input:    "5511999999999",
			expected: "5511999999999",
			wantErr:  nil,
		},
		{
			name:     "valid phone intl",
			input:    "+5511999999999",
			expected: "+5511999999999",
			wantErr:  nil,
		},
		{
			name:     "invalid phone with letters",
			input:    "abc",
			expected: "",
			wantErr:  ErrInvalidPhone,
		},
		{
			name:     "empty phone",
			input:    "",
			expected: "",
			wantErr:  ErrInvalidPhone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPhone(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewPhone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.String() != tt.expected {
				t.Errorf("NewPhone() = %v, expected %v", got.String(), tt.expected)
			}
		})
	}
}
