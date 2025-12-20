# Knot MCP Test Environment

This directory contains Docker Compose setup for testing the Knot MCP server with a PostgreSQL database.

## Quick Start

1. **Start the test database:**
   ```bash
   cd docker
   docker-compose up -d
   ```

2. **Wait for PostgreSQL to be ready:**
   ```bash
   docker-compose logs -f postgres
   # Wait until you see "database system is ready to accept connections"
   ```

3. **Test the MCP server:**
   ```bash
   # From the project root directory:
   ./knot --postgres-endpoint="postgres://knot_user:knot_password@localhost:5432/knot_test?sslmode=disable" mcp server
   ```

## Connection Details

- **Host:** localhost
- **Port:** 5432
- **Database:** knot_test
- **User:** knot_user
- **Password:** knot_password
- **SSL Mode:** disable (for local testing)

## Environment Variable Alternative

You can also set the connection string as an environment variable:

```bash
export KNOT_POSTGRES_ENDPOINT="postgres://knot_user:knot_password@localhost:5432/knot_test?sslmode=disable"
./knot mcp server
```

## Cleanup

To stop and remove the test database:

```bash
cd docker
docker-compose down -v
```

## Health Check

The PostgreSQL container includes a health check. You can verify it's running:

```bash
docker-compose ps
# Should show "healthy" status for postgres
```