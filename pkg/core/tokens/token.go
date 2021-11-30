package tokens

import "github.com/golang-jwt/jwt/v4"

type Token = string

type Claims interface {
	jwt.Claims
	ID() string
	// KeyIDFallback returns KID when it is not provided as a header.
	// It helps us to built backwards compatibility with a tokens that did not have KID in the past.
	KeyIDFallback() (int, error)
	SetRegisteredClaims(claims jwt.RegisteredClaims)
}

const KeyIDHeader = "kid" // standard JWT header that indicates which signing key we should use
