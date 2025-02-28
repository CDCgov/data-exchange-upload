package oauth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/coreos/go-oidc/v3/oidc"
	"gopkg.in/yaml.v3"
)

var ErrTokenVerificationFailed = errors.New("failed to verify token")
var ErrTokenClaimsFailed = errors.New("failed to parse token claims")
var ErrTokenScopesMismatch = errors.New("one or more required scopes not found")

var Providers map[string]Provider

type Config struct {
	Providers map[string]Provider `yaml:"providers"`
}

type Provider struct {
	Name             string `yaml:"name"`
	IssuerURL        string `yaml:"issuerUrl"`
	AuthorizationURL string `yaml:"authorizationUrl"`
	TokenURL         string `yaml:"tokenUrl"`
	ClientID         string `yaml:"clientId"`
	ClientSecret     string `yaml:"clientSecret"`
}

func (p Provider) LogValue() slog.Value {
	return slog.StringValue(p.Name)
}

func UnmarshalOAuthConfig(body string) (*Config, error) {
	confStr := os.ExpandEnv(body)
	c := &Config{}

	err := yaml.Unmarshal([]byte(confStr), &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type Claims struct {
	Expiry int64  `json:"exp"`
	Scopes string `json:"scope"`
}

type Validator interface {
	health.Checkable
	ValidateJWT(ctx context.Context, token string) (Claims, error)
}

type PassthroughValidator struct{}

func (v PassthroughValidator) ValidateJWT(_ context.Context, _ string) (Claims, error) {
	return Claims{}, nil
}
func (v PassthroughValidator) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "no-op oauth validator"
	rsp.Status = models.STATUS_UP

	return rsp
}

func NewOAuthValidator(ctx context.Context, issuerUrl string, requiredScopes string) (*OAuthValidator, error) {
	var scopes []string
	if requiredScopes != "" {
		scopes = strings.Split(requiredScopes, " ")
	}

	p, err := oidc.NewProvider(ctx, issuerUrl)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			slog.Error("failed to reach oidc provider " + issuerUrl)
		} else {
			return nil, err
		}
	}

	return &OAuthValidator{
		IssuerUrl:      issuerUrl,
		RequiredScopes: scopes,
		provider:       p,
	}, nil
}

type OAuthValidator struct {
	IssuerUrl      string
	RequiredScopes []string
	provider       *oidc.Provider
}

func (v *OAuthValidator) ValidateJWT(ctx context.Context, token string) (Claims, error) {
	if v.provider == nil {
		p, err := oidc.NewProvider(ctx, v.IssuerUrl)
		if err != nil {
			return Claims{}, err
		}
		v.provider = p
	}
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

func (v *OAuthValidator) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "oauth validator " + v.IssuerUrl
	rsp.Status = models.STATUS_UP

	wellKnown := strings.TrimSuffix(v.IssuerUrl, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequest("GET", wellKnown, nil)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if resp.StatusCode != http.StatusOK {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "well-known response status " + resp.Status
		return rsp
	}

	return rsp
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
