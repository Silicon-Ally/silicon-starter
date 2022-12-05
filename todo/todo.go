// Package todo holds the domain types for the Silicon Starter project. You'll likely want to
// rename this package to whatever your domain or project actually is.
package todo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Silicon-Ally/silicon-starter/authn"
)

// This block contains typed IDs used in the ecosystem, each one representing a
// different entity. Unless otherwise noted, an instance of an ID should
// uniquely represent a single entity of that type.
//
// Keep this block sorted alphabetically to minimize merge conflicts.
type (
	TaskID string
	UserID string
)

type Task struct {
	ID        TaskID
	Name      string
	Body      string
	Tags      Tags
	CreatedBy UserID
}

func (t *Task) Clone() *Task {
	if t == nil {
		return nil
	}
	return &Task{
		ID:        t.ID,
		Name:      t.Name,
		Body:      t.Body,
		Tags:      t.Tags.Clone(),
		CreatedBy: t.CreatedBy,
	}
}

type Tags []string

func (ts Tags) Add(tag string) Tags {
	for _, t := range ts {
		if t == tag {
			return ts
		}
	}
	return append(ts, tag)
}

func (tags Tags) Remove(tag string) Tags {
	for i, t := range tags {
		if t == tag {
			return append(tags[0:i], tags[i+1:]...)
		}
	}
	return tags
}

func (in Tags) Clone() Tags {
	o := make(Tags, len(in))
	copy(o, in)
	return o
}

func (tags Tags) ToStored() string {
	return strings.Join(tags, ",")
}

func TagsFromStored(in string) Tags {
	return Tags(strings.Split(in, ","))
}

type User struct {
	ID                UserID
	Name              string
	Email             string
	CreatedAt         time.Time
	AuthnProviderType authn.Provider
	AuthnProviderID   authn.UserID
}

func (u *User) Clone() *User {
	if u == nil {
		return nil
	}

	return &User{
		ID:                u.ID,
		Name:              u.Name,
		Email:             u.Email,
		CreatedAt:         u.CreatedAt,
		AuthnProviderType: u.AuthnProviderType,
		AuthnProviderID:   u.AuthnProviderID,
	}
}

type userIDContextKey struct{}

func WithUserID(ctx context.Context, id UserID) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, id)
}

func UserIDFromContext(ctx context.Context) (UserID, error) {
	u := ctx.Value(userIDContextKey{})
	if u == nil {
		return "", errors.New("tried to request a user_id from an anonymous context - check the user is logged in")
	}
	userID, ok := u.(UserID)
	if !ok {
		return "", fmt.Errorf("user_id was the wrong type in context: %T", u)
	}
	return userID, nil
}
