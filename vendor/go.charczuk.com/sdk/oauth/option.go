package oauth

// Option is an option for oauth managers.
type Option func(*Manager) error

// OptConfig sets a manager based on a config.
func OptConfig(cfg Config) Option {
	return func(m *Manager) error {
		secret, err := cfg.DecodeSecret()
		if err != nil {
			return err
		}
		m.oauth2.ClientID = cfg.ClientID
		m.oauth2.ClientSecret = cfg.ClientSecret
		m.oauth2.RedirectURL = cfg.RedirectURL
		m.oauth2.Scopes = cfg.ScopesOrDefault()
		m.Secret = secret
		m.HostedDomain = cfg.HostedDomain
		m.AllowedDomains = cfg.AllowedDomains
		return nil
	}
}

// OptClientID sets the manager cliendID.
func OptClientID(cliendID string) Option {
	return func(m *Manager) error {
		m.oauth2.ClientID = cliendID
		return nil
	}
}

// OptClientSecret sets the manager clientSecret.
func OptClientSecret(clientSecret string) Option {
	return func(m *Manager) error {
		m.oauth2.ClientSecret = clientSecret
		return nil
	}
}

// OptRedirectURL sets the manager redirectURI.
func OptRedirectURL(redirectURL string) Option {
	return func(m *Manager) error {
		m.oauth2.RedirectURL = redirectURL
		return nil
	}
}

// OptScopes sets the manager scopes.
func OptScopes(scopes ...string) Option {
	return func(m *Manager) error {
		m.oauth2.Scopes = scopes
		return nil
	}
}

// OptSecret sets the manager secret.
func OptSecret(secret []byte) Option {
	return func(m *Manager) error {
		m.Secret = secret
		return nil
	}
}

// OptHostedDomain sets the manager hostedDomain.
func OptHostedDomain(hostedDomain string) Option {
	return func(m *Manager) error {
		m.HostedDomain = hostedDomain
		return nil
	}
}

// OptAllowedDomains sets the manager allowedDomains.
func OptAllowedDomains(allowedDomains ...string) Option {
	return func(m *Manager) error {
		m.AllowedDomains = allowedDomains
		return nil
	}
}
