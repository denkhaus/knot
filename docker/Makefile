# ./knot --postgres-endpoint="postgres://knot_user:knot_password@localhost:5432/knot_test?sslmode=disable" --mcp-transport-mode=http --mcp-port=9094 mcp server

.PHONY: up down clean status logs test

# Start the test database
up:
	docker-compose up -d

# Stop the test database
down:
	docker-compose down

# Stop and remove all volumes (clean slate)
clean:
	docker-compose down -v

# Show status of services
status:
	docker-compose ps

# Show logs
logs:
	docker-compose logs -f postgres

# Test connection to database
test:
	docker-compose exec postgres psql -U knot_user -d knot_test -c "SELECT * FROM connection_test;"

# Wait for database to be ready
wait:
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker-compose exec -T postgres pg_isready -U knot_user -d knot_test; do \
		echo "PostgreSQL is unavailable - sleeping"; \
		sleep 2; \
	done
	@echo "PostgreSQL is ready!"
