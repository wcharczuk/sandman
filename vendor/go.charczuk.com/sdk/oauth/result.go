package oauth

import "time"

// Result is the final result of the oauth exchange.
// It is the user profile of the user and the state information.
type Result struct {
	Response Response
	Profile  Profile
	State    State
}

// Response is the response details from the oauth exchange.
type Response struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
	HostedDomain string
}
