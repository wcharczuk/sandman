package oauth

// GoogleClaims are extensions to the jwt standard claims for google oauth.
//
// See additional documentation here: https://developers.google.com/identity/sign-in/web/backend-auth
type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	HD            string `json:"hd"`
	Nonce         string `json:"nonce"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	Locale        string `json:"locale"`
	Picture       string `json:"picture"`
	Profile       string `json:"profile"`
}
