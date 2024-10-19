package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"go.charczuk.com/sdk/configutil"
)

// Config is the config options.
type Config struct {
	// ClientID is part of the oauth credential pair.
	ClientID string `json:"clientID,omitempty" yaml:"clientID,omitempty"`
	// ClientSecret is part of the oauth credential pair.
	ClientSecret string `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`
	// Secret is an encryption key used to verify oauth state.
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty"`
	// RedirectURL is the oauth return url.
	RedirectURL string `json:"redirectURL,omitempty" yaml:"redirectURL,omitempty"`
	// HostedDomain is a specific domain we want to filter identities to.
	HostedDomain string `json:"hostedDomain,omitempty" yaml:"hostedDomain,omitempty"`
	// AllowedDomains is a strict list of hosted domains to allow authenticated users from.
	// If it is unset or empty, it will allow users from *any* hosted domain.
	AllowedDomains []string `json:"allowedDomains,omitempty" yaml:"allowedDomains,omitempty"`
	// Scopes are oauth scopes to request.
	Scopes []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// IsZero returns if the config is set or not.
func (c Config) IsZero() bool {
	return len(c.ClientID) == 0 || len(c.ClientSecret) == 0
}

// Resolve adds extra steps to perform during `configutil.Read(...)`.
func (c *Config) Resolve(ctx context.Context) error {
	return configutil.Resolve(ctx,
		configutil.Set(&c.ClientID, configutil.Env[string]("OAUTH_CLIENT_ID"), configutil.Lazy(&c.ClientID)),
		configutil.Set(&c.ClientSecret, configutil.Env[string]("OAUTH_CLIENT_SECRET"), configutil.Lazy(&c.ClientSecret)),
		configutil.Set(&c.Secret, configutil.Env[string]("OAUTH_SECRET"), configutil.Lazy(&c.Secret), c.GenerateSecret),
		configutil.Set(&c.RedirectURL, configutil.Env[string]("OAUTH_REDIRECT_URL"), configutil.Lazy(&c.RedirectURL)),
		configutil.Set(&c.HostedDomain, configutil.Env[string]("OAUTH_HOSTED_DOMAIN"), configutil.Lazy(&c.HostedDomain)),
	)
}

// ScopesOrDefault returns the scopes or a default set.
func (c Config) ScopesOrDefault() []string {
	if len(c.Scopes) > 0 {
		return c.Scopes
	}
	return DefaultScopes
}

// GenerateSecret generates a secret.
func (c Config) GenerateSecret(_ context.Context) (*string, error) {
	buf := make([]byte, 64)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	value := hex.EncodeToString(buf)
	return &value, nil
}

// DecodeSecret decodes the secret if set from hex encoding.
func (c Config) DecodeSecret() ([]byte, error) {
	if c.Secret != "" {
		decoded, err := hex.DecodeString(c.Secret)
		if err != nil {
			return nil, err
		}
		return decoded, nil
	}
	return nil, nil
}
