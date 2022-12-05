// Package testdb implements an in-memory database for use in tests.
package testdb

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Silicon-Ally/silicon-starter/authn"
	"github.com/Silicon-Ally/silicon-starter/db"
	"github.com/Silicon-Ally/silicon-starter/todo"
)

type DB struct {
	users []*todo.User
	tasks []*todo.Task

	pendingTxns map[*Op]bool
	nextIDs     map[string]int
}

func New() *DB {
	return &DB{
		nextIDs:     make(map[string]int),
		pendingTxns: make(map[*Op]bool),
	}
}

func (db *DB) Begin(_ context.Context) (db.Tx, error) {
	tx := &Op{db: db}
	db.pendingTxns[tx] = true
	return &Op{}, nil
}

func (db *DB) CheckAllTransactionsCommitted(t *testing.T) {
	if len(db.pendingTxns) > 0 {
		t.Fatalf("had %d txns that were begun but not committed", len(db.pendingTxns))
	}
}

func (tdb *DB) RunOrContinueTransaction(tx db.Tx, fn func(tx db.Tx) error) error {
	return fn(tx)
}

func (db *DB) NoTxn(_ context.Context) db.Tx {
	return &Op{}
}

func (db *DB) nextID(ns string) string {
	idx := db.nextIDs[ns]
	db.nextIDs[ns]++
	return fmt.Sprintf("%s.%d", ns, idx)
}

type Op struct {
	db *DB
}

func (tx *Op) Commit() error {
	_, ok := tx.db.pendingTxns[tx]
	if !ok {
		return errors.New("attempting to commit txn prior to beginning it")
	}
	delete(tx.db.pendingTxns, tx)
	return nil
}

func (tx *Op) Rollback() error {
	return nil
}

func (db *DB) Transactional(_ context.Context, fn func(_ db.Tx) error) error {
	if err := fn(nil); err != nil {
		return fmt.Errorf("while running testdb txn: %w", err)
	}
	return nil
}

func (tdb *DB) UserByAuthnProvider(tx db.Tx, authProvider authn.Provider, authID authn.UserID) (*todo.User, error) {
	for _, u := range tdb.users {
		if u.AuthnProviderType == authProvider && u.AuthnProviderID == authID {
			return u.Clone(), nil
		}
	}

	return nil, db.NotFound(string(authProvider)+":"+string(authID), "User")
}

func (tdb *DB) CreateUser(tx db.Tx, provider authn.Provider, authID authn.UserID, name string, email string) (todo.UserID, error) {
	id := todo.UserID(tdb.nextID("user"))
	u := &todo.User{
		ID:                id,
		Name:              name,
		Email:             email,
		AuthnProviderType: provider,
		AuthnProviderID:   authID,
	}
	tdb.users = append(tdb.users, u)
	return id, nil
}

func (tdb *DB) UpdateUser(_ db.Tx, id todo.UserID, ms ...db.UpdateUserFn) error {
	for i, u := range tdb.users {
		if u.ID == id {
			uu := u.Clone()
			for _, m := range ms {
				m(uu)
			}
			tdb.users[i] = uu
			return nil
		}
	}
	return db.NotFound(id, "user")
}

func (tdb *DB) User(_ db.Tx, id todo.UserID) (*todo.User, error) {
	for _, u := range tdb.users {
		if u.ID == id {
			return u.Clone(), nil
		}
	}
	return nil, db.NotFound(id, "user")
}

func (tdb *DB) Users(_ db.Tx) ([]*todo.User, error) {
	var r []*todo.User
	for _, u := range tdb.users {
		r = append(r, u.Clone())
	}
	return r, nil
}

func (tdb *DB) Task(_ db.Tx, id todo.TaskID) (*todo.Task, error) {
	for _, t := range tdb.tasks {
		if t.ID == id {
			return t.Clone(), nil
		}
	}
	return nil, db.NotFound(id, "task")
}

func (tdb *DB) TasksByCreator(_ db.Tx, userID todo.UserID) ([]*todo.Task, error) {
	r := make([]*todo.Task, 0)
	for _, t := range tdb.tasks {
		if t.CreatedBy == userID {
			r = append(r, t.Clone())
		}
	}
	return r, nil
}

func (tdb *DB) CreateTask(_ db.Tx, userID todo.UserID) (todo.TaskID, error) {
	t := &todo.Task{
		ID:        todo.TaskID(tdb.nextID("task")),
		CreatedBy: userID,
	}
	tdb.tasks = append(tdb.tasks, t)
	return t.ID, nil
}

func (tdb *DB) UpdateTask(_ db.Tx, id todo.TaskID, ms ...db.UpdateTaskFn) error {
	for i, t := range tdb.tasks {
		if t.ID == id {
			t := t.Clone()
			for _, m := range ms {
				m(t)
			}
			tdb.tasks[i] = t
			return nil
		}
	}
	return db.NotFound(id, "task")
}

func (tdb *DB) DeleteTask(_ db.Tx, id todo.TaskID) error {
	for i, t := range tdb.tasks {
		if t.ID == id {
			tdb.tasks = append(tdb.tasks[:i], tdb.tasks[i+1:]...)
			return nil
		}
	}
	return db.NotFound(id, "task")
}
