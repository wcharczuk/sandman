package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
)

// State is the oauth state.
type State struct {
	// Token is a plaintext random token.
	Token string
	// SecureToken is the hashed version of the token.
	// If a key is set, it validates that our app created the oauth state.
	SecureToken string
	// RedirectURI is the redirect uri.
	RedirectURI string
	// Extra includes other state you might need to encode.
	Extra map[string]interface{}
}

// DeserializeState deserializes the oauth state.
func DeserializeState(raw string) (state State, err error) {
	var corpus []byte
	corpus, err = base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return
	}
	buffer := bytes.NewBuffer(corpus)
	err = gob.NewDecoder(buffer).Decode(&state)
	return
}

// MustSerializeState serializes a state value but panics if there is an error.
func MustSerializeState(state State) string {
	output, err := SerializeState(state)
	if err != nil {
		panic(err)
	}
	return output
}

// SerializeState serializes the oauth state.
func SerializeState(state State) (output string, err error) {
	buffer := new(bytes.Buffer)
	err = gob.NewEncoder(buffer).Encode(state)
	if err != nil {
		return
	}
	output = base64.RawURLEncoding.EncodeToString(buffer.Bytes())
	return
}

// StateOption is an option for state objects
type StateOption func(*State)

// OptStateSecureToken sets the secure token on the state.
func OptStateSecureToken(secureToken string) StateOption {
	return func(s *State) {
		s.SecureToken = secureToken
	}
}

// OptStateRedirectURI sets the redirect uri on the stae.
func OptStateRedirectURI(redirectURI string) StateOption {
	return func(s *State) {
		s.RedirectURI = redirectURI
	}
}

// OptStateExtra sets the redirect uri on the stae.
func OptStateExtra(key string, value interface{}) StateOption {
	return func(s *State) {
		if s.Extra == nil {
			s.Extra = make(map[string]interface{})
		}
		s.Extra[key] = value
	}
}
