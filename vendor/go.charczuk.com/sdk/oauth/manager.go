package oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"go.charczuk.com/sdk/r2"
	"go.charczuk.com/sdk/uuid"
)

// New returns a new manager mutated by a given set of options.
func New(ctx context.Context, options ...Option) (*Manager, error) {
	oidcProvider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}
	manager := &Manager{
		oauth2: oauth2.Config{
			Endpoint: oidcProvider.Endpoint(),
			Scopes:   DefaultScopes,
		},
	}
	for _, option := range options {
		if err := option(manager); err != nil {
			return nil, err
		}
	}
	if len(manager.Secret) == 0 {
		return nil, ErrSecretRequired
	}
	manager.verifier = oidcProvider.Verifier(&oidc.Config{
		ClientID: manager.oauth2.ClientID,
	})
	return manager, nil
}

// MustNew returns a new manager mutated by a given set of options
// and will panic on error.
func MustNew(ctx context.Context, options ...Option) *Manager {
	m, err := New(ctx, options...)
	if err != nil {
		panic(err)
	}
	return m
}

// Manager is the oauth manager.
type Manager struct {
	Secret         []byte
	HostedDomain   string
	AllowedDomains []string

	oauth2   oauth2.Config
	verifier *oidc.IDTokenVerifier
}

// OAuthURL is the auth url for google with a given clientID.
// This is typically the link that a user will click on to start the auth process.
func (m *Manager) OAuthURL(r *http.Request, stateOptions ...StateOption) (oauthURL string, err error) {
	var state string
	state, err = SerializeState(m.CreateState(stateOptions...))
	if err != nil {
		return
	}
	var opts []oauth2.AuthCodeOption
	if len(m.HostedDomain) > 0 {
		opts = append(opts, oauth2.SetAuthURLParam("hd", m.HostedDomain))
	}
	oauthURL = m.oauth2.AuthCodeURL(state, opts...)
	return
}

// Finish processes the returned code, exchanging for an access token, and fetches the user profile.
func (m *Manager) Finish(r *http.Request) (result *Result, err error) {
	code := r.URL.Query().Get("code")
	if len(code) == 0 {
		err = ErrCodeMissing
		return
	}

	state := r.URL.Query().Get("state")
	result = new(Result)
	if state != "" {
		var deserialized State
		deserialized, err = DeserializeState(state)
		if err != nil {
			return
		}
		result.State = deserialized
	}
	err = m.ValidateState(result.State)
	if err != nil {
		return
	}

	// Handle the exchange code to initiate a transport.
	tok, err := m.oauth2.Exchange(r.Context(), code)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrFailedCodeExchange, err)
		return
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		err = fmt.Errorf("%w: id_token missing", ErrFailedCodeExchange)
		return
	}

	// Parse and verify ID Token payload.
	idToken, err := m.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrFailedCodeExchange, err)
		return
	}

	var claims GoogleClaims
	if err = idToken.Claims(&claims); err != nil {
		err = fmt.Errorf("%w: %v", ErrFailedCodeExchange, err)
		return
	}

	result.Response.AccessToken = tok.AccessToken
	result.Response.TokenType = tok.TokenType
	result.Response.RefreshToken = tok.RefreshToken
	result.Response.Expiry = tok.Expiry

	result.Profile, err = m.FetchProfile(r.Context(), tok.AccessToken)
	if err != nil {
		return
	}
	return
}

// FetchProfile gets a google profile for an access token.
func (m *Manager) FetchProfile(ctx context.Context, accessToken string) (profile Profile, err error) {
	res, err := r2.New("https://www.googleapis.com/oauth2/v1/userinfo",
		r2.OptGet(),
		r2.OptContext(ctx),
		r2.OptQuery("alt", "json"),
		r2.OptHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)),
	).Do()
	if err != nil {
		return
	}
	defer res.Body.Close()
	if code := res.StatusCode; code < 200 || code > 299 {
		err = ErrGoogleResponseStatus
		return
	}
	if err = json.NewDecoder(res.Body).Decode(&profile); err != nil {
		err = fmt.Errorf("%v: %w", ErrProfileJSONUnmarshal, err)
		return
	}
	return
}

// CreateState creates auth state.
func (m *Manager) CreateState(options ...StateOption) (state State) {
	for _, opt := range options {
		opt(&state)
	}
	if len(m.Secret) > 0 && state.Token == "" && state.SecureToken == "" {
		state.Token = uuid.V4().String()
		state.SecureToken = m.hash(state.Token)
	}
	return
}

// --------------------------------------------------------------------------------
// Validation Helpers
// --------------------------------------------------------------------------------

// ValidateState validates oauth state.
func (m *Manager) ValidateState(state State) error {
	if len(m.Secret) > 0 {
		expected := m.hash(state.Token)
		actual := state.SecureToken
		if !hmac.Equal([]byte(expected), []byte(actual)) {
			return ErrInvalidAntiforgeryToken
		}
	}
	return nil
}

// --------------------------------------------------------------------------------
// internal helpers
// --------------------------------------------------------------------------------

func (m *Manager) hash(plaintext string) string {
	return base64.URLEncoding.EncodeToString(m.hmac([]byte(plaintext)))
}

// hmac hashes data with the given key.
func (m *Manager) hmac(plainText []byte) []byte {
	mac := hmac.New(sha512.New, m.Secret)
	_, _ = mac.Write([]byte(plainText))
	return mac.Sum(nil)
}
