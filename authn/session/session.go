// Package session provides functionality for providing auth based on session cookies.
package session

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Silicon-Ally/silicon-starter/authn"
	"github.com/Silicon-Ally/silicon-starter/db"
	"github.com/Silicon-Ally/silicon-starter/todo"
	"go.uber.org/zap"
)

// Auth represents a backing auth system that can handle token verification and
// session management. It's usually represented by a *fireauth.Client.
type Auth interface {
	VerifyIDToken(ctx context.Context, idToken string) (*authn.Token, error)
	SessionCookie(ctx context.Context, idToken string, expiresIn time.Duration) (string, error)
	VerifySessionCookie(ctx context.Context, sessionCookie string) (*authn.Token, error)
	RevokeRefreshTokens(ctx context.Context, uID authn.UserID) error
}

// DB represents a storage system for storing information about users, and creating
// a use-case specific UserID, rather than using the Authorization system's UserID.
// This storage system is responsible only to have transactional semantics and the
// ability to create a user, and retrieve back that same user when requested via the
// same input Authorization IDs.
type DB interface {
	Transactional(context.Context, func(tx db.Tx) error) error
	NoTxn(context.Context) db.Tx
	UserByAuthnProvider(tx db.Tx, provider authn.Provider, userID authn.UserID) (*todo.User, error)
	CreateUser(tx db.Tx, provider authn.Provider, authID authn.UserID, name, email string) (todo.UserID, error)
}

type Client struct {
	auth   Auth
	db     DB
	logger *zap.Logger
	since  func(time.Time) time.Duration // Stubbed out for deterministic tests
}

func New(auth Auth, db DB, logger *zap.Logger) *Client {
	return &Client{
		auth:   auth,
		db:     db,
		logger: logger,
		since: func(t time.Time) time.Duration {
			return time.Since(t)
		},
	}
}

func (c *Client) LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			c.logger.Warn("session login request had invalid HTTP method - only POST is supported", zap.String("http_method", r.Method))
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		req, err := parseLoginRequest(r.Body)
		if err != nil {
			c.logger.Warn("failed to decode session login request", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		tkn, err := c.auth.VerifyIDToken(r.Context(), req.IDToken)
		if err != nil {
			c.logger.Warn("failed to verify ID token", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		ui := tkn.UserInfo

		// Return error if the sign-in is older than 5 minutes.
		signInAge := c.since(tkn.AuthTime)
		if signInAge > 5*time.Minute {
			c.logger.Warn("sign in was older than five minutes ago", zap.Duration("sign-in age", signInAge))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// Set session expiration to 14 days
		expiresIn := time.Hour * 24 * 14

		err = c.db.Transactional(r.Context(), func(tx db.Tx) error {
			_, err := c.db.UserByAuthnProvider(tx, tkn.UserInfo.AuthProvider, tkn.UserInfo.UserID)
			if db.IsNotFound(err) {
				// Since the user isn't found, create the account.
				_, err := c.db.CreateUser(tx, tkn.UserInfo.AuthProvider, tkn.UserInfo.UserID, req.Name, tkn.UserInfo.Email)
				if err != nil {
					return fmt.Errorf("failed to create user (provider id %q): %w", tkn.UserInfo.UserID, err)
				}
			} else if err != nil {
				return fmt.Errorf("failed to get or create user: %w", err)
			}
			return nil
		})
		if err != nil {
			c.logger.Error("failed to create/retrieve user id",
				zap.String("user_id", string(ui.UserID)),
				zap.String("auth_provider", string(ui.AuthProvider)),
				zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Create the session cookie, which will have the same claims as the ID token.
		cookie, err := c.auth.SessionCookie(r.Context(), req.IDToken, expiresIn)
		if err != nil {
			// This one is an error, because we've already validated the token.
			c.logger.Error("failed to create a session cookie", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var uiBuf bytes.Buffer
		if err := json.NewEncoder(&uiBuf).Encode(tkn.UserInfo); err != nil {
			c.logger.Error("failed to encode user info", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		// Set cookie policy for session cookie.
		http.SetCookie(w, &http.Cookie{
			// For why we have to choose this specific name, check out
			// https://firebase.google.com/docs/hosting/manage-cache#using_cookies
			Name:     "__session",
			Path:     "/",
			Value:    cookie,
			MaxAge:   int(expiresIn.Seconds()),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		if _, err := io.Copy(w, &uiBuf); err != nil {
			c.logger.Error("failed to copy JSON body to output", zap.Error(err))
		}
	})
}

func (c *Client) LogoutHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			c.logger.Warn("session logout request had invalid HTTP method - required POST", zap.String("http_method", r.Method))
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "__session",
			Path:     "/",
			Value:    "",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		// If anything below fails, we don't want to fail the request, just log it.
		// We've done our main goal of erasing the user's session-related cookies.
		userInfo, ok := r.Context().Value(userInfoKey{}).(*authn.UserInfo)
		if !ok {
			c.logger.Error("no user info found in context")
			return
		}

		if err := c.auth.RevokeRefreshTokens(r.Context(), userInfo.UserID); err != nil {
			c.logger.Error("failed to revoke firebase refresh tokens",
				zap.String("user_id", string(userInfo.UserID)),
				zap.Error(err))
			return
		}
	})
}

func (c *Client) WithAuthorization(next http.Handler, loginHandlerPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The only endpoint we don't require auth on is session login, which
		// makes sense because that's how the user gets a valid session cookie in
		// the first place.
		if r.URL.Path == loginHandlerPath {
			next.ServeHTTP(w, r)
			return
		}

		// If we're here, we require standard session cookie-based user auth.
		sessionCookie, err := extractSessionCookieFromRequest(r)
		if err != nil {
			c.logger.Warn("request had invalid session cookie", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		token, err := c.auth.VerifySessionCookie(r.Context(), sessionCookie)
		if err != nil {
			c.logger.Warn("request had invalid ID token", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userInfo := token.UserInfo

		c.logger.Debug(
			"verified session cookie",
			zap.String("user_id", string(userInfo.UserID)),
			zap.String("auth_provider", string(userInfo.AuthProvider)),
		)

		ctx := context.WithValue(r.Context(), userInfoKey{}, userInfo)

		user, err := c.db.UserByAuthnProvider(c.db.NoTxn(ctx), userInfo.AuthProvider, userInfo.UserID)
		if err != nil {
			c.logger.Error("failed to load user by auth provider, user had valid session cookie",
				zap.Error(err),
				zap.String("user_id", string(userInfo.UserID)),
				zap.String("auth_provider", string(userInfo.AuthProvider)))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		ctx = todo.WithUserID(ctx, user.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserInfoFromContext(ctx context.Context) (*authn.UserInfo, bool) {
	ui, ok := ctx.Value(userInfoKey{}).(*authn.UserInfo)
	if !ok {
		return nil, false
	}
	return ui, true
}

type userInfoKey struct{}

func extractSessionCookieFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie("__session")
	if err != nil {
		return "", errors.New("session cookie was not found")
	}

	if cookie.Value == "" {
		return "", errors.New("session cookie was empty")
	}

	return cookie.Value, nil
}

// LoginRequest represents the format we expect to receive for session login
// requests, usually in the JSON-formatted body of a POST request.
type LoginRequest struct {
	Name      string `json:"name"`
	IDToken   string `json:"idToken"`
	CSRFToken string `json:"csrfToken"`
}

func parseLoginRequest(r io.Reader) (*LoginRequest, error) {
	var req LoginRequest
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to parse login request: %w", err)
	}
	if req.IDToken == "" {
		return nil, errors.New("no ID token was provided in the request")
	}
	return &req, nil
}
