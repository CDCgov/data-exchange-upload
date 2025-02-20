package oauth

import (
	"context"
	"errors"
	"github.com/coreos/go-oidc/v3/oidc"
	"slices"
	"strings"
)

var ErrTokenVerificationFailed = errors.New("failed to verify token")
var ErrTokenClaimsFailed = errors.New("failed to parse token claims")
var ErrTokenScopesMismatch = errors.New("one or more required scopes not found")

type Claims struct {
	Expiry int64  `json:"exp"`
	Scopes string `json:"scope"`
}

type Validator interface {
	ValidateJWT(ctx context.Context, token string) (Claims, error)
}

type PassthroughValidator struct{}

func (v PassthroughValidator) ValidateJWT(_ context.Context, _ string) (Claims, error) {
	return Claims{}, nil
}

func NewOAuthValidator(ctx context.Context, issuerUrl string, requiredScopes string) (*OAuthValidator, error) {
	var scopes []string
	if requiredScopes != "" {
		scopes = strings.Split(requiredScopes, " ")
	}

	provider, err := oidc.NewProvider(ctx, issuerUrl)
	if err != nil {
		return nil, err
	}

	return &OAuthValidator{
		IssuerUrl:      issuerUrl,
		RequiredScopes: scopes,
		provider:       provider,
	}, nil
}

type OAuthValidator struct {
	IssuerUrl      string
	RequiredScopes []string
	provider       *oidc.Provider
}

func (v OAuthValidator) ValidateJWT(ctx context.Context, token string) (Claims, error) {
	var claims Claims
	verifier := v.provider.Verifier(&oidc.Config{SkipClientIDCheck: true})
	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return claims, errors.Join(ErrTokenVerificationFailed, err)
	}

	if err = idToken.Claims(&claims); err != nil {
		return claims, errors.Join(ErrTokenClaimsFailed, err)
	}

	actualScopes := strings.Split(claims.Scopes, " ")

	if !hasRequiredScopes(actualScopes, v.RequiredScopes) {
		return claims, ErrTokenScopesMismatch
	}

	return claims, nil
}

func hasRequiredScopes(actualScopes, requiredScopes []string) bool {
	if len(requiredScopes) == 0 {
		return true
	}

	for _, reqScope := range requiredScopes {
		if !slices.Contains(actualScopes, reqScope) {
			return false
		}
	}
	return true
}
