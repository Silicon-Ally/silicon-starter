#!/bin/bash
set -euo pipefail

ROOT="$BUILD_WORKSPACE_DIRECTORY"
cd "$ROOT"

source "./scripts/shared/bastion.sh"
source "./scripts/shared/db_prompt.sh"

if [ $# -lt 1 ]; then
  prompt_for_db_env DB_ENV
else
  DB_ENV="$1"
fi

start_tunnel "$DB_ENV"

if ! [ -x "$(command -v sops)" ]; then
  echo 'Error: sops is not installed.' >&2
  exit 1
fi

DSN="user=postgres"
DSN+=" password=$(sops -d --extract '["postgres"]["password"]' "${ROOT}/secrets/${DB_ENV}.enc.json")"
DSN+=" host=localhost"
DSN+=" port=5433"
DSN+=" dbname=<db name>"

echo "Migrating database..."
bazel run \
  --ui_event_filters=-info,-stdout,-stderr \
  --noshow_progress \
  //cmd/tools/migratesqldb -- --dsn="$DSN" apply
