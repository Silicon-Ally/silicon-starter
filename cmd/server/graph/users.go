package graph

import (
	"context"

	"github.com/Silicon-Ally/gqlerr"
	"github.com/Silicon-Ally/silicon-starter/cmd/server/graph/graphconv"
	"github.com/Silicon-Ally/silicon-starter/cmd/server/model"
	"github.com/Silicon-Ally/silicon-starter/db"
	"github.com/Silicon-Ally/silicon-starter/todo"
	"go.uber.org/zap"
)

func (q *queryResolver) Me(ctx context.Context) (*model.User, error) {
	userID, err := q.userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, err := q.db.User(q.db.NoTxn(ctx), todo.UserID(userID))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't read user", zap.String("user_id", string(userID)), zap.Error(err))
	}
	return graphconv.UserToGQL(user), nil
}

func (m *mutationResolver) SetUserName(ctx context.Context, userName string) (*bool, error) {
	userID, err := m.userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	err = m.db.UpdateUser(m.db.NoTxn(ctx), todo.UserID(userID), db.SetUserName(userName))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't update user", zap.String("user_id", string(userID)), zap.Error(err))
	}
	return emptySuccess()
}
