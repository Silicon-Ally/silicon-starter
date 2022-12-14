package sqldb

import (
	"context"
	"log"
	"math/rand"
	"os"
	"testing"

	"github.com/Silicon-Ally/idgen"
	"github.com/Silicon-Ally/testpgx"
	"github.com/Silicon-Ally/testpgx/migrate"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

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

	testEnv, err := testpgx.New(ctx, testpgx.WithMigrator(migrator))
	if err != nil {
		log.Fatalf(" while creating/getting the test env: %v", err)
	}
	defer func() {
		err = testEnv.TearDown(ctx)
		if err != nil {
			log.Fatalf("while trying to tear down env: %v", err)
		}
	}()
	env = testEnv
	result := m.Run()
	return result
}

var env *testpgx.Env

func noErrDuringSetup(t testing.TB, errs ...error) {
	t.Helper()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("error during setup at index %d: %v", i, err)
		}
	}
}

func TestSchemaHistory(t *testing.T) {
	ctx := context.Background()
	db := env.GetMigratedDB(ctx, t)

	q := `SELECT id, version FROM schema_migrations_history ORDER BY id`
	rows, err := db.Query(ctx, q)
	if err != nil {
		t.Fatalf("failed to query schema migrations history: %v", err)
	}

	type versionHistory struct {
		ID      int
		Version int
	}

	var got []versionHistory
	for rows.Next() {
		var vh versionHistory
		if err := rows.Scan(&vh.ID, &vh.Version); err != nil {
			t.Fatalf("failed to load version history entry: %v", err)
		}
		got = append(got, vh)
	}

	want := []versionHistory{
		{ID: 1, Version: 1}, // 0001_create_schema_migrations_history
		{ID: 2, Version: 2}, // 0002_create_user_table
		{ID: 3, Version: 3}, // 0003_create_todo_table
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected schema version history (-want +got)\n%s", diff)
	}
}

func createDBForTesting(t *testing.T) *DB {
	r := rand.New(rand.NewSource(0))
	idg, err := idgen.New(r, idgen.WithCharSet([]rune("abcdefhijklmnopqrstuvwxyz")))
	if err != nil {
		t.Fatalf("creating id generator: %v", err)
	}
	ctx := context.Background()
	pool := env.GetMigratedDB(ctx, t)
	return &DB{
		db:          pool,
		idGenerator: idg,
	}
}
