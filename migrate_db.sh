#!/bin/sh
#
# Usage: ./migrate_db.sh <local|prod> <migrate args...>
#   ./migrate_db.sh local up
#   ./migrate_db.sh prod down 1
#
# Database URLs come from LOCAL_DB_URL / PROD_DB_URL. These are read from .env
# when it exists, so values in .env win over the surrounding environment; on a
# machine with no .env the environment is used as-is.

set -eu

env_file="$(dirname "$0")/.env"
if [ -f "$env_file" ]; then
  set -a
  . "$env_file"
  set +a
fi

case "${1:-}" in
  "local") DB_URL="${LOCAL_DB_URL:-}";;
  "prod") DB_URL="${PROD_DB_URL:-}";;
  *) echo "usage: $0 <local|prod> <migrate args...>" >&2; exit 1;;
esac

if [ -z "$DB_URL" ]; then
  echo "$0: no URL configured for '$1' - set LOCAL_DB_URL / PROD_DB_URL in .env" >&2
  exit 1
fi

shift
migrate -source file://./db/migrations/ -database "$DB_URL" "$@"
