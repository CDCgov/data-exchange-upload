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
	Expiry float64 `json:"exp"`
	Scopes string  `json:"scope"`
}

func NewOAuthValidator(issuerUrl string, requiredScopes string) OAuthValidator {
	var scopes []string
	if requiredScopes != "" {
		scopes = strings.Split(requiredScopes, " ")
	}

	return OAuthValidator{
		IssuerUrl:      issuerUrl,
		RequiredScopes: scopes,
	}
}

type OAuthValidator struct {
	IssuerUrl      string
	RequiredScopes []string
}

func (v OAuthValidator) ValidateJWT(ctx context.Context, token string) error {
	provider, err := oidc.NewProvider(ctx, v.IssuerUrl)
	if err != nil {
		return err
	}

	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})
	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return errors.Join(ErrTokenVerificationFailed, err)
	}

	var claims Claims
	if err = idToken.Claims(&claims); err != nil {
		return errors.Join(ErrTokenClaimsFailed, err)
	}

	actualScopes := strings.Split(claims.Scopes, " ")

	if !hasRequiredScopes(actualScopes, v.RequiredScopes) {
		return ErrTokenScopesMismatch
	}

	return nil
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
