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

func (q *queryResolver) Task(ctx context.Context, taskID string) (*model.Task, error) {
	task, err := q.db.Task(q.db.NoTxn(ctx), todo.TaskID(taskID))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't read task", zap.String("task_id", taskID), zap.Error(err))
	}
	return graphconv.TaskToGQL(task)
}

func (q *queryResolver) TasksByCreator(ctx context.Context, userID string) ([]*model.Task, error) {
	tasks, err := q.db.TasksByCreator(q.db.NoTxn(ctx), todo.UserID(userID))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't read tasks by creator", zap.String("user_id", userID), zap.Error(err))
	}
	return graphconv.TasksToGQL(tasks)
}

func (m *mutationResolver) CreateTask(ctx context.Context) (string, error) {
	userID, err := m.userIDFromContext(ctx)
	if err != nil {
		return "", err
	}
	taskID, err := m.db.CreateTask(m.db.NoTxn(ctx), todo.UserID(userID))
	if err != nil {
		return "", gqlerr.Internal(ctx, "couldn't create task", zap.Error(err))
	}
	return string(taskID), nil
}

func (m *mutationResolver) SetTaskName(ctx context.Context, taskID string, taskName string) (*bool, error) {
	err := m.db.UpdateTask(m.db.NoTxn(ctx), todo.TaskID(taskID), db.SetTaskName(taskName))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't update task name", zap.String("task_id", taskID), zap.Error(err))
	}
	return emptySuccess()
}

func (m *mutationResolver) SetTaskBody(ctx context.Context, taskID string, taskBody string) (*bool, error) {
	err := m.db.UpdateTask(m.db.NoTxn(ctx), todo.TaskID(taskID), db.SetTaskBody(taskBody))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't update task body", zap.String("task_id", taskID), zap.Error(err))
	}
	return emptySuccess()
}

func (m *mutationResolver) AddTaskTag(ctx context.Context, taskID string, tag string) (*bool, error) {
	err := m.db.UpdateTask(m.db.NoTxn(ctx), todo.TaskID(taskID), db.AddTaskTag(tag))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't add task tag", zap.String("task_id", taskID), zap.Error(err))
	}
	return emptySuccess()
}

func (m *mutationResolver) RemoveTaskTag(ctx context.Context, taskID string, tag string) (*bool, error) {
	err := m.db.UpdateTask(m.db.NoTxn(ctx), todo.TaskID(taskID), db.RemoveTaskTag(tag))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't remove task tag", zap.String("task_id", taskID), zap.Error(err))
	}
	return emptySuccess()
}

func (m *mutationResolver) DeleteTask(ctx context.Context, taskID string) (*bool, error) {
	err := m.db.DeleteTask(m.db.NoTxn(ctx), todo.TaskID(taskID))
	if err != nil {
		return nil, gqlerr.Internal(ctx, "couldn't delete task", zap.String("task_id", taskID), zap.Error(err))
	}
	return emptySuccess()
}
