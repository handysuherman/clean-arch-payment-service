package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/constants"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/serializer"
	"github.com/handysuherman/clean-arch-payment-service/internal/pkg/unierror"
)

const nonceSize = 12 // The recommended size for ChaCha20-Poly1305 nonce is 12 bytes

// PasetoMaker is responsible for creating and verifying Paseto tokens.
type PasetoMaker struct {
	privKey      paseto.V4AsymmetricSecretKey
	pubKey       paseto.V4AsymmetricPublicKey
	issuer       string
	implicitData []byte
}

// NewPaseto creates a new PasetoMaker with the specified private key, public key, and issuer.
func NewPaseto(privKey ed25519.PrivateKey, pubKey ed25519.PublicKey, issuer string, nonce string) (Maker, error) {
	pasetoPrivateKey, err := paseto.NewV4AsymmetricSecretKeyFromEd25519(privKey)
	if err != nil {
		return nil, err
	}
	pasetoPublicKey, err := paseto.NewV4AsymmetricPublicKeyFromEd25519(pubKey)
	if err != nil {
		return nil, err
	}

	implicitData, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return nil, err
	}

	return &PasetoMaker{
		privKey:      pasetoPrivateKey,
		pubKey:       pasetoPublicKey,
		issuer:       issuer,
		implicitData: implicitData,
	}, nil
}

// CreateToken creates a new Paseto token with the given claimer and duration.
// It returns the generated token string, payload, and an error if any.
func (maker *PasetoMaker) CreateToken(claimer *Claimer, duration time.Duration, tokenType string) (string, *Payload, error) {
	// Check The token type whether its was access or refresh, else return an error
	if tokenType != constants.AccessType && tokenType != constants.RefreshType {
		return "", nil, unierror.ErrInvalidTokenType
	}

	// Create a new payload based on the claimer and duration
	payload, err := NewPayload(claimer, duration, tokenType)
	if err != nil {
		return "", nil, err
	}

	// Serialize the payload to JSON
	body, err := serializer.Marshal(payload)
	if err != nil {
		return "", nil, err
	}

	// Create a new Paseto token and set its claims and issuer
	token, err := paseto.NewTokenFromClaimsJSON(body, nil)
	if err != nil {
		return "", nil, err
	}
	token.SetIssuedAt(payload.IssuedAt)
	token.SetNotBefore(payload.IssuedAt)
	token.SetExpiration(payload.ExpiredAt)
	token.SetIssuer(maker.issuer)

	// Sign the token using the private key
	pasetoToken := token.V4Sign(maker.privKey, maker.implicitData)

	return pasetoToken, payload, nil
}

// VerifyToken verifies a Paseto token and returns its payload.
// It checks the issuer and the token's validity at the current time.
// It returns the payload and an error if the verification fails.
func (maker *PasetoMaker) VerifyToken(tokens, expectedTokenType string) (*Payload, error) {
	// Create a new Paseto parser
	parser := paseto.NewParser()

	// Add rules for verification (issuer and validity at the current time)
	parser.AddRule(paseto.IssuedBy(maker.issuer))
	parser.AddRule(paseto.ValidAt(time.Now()))

	// Parse the Paseto token using the public key
	parsedToken, err := parser.ParseV4Public(maker.pubKey, tokens, maker.implicitData)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "expired") {
			return nil, unierror.ErrTokenExpired
		}
		return nil, err
	}

	// Unmarshal the token's claims from JSON
	var body Payload
	if err := serializer.Unmarshal(parsedToken.ClaimsJSON(), &body); err != nil {
		return nil, err
	}

	if expectedTokenType != body.TokenType {
		return nil, unierror.ErrInvalidTokenType
	}

	return &body, nil
}

// Function to generate ChaCha20-Poly1305 nonce
func generateChaCha20Poly1305Nonce() ([]byte, error) {
	nonce := make([]byte, nonceSize)
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}
	return nonce, nil
}
