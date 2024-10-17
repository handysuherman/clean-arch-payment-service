package helper

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
)

var (
	isValidNumber             = regexp.MustCompile(`^[0-9]+$`).MatchString
	ErrPhoneNumberNotVerified = errors.New("unverified phone number")
	ErrEmailNotVerified       = errors.New("unverified email")
	ErrInvalidPhoneNumber     = errors.New("invalid phone number")
	ErrInvalidEmail           = errors.New("invalid email")
	ErrInvalidSerialDigits    = errors.New("invalid serial digits")
	ErrInvalidStokistID       = errors.New("invalid stokist ID")
)

// validateString checks if the length of the given value is within the specified range.
// It returns an error if the length is not within the range.
func validateString(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("must contain from %d-%d characters", minLength, maxLength)
	}
	return nil
}

// validatePhoneNumber validates the given phone number.
func validatePhoneNumber(value string) (string, error) {
	if err := validateString(value, 9, 13); err != nil {
		return "", err
	}

	if value[0:3] != "628" && value[0:2] != "08" {
		return "", fmt.Errorf("not a valid number")
	}

	if value[0:2] == "08" {
		return strings.Replace(value, value[0:2], "628", 1), nil
	}

	return value, nil
}

// validateEmail validates the given email address.
func validateEmail(value string) (string, error) {
	if err := validateString(value, 3, 200); err != nil {
		return "", err
	}

	return value, nil
}

// validateStokists validates the given stokist ID.
func validateStokists(value string) (string, error) {
	if err := validateString(value, 2, 32); err != nil {
		return "", err
	}

	if len(value) >= 5 {
		if value[0:3] != "998" {
			return "", fmt.Errorf("invalid stokist id")
		}

		return value[3:], nil
	}

	return value, nil
}

// ValidatePhoneNumber validates and normalizes the given phone number.
func ValidatePhoneNumber(value string) (string, error) {
	if err := validateString(value, 9, 13); err != nil {
		return "", err
	}

	if value[0:3] != "628" && value[0:2] != "08" {
		return "", fmt.Errorf("not a valid number")
	}

	if value[0:2] == "08" {
		return strings.Replace(value, value[0:2], "628", 1), nil
	}

	return value, nil
}

// ValidateContact validates and determines the type of contact (email or phone number).
func ValidateContact(value string) (string, string, error) {
	if isValidNumber(value) {
		res, err := validatePhoneNumber(value)
		if err != nil {
			return "", "", err
		}
		return res, constants.Whatsapp, nil
	}

	if _, err := mail.ParseAddress(value); err == nil {
		res, err := validateEmail(value)
		if err != nil {
			return "", "", err
		}
		return res, constants.Email, nil
	}

	return "", "", fmt.Errorf("should supply a valid email or phone number")
}

// ValidateDigits validates a string containing only digits.
func ValidateDigits(value string) error {
	if isValidNumber(value) {
		if err := validateString(value, 6, 10); err != nil {
			return err
		}

		return nil
	}
	return fmt.Errorf("not a valid serial digits")
}

// ValidateStokists validates and normalizes the given stokist ID.
func ValidateStokists(value string) (string, error) {
	if isValidNumber(value) {
		val, err := validateStokists(value)
		if err != nil {
			return "", fmt.Errorf("%s: %v", ErrInvalidStokistID, err)
		}

		return val, nil
	}

	return "", fmt.Errorf("%s: %v", ErrInvalidStokistID, value)
}
