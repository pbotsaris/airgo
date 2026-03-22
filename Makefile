.PHONY: test test-verbose test-cover test-unit test-integration test-record lint

# Run all tests (unit tests only, skips integration if no fixtures)
test:
	go test ./airtable/... ./retry/... ./utils/...

# Verbose test output
test-verbose:
	go test -v ./airtable/... ./retry/... ./utils/...

# Run tests with coverage report
test-cover:
	go test -cover ./airtable/... ./retry/... ./utils/...

# Run unit tests only (excludes integration tests)
test-unit:
	go test ./airtable/... ./retry/... ./utils/... -skip "TestIntegration"

# Run integration tests only (replays from fixtures)
test-integration:
	go test -v ./airtable/... -run "TestIntegration"

# Record new integration test fixtures (loads from .env)
# Usage: make test-record
test-record:
	@if [ -f .env ]; then \
		set -a && . ./.env && set +a && \
		AIRTABLE_RECORD=1 go test -v ./airtable/... -run "TestIntegration"; \
	elif [ -n "$(AIRTABLE_KEY)" ]; then \
		AIRTABLE_RECORD=1 AIRTABLE_KEY=$(AIRTABLE_KEY) go test -v ./airtable/... -run "TestIntegration"; \
	else \
		echo "Error: AIRTABLE_KEY is required (set in .env or pass as argument)"; \
		echo "Usage: make test-record"; \
		echo "   or: make test-record AIRTABLE_KEY=pat..."; \
		exit 1; \
	fi

# Run linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	golangci-lint run ./...

