package cli

import (
	"log/slog"
	"os"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/oauth"
)

func RegisterOAuthProviders(appConfig appconfig.AppConfig) error {
	oauth.Providers = make(map[string]oauth.Provider)

	dat, err := os.ReadFile(appConfig.OAuthConfigFile)
	if err != nil {
		return err
	}
	cfg, err := oauth.UnmarshalOAuthConfig(string(dat))
	if err != nil {
		return err
	}

	for k, p := range cfg.Providers {
		oauth.Providers[k] = p
		slog.Info("registered oauth provider", "provider", p)
	}

	return nil
}
