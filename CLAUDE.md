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

## Architecture

### Core Types (airtable package)

- **Table[T]** (`table.go`): Generic table representation. Created via `NewTable[T](baseId, tableId)` where T is the schema struct. Supports method chaining with `WithLimit()`, `WithFilter()`, `WithSort()`, `WithTypecast()`, and `WithOptions()`. Options reset to defaults after each `List()` call.

- **Record[T]** (`record.go`): Single record with Fields of type T. Can be created via `table.NewRecord()` or `NewRecord[T](baseId, tableId)`. Supports `Save()` (creates or updates based on Id presence) and `Destroy()`.

- **Records[T]** (`record.go`): Slice of `*Record[T]` returned by `List()` and `Find()`.

### Operations

- `list.go`: Handles pagination automatically, builds queries via `queryBuilder`
- `upsert.go`: Unified create/update logic with `insert()` and `update()` functions
- `get.go`: Single record retrieval
- `destroy.go`: Batch deletion

### HTTP Layer

- `client.go`: HTTP client interface with `AirtableClient` implementation and mock clients for testing (`mockClient`, `mockClientPaginate`)
- `config.go`: Global token and client configuration via `SetToken()` and `Configure()`
- `common.go`: Shared types (`Options`, `Sort`, `Sorts`) and request helpers

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
