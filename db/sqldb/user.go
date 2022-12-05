package sqldb

import (
	"fmt"
	"time"

	"github.com/Silicon-Ally/silicon-starter/authn"
	"github.com/Silicon-Ally/silicon-starter/db"
	"github.com/Silicon-Ally/silicon-starter/todo"
	"github.com/jackc/pgx/v4"
)

func (db *DB) User(tx db.Tx, id todo.UserID) (*todo.User, error) {
	row := db.queryRow(tx, `
		SELECT 
			id, name, email, created_at, auth_provider_type, auth_provider_id	
		FROM user_account
		WHERE id = $1;
		`, id)
	user, err := rowToUser(row)
	if err != nil {
		return nil, fmt.Errorf("reading user: %w", err)
	}
	return user, nil
}

func (d *DB) UserByAuthnProvider(tx db.Tx, authnProvider authn.Provider, authnProvidedUserID authn.UserID) (*todo.User, error) {
	rows, err := d.query(tx, `
		SELECT 
			id, name, email, created_at, auth_provider_type, auth_provider_id	
		FROM user_account
		WHERE auth_provider_type = $1 AND auth_provider_id = $2;
		`, authnProvider, authnProvidedUserID)
	if err != nil {
		return nil, fmt.Errorf("reading user by auth: %w", err)
	}
	users, err := rowsToUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("translating rows to users: %w", err)
	}
	if len(users) == 0 {
		return nil, db.NotFound(authnProvidedUserID, "user")
	} else if len(users) == 1 {
		return users[0], nil
	} else {
		return nil, fmt.Errorf("expected exactly one user in result but got %d", len(users))
	}
}

func (db *DB) Users(tx db.Tx) ([]*todo.User, error) {
	rows, err := db.query(tx, `
		SELECT 
			id, name, email, created_at, auth_provider_type, auth_provider_id	
		FROM user_account;`)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	users, err := rowsToUsers(rows)
	if err != nil {
		return nil, fmt.Errorf("reading user: %w", err)
	}
	return users, nil
}

const userIDNamespace = "user"

const defaultUserName = "Unnamed User"

func (db *DB) CreateUser(
	tx db.Tx,
	authProviderType authn.Provider,
	authProviderId authn.UserID,
	name string,
	email string) (todo.UserID, error) {
	id := todo.UserID(db.randomID(userIDNamespace))
	createdAt := time.Now()
	err := db.exec(tx, `
		INSERT INTO user_account
			(id, name, email, created_at, auth_provider_type, auth_provider_id)
			VALUES
			($1, $2, $3, $4, $5, $6);
		`, id, name, email, createdAt, authProviderType, authProviderId)
	if err != nil {
		return "", fmt.Errorf("creating user_account row for %s: %w", id, err)
	}
	return id, nil
}

func (d *DB) UpdateUser(
	tx db.Tx,
	userID todo.UserID,
	userMutations ...db.UpdateUserFn) error {
	err := d.RunOrContinueTransaction(tx, func(tx db.Tx) error {
		user, err := d.User(tx, userID)
		if err != nil {
			return fmt.Errorf("reading user pre-mutations: %w", err)
		}
		for i, m := range userMutations {
			err := m(user)
			if err != nil {
				return fmt.Errorf("running mutation #%d: %w", i, err)
			}
		}
		err = d.putUser(tx, user)
		if err != nil {
			return fmt.Errorf("writing user post-mutations: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("running update user txn: %w", err)
	}
	return nil
}

func (db *DB) putUser(tx db.Tx, user *todo.User) error {
	err := db.exec(tx, `
		UPDATE user_account SET
			name = $2,
			email = $3
		WHERE id = $1;
		`, user.ID, user.Name, user.Email)
	if err != nil {
		return fmt.Errorf("updating user_account writable fields: %w", err)
	}
	return nil
}

func rowToUser(s rowScanner) (*todo.User, error) {
	u := &todo.User{}
	err := s.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
		&u.AuthnProviderType,
		&u.AuthnProviderID)
	if err != nil {
		return nil, fmt.Errorf("scanning into user: %w", err)
	}
	return u, nil
}

func rowsToUsers(rows pgx.Rows) ([]*todo.User, error) {
	defer rows.Close()
	var us []*todo.User
	for rows.Next() {
		u, err := rowToUser(rows)
		if err != nil {
			return nil, fmt.Errorf("converting row to user: %w", err)
		}
		us = append(us, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("while processing user rows: %w", err)
	}
	return us, nil
}
