package token

import "time"

type Duration struct {
	AccessToken  time.Duration `mapstructure:"accessToken"`
	RefreshToken time.Duration `mapstructure:"refreshToken"`
}

type Maker interface {
	CreateToken(claimer *Claimer, duration time.Duration, tokenType string) (string, *Payload, error)
	VerifyToken(tokens, expectedTokenType string) (*Payload, error)
}
