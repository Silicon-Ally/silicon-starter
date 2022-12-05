# Database

This project uses [PostgreSQL](https://www.postgresql.org/) for its
database, and almost all of the code for it lives in `/db/sqldb`.

## Development

You can start up the database for local development by running:

```
bazel run //scripts:run_db 
```

This will set it up in such a way that it can recieve
connections from other parts of the stack running locally.

To run tests in this package (which will spin up their own databases,
no need to have the DB running), run:

```
bazel test //db/...
```

## About

### Migrations

In order to support safe versioning of your schema, we use migrations
to describe any and all mutations to your database schema. You can read
more about this approach in the documentation for `testpgx/migrations`.
The long and short of it is simple: any time you want to alter your
database's schema, create two files, one describing the transaction that
you want to apply, and the second describing how you would roll it back. 

The naming convention is:

```
[Monotonically Increasing Number]_[brief description].[up|down].sql
```

We've written a migration test will validate that your rollup and rollback
are true inverses of one another, by comparing the initial database state
to the data base state if the rollup and the rollback mutations are done in
sequence.

Once you check in a mutation, do not alter it! Instead, create a novel 
mutation to accomplish your edit. This approach allows for robust data
handling and data migrations that can have thorough integration tests
and backward compatability tests!

Note we've provided two initial migrations to demonstrate this. Feel free
to delete those two if you do not need them, but do not change the first 
migration (`create_schema_migrations_history`)!

### Goldens

Two golden files are expected to be checked in alongside your code in
the `golden` repo: a dump of your schema, and a human readable version
of your schema. These are purely for code review purposes - it allows
a reviewer to know what mutations your migrations are proposing, and
validates that the migrations have in fact passed the basic tests for
stability and sanity embedded in the golden regeneration tests. You can
run these at any time (idempotently) with:

```
bazel run //scripts:regen_db_goldens
```

### Testing

Testing is demonstrated in the `_test.go` files. Clean versions of
the database are stood up to run each test, and since you're testing
against a local version of postgres, these tests are really 
integration tests, and you can use them to not only validate your
business logic, but how you anticipate postgres responding to various
situations (foreign key constraint violations, etc).

## Hosting on the cloud

Like all other cloud resources, the instance itself is managed by
[our Terraform configuration](/terraform/modules/database). For more information,
read [the initial setup guide](/INITIAL_SETUP.md), specifically the 'Create the
dev database' section.

To connect to our dev Cloud SQL Postgres database, run:

```bash
# This script assumes you've populated secrets/<env>.enc.json with the
# credentials needed to connect to the database and configured
# scripts/shared/bastion.sh with environment-specific details.
bazel run //scripts:cloudsql_shell
```

This will spin up a [bastion host](https://en.wikipedia.org/wiki/Bastion_host),
[tunnel through to the DB using
IAP](https://cloud.google.com/iap/docs/using-tcp-forwarding), and connect a
Postgres shell.

## Migrating a deployed SQL database

To migrate our dev Cloud SQL Postgres database, run:

```bash
# This script assumes you've populated secrets/<env>.enc.json with the
# credentials needed to connect to the database and configured
# scripts/shared/bastion.sh with environment-specific details.
bazel run //scripts:migrate_cloudsql
```

This will do a similar process as connecting above, but instead of opening a
shell, it'll just run the [`migratesqldb` tool](/cmd/tools/migratesqldb)
against the DB.
