#!/bin/bash
set -euo pipefail

function prompt_for_db_env {
  if [ -z "$1" ]; then
    echo 'No db env variable specified'
    exit 1
  fi

  read -p 'Environment (e.g. dev or prod): ' "$1"

  case "$DB_ENV" in
    dev | prod)
      echo "Connecting to ${DB_ENV}..."
      ;;

    *)
      echo "Unknown environment ${DB_ENV}"
      exit 1
      ;;
  esac
}
