# airgo
A type-safe Go client for the Airtable API using generics.

## Design

airgo uses Go generics to map Airtable tables to typed structs. You define your schema once, and all operations return typed data instead of `map[string]interface{}`.

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
go get github.com/Antfood/airgo
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

### 2. Configure the Client

```go
import "github.com/Antfood/airgo/airtable"

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

## License

MIT
