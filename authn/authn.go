// Package authn provides some domain types shared across our auth system.
package authn

import (
	"time"
)

// UserID is a unique ID assigned to an authenticated user by our auth system.
type UserID string

// Provider is the underlying auth provider, since our auth system will
// delegate to multiple under the hood.
type Provider string

const (
	UnknownProvider = Provider("")
	Google          = Provider("GOOGLE")
	EmailAndPass    = Provider("EMAIL_AND_PASS")
	Facebook        = Provider("FACEBOOK")
)

// Token is a representation of a user's auth token, which contains
// basic user information and when they authenticated with the system.
type Token struct {
	UserInfo *UserInfo
	AuthTime time.Time
}

// UserInfo contains basic information relevant for authentication.
type UserInfo struct {
	UserID       UserID   `json:"user_id"` // The Auth-provider's UserId
	Email        string   `json:"email"`
	AuthProvider Provider `json:"auth_provider"`
}
