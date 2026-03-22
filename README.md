# airgo
airgo uses Go generics to map Airtable tables to typed structs. You define your schema once, and all operations return typed data instead of `map[string]interface{}`.

## Design

```go
type Track struct {
    Title    string   `json:"Title"`
    Artist   string   `json:"Artist"`
    Duration int      `json:"Duration (seconds)"`
    Tags     []string `json:"Tags"`
}

table := airtable.NewTable[Track]("baseId", "Tracks")
records, _ := table.List()

// records is []Record[Track]
for _, r := range records {
    fmt.Println(r.Fields.Title)       // string
    fmt.Println(r.Fields.Duration)    // int
}
```

The generic `Table[T]` and `Record[T]` types carry through all operations—listing, creating, updating—so you work with your schema type throughout.

## Installation

```bash
go get github.com/pbotsaris/airgo
```

## Quick Start

### 1. Define Your Schema

Define a struct that maps to your Airtable table fields using JSON tags:

```go
type Track struct {
    Title       string   `json:"Title"`
    Artist      string   `json:"Artist"`
    Duration    int      `json:"Duration (seconds)"`
    Tags        []string `json:"Tags"`
    ReleaseDate string   `json:"Release Date"`
}
```

For complex Airtable fields, use the built-in types:

```go
type Track struct {
    Title     string                  `json:"Title"`
    Artwork   []airtable.Attachment   `json:"Artwork"`    // file attachments
    CreatedBy *airtable.Collaborator  `json:"Created By"` // user field (read-only)
}
```

## Field Type Reference

Map Airtable field types to Go types as follows:

| Airtable Field | Go Type | Notes |
|----------------|---------|-------|
| Single line text | `string` | |
| Long text | `string` | May contain markdown or mentions |
| Email | `string` | |
| URL | `string` | |
| Phone number | `string` | |
| Number | `float64` or `int` | Use `float64` for decimals |
| Currency | `float64` | |
| Percent | `float64` | Stored as decimal (0.5 = 50%) |
| Duration | `int` | Seconds |
| Rating | `int` | 1-5 (or max configured) |
| Checkbox | `bool` | |
| Date | `string` | ISO 8601 format: "2024-03-22" |
| Date and time | `string` | ISO 8601: "2024-03-22T14:30:00.000Z" |
| Single select | `string` | Option name |
| Multiple select | `[]string` | Array of option names |
| Attachment | `[]airtable.Attachment` | See Attachment type below |
| Collaborator | `*airtable.Collaborator` | Single user |
| Multiple collaborators | `[]airtable.Collaborator` | Multiple users |
| Link to another record | `[]string` | Array of record IDs |
| Barcode | `*YourBarcodeType` | Custom struct (see below) |
| Button | Read-only | Cannot be set via API |
| Formula | `string`, `float64`, or `any` | Depends on formula result |
| Rollup | `any` | Depends on rollup configuration |
| Lookup | `[]any` | Array of looked-up values |
| Count | `int` | Read-only |
| Auto number | `int` | Read-only |
| Created time | `string` | Read-only, ISO 8601 |
| Last modified time | `string` | Read-only, ISO 8601 |
| Created by | `*airtable.Collaborator` | Read-only |
| Last modified by | `*airtable.Collaborator` | Read-only |

### Complete Example

Here's a schema using all common field types:

```go
type Record struct {
    // Editable fields
    Name         string                  `json:"Name"`
    Notes        string                  `json:"Notes,omitempty"`
    Email        string                  `json:"Email,omitempty"`
    Website      string                  `json:"Website,omitempty"`
    Phone        string                  `json:"Phone,omitempty"`
    Status       string                  `json:"Status,omitempty"`       // single select
    Tags         []string                `json:"Tags,omitempty"`         // multi select
    Price        float64                 `json:"Price,omitempty"`
    Quantity     int                     `json:"Quantity,omitempty"`
    Discount     float64                 `json:"Discount,omitempty"`     // percent (0.1 = 10%)
    Duration     int                     `json:"Duration,omitempty"`     // seconds
    Rating       int                     `json:"Rating,omitempty"`
    IsActive     bool                    `json:"Active,omitempty"`
    DueDate      string                  `json:"Due Date,omitempty"`
    Assignee     *airtable.Collaborator  `json:"Assignee,omitempty"`
    Team         []airtable.Collaborator `json:"Team,omitempty"`
    Attachments  []airtable.Attachment   `json:"Attachments,omitempty"`
    RelatedItems []string                `json:"Related Items,omitempty"` // linked records

    // Read-only fields (use update:"ignore" to exclude from updates)
    Formula      string                  `json:"Formula,omitempty" update:"ignore"`
    CreatedBy    *airtable.Collaborator  `json:"Created By,omitempty" update:"ignore"`
    CreatedTime  string                  `json:"Created Time,omitempty" update:"ignore"`
    ModifiedBy   *airtable.Collaborator  `json:"Modified By,omitempty" update:"ignore"`
    ModifiedTime string                  `json:"Modified Time,omitempty" update:"ignore"`
    RecordCount  int                     `json:"Record Count,omitempty" update:"ignore"`
    AutoNumber   int                     `json:"ID,omitempty" update:"ignore"`
}
```

### Built-in Types

**Attachment:**
```go
type Attachment struct {
    ID         string      `json:"id,omitempty"`
    URL        string      `json:"url"`           // Required when creating
    Filename   string      `json:"filename,omitempty"`
    Size       int         `json:"size,omitempty"`
    Type       string      `json:"type,omitempty"`
    Thumbnails *Thumbnails `json:"thumbnails,omitempty"`
}
```

**Collaborator:**
```go
type Collaborator struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}
```

**Custom Barcode type (define in your code):**
```go
type Barcode struct {
    Text string `json:"text,omitempty"`
    Type string `json:"type,omitempty"` // e.g., "upce", "code39"
}
```

### 2. Configure the Client

```go
import "github.com/pbotsaris/airgo/airtable"

func main() {
    airtable.SetToken(os.Getenv("AIRTABLE_TOKEN"))
}
```

### 3. Create a Table and Query

```go
table := airtable.NewTable[Track]("appXXXXXXXXXXXXXX", "Tracks")

// List all records
records, err := table.List()

// List with options
records, err := table.
    WithFilter("{Artist} = 'Miles Davis'").
    WithSort(airtable.Sorts{{Field: "Release Date", Direction: "desc"}}).
    WithLimit(10).
    List()

// Access record data
for _, record := range records {
    fmt.Printf("%s by %s\n", record.Fields.Title, record.Fields.Artist)
}
```

## Configuration

### Basic Setup

For simple use cases, set only the token:

```go
airtable.SetToken(os.Getenv("AIRTABLE_TOKEN"))
```

### Custom HTTP Client

Inject a custom HTTP client for testing or custom transport:

```go
airtable.Configure(customClient, os.Getenv("AIRTABLE_TOKEN"))
```

### Full Configuration

For complete control, use `ConfigureWithOptions`:

```go
airtable.ConfigureWithOptions(airtable.Config{
    Token:                os.Getenv("AIRTABLE_TOKEN"),
    EndpointUrl:          "https://api.airtable.com/v0", // custom endpoint/proxy
    MaxPageSize:          50,                            // records per page (max 100)
    MaxUrlLength:         15000,                         // threshold for POST fallback
    RequestTimeout:       2 * time.Minute,               // HTTP timeout
    NoRetryIfRateLimited: false,                         // disable retry on 429
    CustomHeaders:        map[string]string{             // additional headers
        "X-Custom-Header": "value",
    },
})
```

## Query Options

Chain methods to build queries. Options reset after each `List()` call.

```go
table := airtable.NewTable[Track]("baseId", "tableId")

records, err := table.
    WithFilter("{Status} = 'Active'").       // Airtable formula
    WithSort(airtable.Sorts{
        {Field: "Created", Direction: "desc"},
        {Field: "Title", Direction: "asc"},
    }).
    WithLimit(25).                           // records per page
    WithMaxRecords(100).                     // total records across pages
    WithFields("Title", "Artist").           // select specific fields
    WithView("Grid view").                   // use a saved view
    WithCellFormat("string").                // "json" (default) or "string"
    WithTimeZone("America/New_York").        // for date formatting
    WithUserLocale("en-US").                 // for value formatting
    WithRecordMetadata("commentCount").      // include metadata
    List()
```

| Method | Description |
|--------|-------------|
| `WithFilter(formula)` | Filter using Airtable formula |
| `WithSort(sorts)` | Sort by one or more fields |
| `WithLimit(n)` | Records per page (max 100) |
| `WithMaxRecords(n)` | Total records to return |
| `WithFields(fields...)` | Select specific fields |
| `WithView(name)` | Use a saved view's configuration |
| `WithCellFormat(format)` | Response format: "json" or "string" |
| `WithTimeZone(tz)` | Timezone for date fields |
| `WithUserLocale(locale)` | Locale for formatting |
| `WithRecordMetadata(fields...)` | Request metadata like "commentCount" |
| `WithTypecast()` | Enable type coercion on write |

## Operations

### List Records

```go
table := airtable.NewTable[Track]("baseId", "tableId")
records, err := table.List()
```

Pagination is handled automatically. For large filter formulas, the client automatically switches to POST requests.

### Find Records by ID

```go
records, err := table.Find("recXXXXXXXXXXXXXX", "recYYYYYYYYYYYYYY")
```

### Create Records

```go
// Create via Table
record := table.NewRecord()
record.Fields = Track{Title: "New Song", Artist: "New Artist"}
err := record.Save()

// Or create multiple
records := airtable.Records[Track]{record1, record2}
err := table.Create(records...)
```

### Update Records (PATCH)

Updates only the fields present in your struct:

```go
record.Fields.Title = "Updated Title"
err := record.Save() // Uses PATCH when record has an ID
```

Use `update:"ignore"` to exclude fields from updates:

```go
type Track struct {
    Title     string `json:"Title"`
    CreatedAt string `json:"Created At" update:"ignore"` // never sent on update
}
```

### Replace Records (PUT)

Replaces the entire record, clearing any fields not in your struct:

```go
err := record.Replace()

// Or replace multiple
err := table.Replace(record1, record2)
```

### Delete Records

```go
err := record.Destroy()

// Or delete multiple
err := table.Destroy(record1, record2)
```

## Record Metadata

Records include metadata from Airtable:

```go
record.Id           // Record ID (e.g., "recXXXXXXXXXXXXXX")
record.CreatedTime  // ISO 8601 timestamp
record.CommentCount // Comment count (when requested via WithRecordMetadata)
```

## Field Metadata

Access field definitions via the Meta API:

```go
// Get all fields (cached after first call)
fields, err := table.GetFields()

for _, field := range fields {
    fmt.Printf("%s (%s): %s\n", field.Name, field.Type, field.Id)
}

// Get a specific field by name or ID
field, err := table.GetField("Title")
fmt.Printf("Field ID: %s, Type: %s\n", field.Id, field.Type)

// Force refresh the cache
fields, err := table.RefreshFields()
```

The `Field` struct contains:

```go
type Field struct {
    Id          string         // Field ID (e.g., "fldXXXXXXXXXXXXXX")
    Name        string         // Field name
    Type        string         // Field type (e.g., "singleLineText", "number")
    Description string         // Field description (if set)
    Options     map[string]any // Type-specific options
}
```

## Context Support

All operations have `*Ctx` variants for cancellation and timeout control:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

records, err := table.ListCtx(ctx)
record, err := table.GetCtx(ctx, "recXXX")
err = table.CreateCtx(ctx, records...)
err = record.SaveCtx(ctx)
```

Context is checked between pagination pages and retry attempts. Methods without context (e.g., `List()`) use `context.Background()` internally.

## Error Handling

The client includes automatic retry with exponential backoff for rate limits (429) and server errors (5xx). Disable retries via config:

```go
airtable.ConfigureWithOptions(airtable.Config{
    Token:                os.Getenv("AIRTABLE_TOKEN"),
    NoRetryIfRateLimited: true,
})
```

### Error Types

All errors returned by the library are structured types that preserve context from the Airtable API. Use `errors.Is()` to check for common error conditions:

```go
records, err := table.List()
if err != nil {
    if errors.Is(err, airtable.ErrNotFound) {
        // Resource not found (404 or NOT_FOUND error type)
    }
    if errors.Is(err, airtable.ErrUnauthorized) {
        // Authentication/permission issue (401, 403)
    }
    if errors.Is(err, airtable.ErrRateLimited) {
        // Rate limited (429)
    }
    if errors.Is(err, airtable.ErrValidation) {
        // Invalid request or value
    }
    if errors.Is(err, airtable.ErrNotConfigured) {
        // Client not configured (forgot to call SetToken)
    }
}
```

### Extracting Error Details

Use `errors.As()` to extract detailed error information:

```go
records, err := table.List()
if err != nil {
    var apiErr *airtable.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("Operation: %s\n", apiErr.Op)         // e.g., "List"
        fmt.Printf("Status: %d\n", apiErr.StatusCode)    // e.g., 403
        fmt.Printf("Type: %s\n", apiErr.Type)            // e.g., "INVALID_PERMISSIONS_OR_MODEL_NOT_FOUND"
        fmt.Printf("Message: %s\n", apiErr.Message)      // Human-readable message
    }
}
```

### Sentinel Errors

| Error | Description |
|-------|-------------|
| `ErrNotFound` | Resource not found (404 or `NOT_FOUND` type) |
| `ErrUnauthorized` | Auth issues (401, 403, `UNAUTHORIZED`, `INVALID_PERMISSIONS_OR_MODEL_NOT_FOUND`) |
| `ErrRateLimited` | Rate limited (429) |
| `ErrValidation` | Invalid request or value (`INVALID_VALUE`, `INVALID_REQUEST`) |
| `ErrNotConfigured` | Client not configured |
| `ErrMissingRecordID` | Operation requires a record ID |

### Error Structs

| Type | Fields | Description |
|------|--------|-------------|
| `APIError` | `Op`, `StatusCode`, `Type`, `Message` | Airtable API errors |
| `ValidationError` | `Op`, `Message`, `Field` | Local validation errors |
| `ConfigError` | `Op`, `Message` | Configuration errors |
| `HTTPError` | `Op`, `StatusCode`, `Message` | HTTP-level errors |

## Development

### Make Commands

| Command | API Calls | Description |
|---------|-----------|-------------|
| `make test` | No | Run all unit tests |
| `make test-verbose` | No | Run unit tests with verbose output |
| `make test-unit` | No | Run unit tests only (skips integration) |
| `make test-integration` | No | Replay integration tests from recorded fixtures |
| `make test-record` | **Yes** | Record new fixtures by calling the real Airtable API |
| `make lint` | No | Run golangci-lint |

### Integration Tests

Integration tests use [go-vcr](https://github.com/dnaeon/go-vcr) to record and replay HTTP interactions. This allows testing against real API responses without making API calls on every test run.

**First time setup:**

1. Create a `.env` file with your Airtable token:
   ```
   AIRTABLE_KEY=patXXXXXXXXXXXXXX
   ```

2. Record fixtures (calls the real API):
   ```bash
   make test-record
   ```

3. Fixtures are saved to `airtable/testdata/fixtures/*.yaml`

**Normal development:**

```bash
make test-integration  # Replays from fixtures, no API calls
make test              # Runs all unit tests
```

**Refreshing fixtures:**

When the Airtable API changes or you need fresh data:
```bash
make test-record  # Re-records all fixtures
```

Recorded fixtures have auth headers automatically scrubbed and can be committed to version control.

## License

MIT
