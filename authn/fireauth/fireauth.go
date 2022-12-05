// Package fireauth is a wrapper around Firebase auth that works with our
// session cookie management. See
// https://pkg.go.dev/firebase.google.com/go/v4/auth for more info.
package fireauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/Silicon-Ally/silicon-starter/authn"
)

type Client struct {
	client *firebaseauth.Client
}

func New(client *firebaseauth.Client) *Client {
	return &Client{client: client}
}

func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (*authn.Token, error) {
	fbToken, err := c.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	return toAuthToken(fbToken)
}

func toAuthToken(fbToken *firebaseauth.Token) (*authn.Token, error) {
	authTime, err := parseAuthTime(fbToken.Claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth time: %w", err)
	}

	userInfo, err := userInfoFromToken(fbToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info from token: %w", err)
	}

	return &authn.Token{
		UserInfo: userInfo,
		AuthTime: authTime,
	}, nil
}

func parseAuthTime(claims map[string]interface{}) (time.Time, error) {
	authTimeVal, ok := claims["auth_time"]
	if !ok {
		return time.Time{}, errors.New("'auth_time' was not set")
	}

	authTimeUnix, ok := authTimeVal.(float64)
	if !ok {
		return time.Time{}, fmt.Errorf("'auth_time' claim was of wrong type %T", authTimeVal)
	}

	return time.Unix(int64(authTimeUnix), 0), nil
}

func userInfoFromToken(tkn *firebaseauth.Token) (*authn.UserInfo, error) {
	if tkn == nil {
		return nil, errors.New("no token provided")
	}

	provider, ok := toAuthProvider(tkn.Firebase.SignInProvider)
	if !ok {
		return nil, fmt.Errorf("no auth provider found for %q", tkn.Firebase.SignInProvider)
	}

	emailVal, ok := tkn.Firebase.Identities["email"]
	if !ok {
		return nil, errors.New("no emails found in firebase identities")
	}

	emails, ok := emailVal.([]interface{})
	if !ok {
		return nil, fmt.Errorf("email identities of type %T, expected []interface", emailVal)
	}
	if len(emails) != 1 {
		return nil, fmt.Errorf("got %d emails, wanted one", len(emails))
	}

	email, ok := emails[0].(string)
	if !ok {
		return nil, fmt.Errorf("email identity of type %T, expected string", emails[0])
	}

	return &authn.UserInfo{
		UserID:       authn.UserID(tkn.UID),
		Email:        email,
		AuthProvider: provider,
	}, nil
}

// These are the only Authn Providers the starter project supports, but there
// is nothing to prevent you from supporting additional providers here, ex, SMS.
func toAuthProvider(provider string) (authn.Provider, bool) {
	switch provider {
	case "google.com":
		// See https://firebase.google.com/docs/reference/js/auth.googleauthprovider
		return authn.Google, true
	case "password":
		// See https://firebase.google.com/docs/reference/js/auth.emailauthprovider#emailauthprovideremail_password_sign_in_method
		return authn.EmailAndPass, true
	case "facebook.com":
		// See https://firebase.google.com/docs/reference/js/auth.facebookauthprovider
		return authn.Facebook, true
	default:
		return authn.UnknownProvider, false
	}
}

func (c *Client) VerifySessionCookie(ctx context.Context, sessionCookie string) (*authn.Token, error) {
	fbToken, err := c.client.VerifySessionCookie(ctx, sessionCookie)
	if err != nil {
		return nil, fmt.Errorf("failed to verify session cookie: %w", err)
	}

	return toAuthToken(fbToken)
}

func (c *Client) SessionCookie(ctx context.Context, idToken string, expiresIn time.Duration) (string, error) {
	return c.client.SessionCookie(ctx, idToken, expiresIn)
}

func (c *Client) RevokeRefreshTokens(ctx context.Context, uID authn.UserID) error {
	return c.client.RevokeRefreshTokens(ctx, string(uID))
}
