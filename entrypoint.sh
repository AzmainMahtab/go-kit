#!/bin/sh
set -e

# Environment-aware entrypoint for Go Kit API.
# - Production: runs compiled binary.
# - Development: runs Air hot-reload (intended for docker-compose.override.yml).

ENVIRONMENT="${ENVIRONMENT:-production}"
POSTGRES_HOST="${POSTGRES_HOST:-postgres}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-app}"
POSTGRES_DB="${POSTGRES_DB:-appdb}"
POSTGRES_DSN="${POSTGRES_DSN:-}"

echo ">>> Environment: ${ENVIRONMENT}"

# Build DSN for pg_isready if no explicit DSN is provided.
if [ -z "$POSTGRES_DSN" ]; then
  POSTGRES_DSN="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE:-disable}"
fi

echo ">>> Waiting for PostgreSQL at ${POSTGRES_HOST}:${POSTGRES_PORT}..."
until pg_isready -d "${POSTGRES_DSN}" -q; do
  echo ">>> Postgres is unavailable - sleeping 1s"
  sleep 1
done
echo ">>> PostgreSQL is ready!"

echo ">>> Running Goose migrations..."
goose -dir /migrations postgres "${POSTGRES_DSN}" up
echo ">>> Migrations complete!"

if [ "$ENVIRONMENT" = "development" ] || [ "$ENVIRONMENT" = "local" ]; then
  echo ">>> Starting development server (air hot-reload)..."
  exec air -c /app/.air.toml
else
  echo ">>> Starting production server..."
  exec api
fi
