package tokens

import "github.com/golang-jwt/jwt/v5"

type Token = string

type Claims interface {
	jwt.Claims
	ID() string
	SetRegisteredClaims(claims jwt.RegisteredClaims)
}

type KeyID = string

const KeyIDFallbackValue = "0"

const KeyIDHeader = "kid" // standard JWT header that indicates which signing key we should use
