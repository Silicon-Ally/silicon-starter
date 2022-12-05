package graph

import (
	"context"
	"errors"
	"fmt"

	"github.com/Silicon-Ally/gqlerr"
	"github.com/Silicon-Ally/silicon-starter/authn"
	"github.com/Silicon-Ally/silicon-starter/cmd/server/generated"
	"github.com/Silicon-Ally/silicon-starter/db"
	"github.com/Silicon-Ally/silicon-starter/todo"
	"go.uber.org/zap"
)

type DB interface {
	Begin(context.Context) (db.Tx, error)
	NoTxn(context.Context) db.Tx
	Transactional(context.Context, func(tx db.Tx) error) error
	RunOrContinueTransaction(db.Tx, func(tx db.Tx) error) error

	User(db.Tx, todo.UserID) (*todo.User, error)
	Users(db.Tx) ([]*todo.User, error)
	CreateUser(db.Tx, authn.Provider, authn.UserID, string, string) (todo.UserID, error)
	UpdateUser(db.Tx, todo.UserID, ...db.UpdateUserFn) error

	Task(db.Tx, todo.TaskID) (*todo.Task, error)
	TasksByCreator(db.Tx, todo.UserID) ([]*todo.Task, error)
	CreateTask(db.Tx, todo.UserID) (todo.TaskID, error)
	UpdateTask(db.Tx, todo.TaskID, ...db.UpdateTaskFn) error
	DeleteTask(db.Tx, todo.TaskID) error
}

type Resolver struct {
	db     DB
	logger *zap.Logger
}

// These are part of the gqlgen interface, see https://gqlgen.com/
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }
func (r *Resolver) Query() generated.QueryResolver       { return &queryResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
)

type ResolverConfig struct {
	DB     DB
	Logger *zap.Logger
}

func (c *ResolverConfig) validate() error {
	if c.DB == nil {
		return errors.New("no DB was given")
	}

	if c.Logger == nil {
		return errors.New("no logger given")
	}
	return nil
}

// NewResolver returns an initialized Resolver that can handle GraphQL queries.
func NewResolver(cfg *ResolverConfig) (*Resolver, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config given: %w", err)
	}

	return &Resolver{
		db:     cfg.DB,
		logger: cfg.Logger,
	}, nil
}

func emptySuccess() (*bool, error) {
	b := true
	return &b, nil
}

func (r *Resolver) userIDFromContext(ctx context.Context) (todo.UserID, error) {
	userID, err := todo.UserIDFromContext(ctx)
	if err != nil || userID == "" {
		return "", gqlerr.Internal(ctx, "failed to get user id from context", zap.Error(err))
	}
	return userID, nil
}
