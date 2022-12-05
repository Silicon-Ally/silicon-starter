package graph

import (
	"context"
	"testing"

	"github.com/Silicon-Ally/silicon-starter/cmd/server/model"
	"github.com/google/go-cmp/cmp"
)

func TestSetUserName(t *testing.T) {
	r, env := setup(t)
	testSetUserName(t, r, env)
}

func TestSetUserNameRealDB(t *testing.T) {
	r, env := setup(t, withRealDB())
	testSetUserName(t, r, env)
}

func testSetUserName(t *testing.T, r *Resolver, env *testEnv) {
	userID, ctx := createUserForTest(t, env)
	name := "Inego Montoya"

	anonCtx := context.Background()
	_, err := r.Mutation().SetUserName(anonCtx, name)
	if err == nil {
		t.Fatalf("expected an error when setting user name with anonymous context, but got none")
	}

	_, err = r.Mutation().SetUserName(ctx, name)
	if err != nil {
		t.Fatalf("expected no error when setting user name with logged-in context, but got %v", err)
	}

	actual, err := r.Query().Me(ctx)
	if err != nil {
		t.Fatalf("reading me: %v", err)
	}
	expected := &model.User{
		ID:   string(userID),
		Name: name,
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected diff (-want +got):\n %s", diff)
	}
}
