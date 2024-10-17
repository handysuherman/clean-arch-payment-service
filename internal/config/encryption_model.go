package config

import "github.com/handysuherman/clean-arch-payment-service/internal/pkg/token"

type EncryptionKeys string

const (
	ENCRYPTION EncryptionKeys = "encryption"
)

// Encryption holds information security related of the app.
type Encryption struct {
	Paseto *Paseto `mapstructure:"paseto"`
}

// Paseto holds information Paseto of the app.
type Paseto struct {
	BrowserDuration      *token.Duration `mapstructure:"browserDuration"`
	NativeMobileDuration *token.Duration `mapstructure:"nativeMobileDuration"`
}
