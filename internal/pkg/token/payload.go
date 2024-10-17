package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type Claimer struct {
	UID              string `json:"uid"`
	SystemLevel      int32  `json:"systemLevel"`
	SystemLevelLabel string `json:"systemLevelLabel"`
	RoleLevel        int32  `json:"roleLevel"`
}

type Payload struct {
	ID        uuid.UUID
	Claimer   *Claimer  `json:"credentials"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
	TokenType string    `json:"token_type"`
}

func NewPayload(claimer *Claimer, duration time.Duration, tokenType string) (*Payload, error) {
	tokenId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenId,
		Claimer:   claimer,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
		TokenType: tokenType,
	}

	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}

	return nil
}
