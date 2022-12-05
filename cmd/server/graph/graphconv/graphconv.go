// Package graphconv handles translation between our API/wire layer (i.e.
// `model` package) and our domain types layer (i.e. `todo` package).
package graphconv

import (
	"fmt"

	"github.com/Silicon-Ally/silicon-starter/cmd/server/model"
	"github.com/Silicon-Ally/silicon-starter/todo"
)

func TagsToGQL(in todo.Tags) []*string {
	out := make([]*string, len(in))
	for i, t := range in {
		tt := t
		out[i] = &tt
	}
	return out
}

func TaskToGQL(tsk *todo.Task) (*model.Task, error) {
	if tsk == nil {
		return nil, nil
	}

	return &model.Task{
		ID:   string(tsk.ID),
		Name: tsk.Name,
		Body: tsk.Body,
		Tags: TagsToGQL(tsk.Tags),
	}, nil
}

func TasksToGQL(tsks []*todo.Task) ([]*model.Task, error) {
	return sliceToGQLWithErrHandling(tsks, TaskToGQL)
}

func UserToGQL(user *todo.User) *model.User {
	if user == nil {
		return nil
	}

	return &model.User{
		ID:   string(user.ID),
		Name: user.Name,
	}
}

func sliceToGQLWithErrHandling[I any, O any](is []I, fn func(I) (O, error)) ([]O, error) {
	out := make([]O, len(is))
	for index, i := range is {
		o, err := fn(i)
		if err != nil {
			return nil, fmt.Errorf("error in GQL slice conversion at index %d: %w", index, err)
		}
		out[index] = o
	}
	return out, nil
}
