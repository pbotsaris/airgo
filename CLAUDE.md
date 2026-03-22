# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

airgo (`github.com/Antfood/airgo`) is a Go library for interacting with the Airtable API. It provides a type-safe, generic interface for CRUD operations on Airtable tables. The library has no external dependencies.

## Build and Test Commands

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run a specific test
go test -v ./airtable -run TestTableFind
```

## Linting

```bash
# Install golangci-lint (system-wide, not a project dependency)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
make lint

# Auto-fix some issues (formatting, imports)
go fmt ./...
goimports -w .
```

## Configuration

Before making API calls, set the Airtable token:

```go
airtable.SetToken(os.Getenv("AIRTABLE_TOKEN"))
```

Or use `Configure()` for custom HTTP clients:

```go
airtable.Configure(customClient, token)
```

For full configuration control, use `ConfigureWithOptions()`:

```go
airtable.ConfigureWithOptions(airtable.Config{
    Token:                os.Getenv("AIRTABLE_TOKEN"),
    EndpointUrl:          "https://api.airtable.com/v0", // for proxies/enterprise
    MaxPageSize:          50,                            // default: 100
    MaxUrlLength:         10000,                         // default: 15000
    RequestTimeout:       2 * time.Minute,               // default: 5 minutes
    NoRetryIfRateLimited: false,
    CustomHeaders:        map[string]string{"X-Custom": "value"},
})
```

## Architecture

### Core Types (airtable package)

- **Table[T]** (`table.go`): Generic table representation. Created via `NewTable[T](baseId, tableId)` where T is the schema struct. Supports method chaining with query options (see below). Options reset to defaults after each `List()` call.

- **Record[T]** (`record.go`): Single record with Fields of type T. Can be created via `table.NewRecord()` or `NewRecord[T](baseId, tableId)`. Supports `Save()` (creates or updates based on Id presence), `Replace()` (PUT - full replacement), and `Destroy()`. Includes metadata fields `CreatedTime` and `CommentCount`.

- **Field** (`meta.go`): Field metadata from the Meta API. Contains `Id`, `Name`, `Type`, `Description`, `Options`.

- **Records[T]** (`record.go`): Slice of `*Record[T]` returned by `List()` and `Find()`.

### Query Options (chainable methods on Table)

- `WithLimit(n)`: Records per page (max 100)
- `WithMaxRecords(n)`: Total records to return across all pages
- `WithFilter(formula)`: Airtable formula filter
- `WithSort(sorts)`: Sort by fields
- `WithFields(fields...)`: Select specific fields to return
- `WithView(name)`: Use a saved view's filters/sorts
- `WithCellFormat(format)`: "json" (default) or "string"
- `WithTimeZone(tz)`: Timezone for date formatting
- `WithUserLocale(locale)`: Locale for value formatting
- `WithRecordMetadata(fields...)`: Request metadata like "commentCount"
- `WithTypecast()`: Enable type coercion for creates/updates

### Operations

- `list.go`: Handles pagination automatically, builds queries via `queryBuilder`. Auto-switches to POST for queries exceeding 15KB URL length.
- `upsert.go`: Unified create/update/replace logic with `insert()`, `update()`, and `replace()` functions
- `get.go`: Single record retrieval
- `destroy.go`: Batch deletion
- `meta.go`: Field metadata via Meta API with thread-safe caching. Provides `GetFields()`, `GetField(nameOrId)`, `RefreshFields()`.

### HTTP Layer

- `client.go`: HTTP client interface with `AirtableClient` implementation and mock clients for testing (`mockClient`, `mockClientPaginate`)
- `config.go`: Global configuration via `Config` struct. Supports `SetToken()`, `Configure()`, and `ConfigureWithOptions()`. Configurable: `EndpointUrl`, `MaxPageSize`, `MaxUrlLength`, `RequestTimeout`, `NoRetryIfRateLimited`, `CustomHeaders`.
- `common.go`: Shared types (`Options`, `Sort`, `Sorts`, request types) and HTTP helpers

### Internal Packages

- `retry/`: HTTP retry logic with exponential backoff and jitter
- `utils/`: Struct reflection utilities (`StructJsonToMap`, `GetStructFieldJsonNames`, `Map`, `Filter`, etc.)
- `testutils/testutils/`: Test helpers (`Assert()`, `Ok()`, `Equals()`)

### Testing Pattern

Tests use mock HTTP clients injected via `Configure(client, token)`. Test data is stored in `airtable/testdata/` JSON files. The `testutils` package provides `Assert()`, `Ok()`, and `Equals()` helpers imported with dot notation.

## Schema Definition

Define table schemas as structs with JSON tags matching Airtable field names:

```go
type MySchema struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
```

Use `update:"ignore"` tag to exclude fields when updating records:

```go
type MySchema struct {
    Name      string `json:"name"`
    CreatedAt string `json:"created_at" update:"ignore"`
}
```
