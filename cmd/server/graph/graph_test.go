package graph

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/Silicon-Ally/silicon-starter/authn"
	"github.com/Silicon-Ally/silicon-starter/db/sqldb"
	"github.com/Silicon-Ally/silicon-starter/testing/testdb"
	"github.com/Silicon-Ally/silicon-starter/todo"
	"github.com/Silicon-Ally/testpgx"
	"github.com/Silicon-Ally/testpgx/migrate"
	"github.com/bazelbuild/rules_go/go/tools/bazel"

	"go.uber.org/zap/zaptest"
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

var sqlEnv *testpgx.Env

func runTests(m *testing.M) int {
	migrationsPath, err := bazel.Runfile("db/sqldb/migrations")
	if err != nil {
		log.Fatalf("failed to get a path to migrations: %v", err)
	}
	migrator, err := migrate.New(migrationsPath)
	if err != nil {
		log.Fatalf("failed to init migrator: %v", err)
	}
	ctx := context.Background()

	env, err := testpgx.New(ctx, testpgx.WithMigrator(migrator))
	if err != nil {
		log.Fatalf("while creating/getting the test env: %v", err)
	}
	sqlEnv = env
	defer func() {
		err = sqlEnv.TearDown(ctx)
		if err != nil {
			log.Fatalf("while trying to tear down env: %v", err)
		}
	}()
	return m.Run()
}

type testEnv struct {
	resolver *Resolver
	db       DB // Can be testdb or sqldb
}

func (env *testEnv) getFakeDB(t *testing.T) *testdb.DB {
	tdb, ok := env.db.(*testdb.DB)
	if !ok {
		t.Fatalf("fake DB was requested, but DB was of type %T", env.db)
	}
	return tdb
}

type dbType int

const (
	unknownDBType dbType = iota
	fakeDB
	realDB
)

type envOpt func(*envOpts)

func withRealDB() envOpt {
	return func(eo *envOpts) {
		eo.dbType = realDB
	}
}

type envOpts struct {
	dbType dbType
}

func (eo *envOpts) initDB(t *testing.T) DB {
	t.Helper()
	switch eo.dbType {
	case fakeDB:
		return testdb.New()
	case realDB:
		pool := sqlEnv.GetMigratedDB(context.Background(), t)
		tdb, err := sqldb.New(pool)
		if err != nil {
			t.Fatalf("creating sqldb handle: %v", err)
		}
		return tdb
	default:
		t.Fatalf("unknown DB type %d", eo.dbType)
		return nil
	}
}

func setup(t *testing.T, opts ...envOpt) (*Resolver, *testEnv) {
	eOpts := &envOpts{
		dbType: fakeDB,
	}
	for _, opt := range opts {
		opt(eOpts)
	}

	tdb := eOpts.initDB(t)
	logger := zaptest.NewLogger(t)
	env := &testEnv{db: tdb}

	r, err := NewResolver(&ResolverConfig{
		DB:     env.db,
		Logger: logger,
	})
	if err != nil {
		t.Fatalf("failed to init resolver: %v", err)
	}
	env.resolver = r
	t.Cleanup(func() {
		if eOpts.dbType == fakeDB {
			env.getFakeDB(t).CheckAllTransactionsCommitted(t)
		}
	})
	return r, env
}

func createUserForTest(t *testing.T, env *testEnv) (todo.UserID, context.Context) {
	t.Helper()
	userID, err := env.db.CreateUser(env.db.NoTxn(context.Background()), authn.EmailAndPass, "user@example.com", "User", "user@example.com")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}
	ctx := todo.WithUserID(context.Background(), todo.UserID(userID))
	return userID, ctx
}

func noErrDuringSetup(t testing.TB, errs ...error) {
	t.Helper()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("error during setup at index %d: %v", i, err)
		}
	}
}
