#!/bin/sh

set -e

echo "run db migration"
/app/migrate -path /app/migrations -database "$GREENLIGHT_DB_DSN" -verbose up

echo "start the app"
exec "$@"