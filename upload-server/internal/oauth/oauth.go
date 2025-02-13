package oauth

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
)

type Claims struct {
	Expiry float64 `json:"exp"`
	Scopes string  `json:"scope"`
}

type OAuthValidator struct {
	OAuthEnabled   bool
	IssuerUrl      string
	RequiredScopes string
}

func (v OAuthValidator) ValidateJWT(ctx context.Context, token string) error {
	provider, err := oidc.NewProvider(ctx, v.IssuerUrl)
	if err != nil {
		return err
	}

	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	return nil
}
