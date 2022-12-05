// Package db provides generalized utilities for interacting with some
// database, whether its an in-memory mock, a live SQL database, or something
// else.
package db

import (
	"errors"
	"fmt"

	"github.com/Silicon-Ally/silicon-starter/todo"
)

type errNotFound struct {
	// id is the ID that wasn't found.
	id string
	// entityType is the type of the entity that the caller was looking for.
	entityType string
}

func (e *errNotFound) Error() string {
	return fmt.Sprintf("entity of type %q with ID %q was not found", e.entityType, e.id)
}

func NotFound[T ~string](id T, entityType string) error {
	return &errNotFound{id: string(id), entityType: entityType}
}

func (e *errNotFound) Is(target error) bool {
	_, ok := target.(*errNotFound)
	return ok
}

func IsNotFound(err error) bool {
	return errors.Is(err, &errNotFound{})
}

type Tx interface {
	Commit() error
	Rollback() error
}

type UpdateUserFn func(*todo.User) error

func SetUserName(value string) UpdateUserFn {
	return func(u *todo.User) error {
		u.Name = value
		return nil
	}
}

func SetUserEmail(value string) UpdateUserFn {
	return func(u *todo.User) error {
		u.Email = value
		return nil
	}
}

type UpdateTaskFn func(*todo.Task) error

func SetTaskName(value string) UpdateTaskFn {
	return func(t *todo.Task) error {
		t.Name = value
		return nil
	}
}

func SetTaskBody(value string) UpdateTaskFn {
	return func(t *todo.Task) error {
		t.Body = value
		return nil
	}
}

func AddTaskTag(value string) UpdateTaskFn {
	return func(tsk *todo.Task) error {
		tsk.Tags = tsk.Tags.Add(value)
		return nil
	}
}

func RemoveTaskTag(value string) UpdateTaskFn {
	return func(p *todo.Task) error {
		p.Tags = p.Tags.Remove(value)
		return nil
	}
}
