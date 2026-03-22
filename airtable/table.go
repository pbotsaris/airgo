package airtable

import (
	"context"
	"fmt"
	"net/url"
	"slices"

	"github.com/Antfood/airgo/utils"
)

// filterFormula builds an Airtable filter formula for exact field matching
func filterFormula(field, value string) string {
	return fmt.Sprintf("{%s} = '%s'", field, value)
}

var defaultOptions = Options{
	Limit:          10,
	MaxRecords:     0,
	Filter:         "",
	Sort:           Sorts{},
	Typecast:       false,
	Fields:         nil,
	View:           "",
	CellFormat:     "",
	TimeZone:       "",
	UserLocale:     "",
	RecordMetadata: nil,
}

/*
NewTable returns a new Table instance for a given baseId and tableId.

You Must provide a struct T as the schema for the table.

Example:

	type MySchema struct {
	   Name string `json:"name"`
	   Age  int    `json:"age"`
	}

	table := airtable.NewTable[MySchema]("baseId", "tableId")
*/
func NewTable[T any](baseId, tableId string) *Table[T] {
	return &Table[T]{
		BaseId:  baseId,
		TableId: tableId,
		Options: defaultOptions,
	}
}

/* Table represents an Airtable table while type T is the schema of the table. */
type Table[T any] struct {
	BaseId  string
	TableId string
	Options Options
}

/*
WithOptions sets the options for listing records from a table. Available options are Limit, Filter and Sort.

Options are only used while listing records from the Airtable.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")

			opts := airtable.Options{
			   Limit:  10,
			   Filter: "{name} = 'John'",
			   Sort:   airtable.Sorts{
			      {Field: "name", Direction: airtable.Asc},
			      {Field: "age", Direction: airtable.Desc},
			   },
		   }

	records, err := table.WithOptions(opts).List()
*/
func (t *Table[T]) WithOptions(opts Options) *Table[T] {
	t.Options = opts
	return t
}

/*
WithLimit sets the limit of records when listing.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithLimit(10).List()
*/
func (t *Table[T]) WithLimit(limit int) *Table[T] {
	t.Options.Limit = limit
	return t
}

/*
WithFilter sets the filter when listing records.

See https://support.airtable.com/docs/formula-field-reference for info filtering options.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithFilter("{name} = 'John'").List()
*/
func (t *Table[T]) WithFilter(filter string) *Table[T] {
	t.Options.Filter = filter
	return t
}

/*
WithSort allows you to sort the records when listing. More than one sort option can be provided.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	table.WithSort(airtable.Sorts{
	   {Field: "name", Direction: airtable.Asc},
	   {Field: "age", Direction: airtable.Desc},
	}).List()
*/
func (t *Table[T]) WithSort(sorts Sorts) *Table[T] {
	t.Options.Sort = sorts
	return t
}

/*
WithTypecast will enable typecast conversion in airtable when creating and updating records.
This is useful when creating single or mutliple select fields as it allows to create new values on the fly.

Example:

	 table := airtable.NewTable[MySchema]("baseId", "tableId").WithTypecast()

	 record := table.NewRecord()
    record.Fields.SomeSingleSelectField = "New Name" // unexisting single select valuetable

	  err := table.Create(record)
*/
func (t *Table[T]) WithTypecast() *Table[T] {
	t.Options.Typecast = true
	return t
}

/*
WithMaxRecords sets the maximum total number of records to return across all pages.
This is different from Limit which controls records per page.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithMaxRecords(500).List() // Returns up to 500 records total
*/
func (t *Table[T]) WithMaxRecords(max int) *Table[T] {
	t.Options.MaxRecords = max
	return t
}

/*
WithFields specifies which fields to return from the Airtable.
If not specified, all fields from the schema are returned.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithFields("name", "age").List()
*/
func (t *Table[T]) WithFields(fields ...string) *Table[T] {
	t.Options.Fields = fields
	return t
}

/*
WithView specifies a view to use for filtering and sorting.
The view's filters and sorts will be applied to the results.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithView("Grid view").List()
*/
func (t *Table[T]) WithView(view string) *Table[T] {
	t.Options.View = view
	return t
}

/*
WithCellFormat specifies the format for cell values.
Options are "json" (default) or "string".

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithCellFormat("string").List()
*/
func (t *Table[T]) WithCellFormat(format string) *Table[T] {
	t.Options.CellFormat = format
	return t
}

/*
WithTimeZone specifies the time zone for formatting dates.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithTimeZone("America/New_York").List()
*/
func (t *Table[T]) WithTimeZone(tz string) *Table[T] {
	t.Options.TimeZone = tz
	return t
}

/*
WithUserLocale specifies the locale for formatting values.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithUserLocale("en-US").List()
*/
func (t *Table[T]) WithUserLocale(locale string) *Table[T] {
	t.Options.UserLocale = locale
	return t
}

/*
WithRecordMetadata specifies additional metadata to return with each record.
For example, "commentCount" returns the number of comments on each record.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithRecordMetadata("commentCount").List()
*/
func (t *Table[T]) WithRecordMetadata(metadata ...string) *Table[T] {
	t.Options.RecordMetadata = metadata
	return t
}

/*
ListCtx retrieves records from a table based on the provided schema T.
It accepts a context for cancellation and timeout control.

Example:

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.ListCtx(ctx)
*/
func (t *Table[T]) ListCtx(ctx context.Context) (Records[T], error) {
	url := createRequestUrl(t.BaseId, t.TableId)
	records, err := list[T](ctx, url, t.Options)

	for _, r := range records {
		r.BaseId = t.BaseId
		r.TableId = t.TableId
	}

	t.Options = defaultOptions
	return records, err
}

/*
List retrieves records from from a table base on the provided schema T.

Example:

	type MySchema struct {
	   Name string `json:"name"`
	   Age  int    `json:"age"`
	}

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.List()
*/
func (t *Table[T]) List() (Records[T], error) {
	return t.ListCtx(context.Background())
}

/*
GetCtx retrieves a single record with a given ID.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	record, err := table.GetCtx(ctx, "rec123")
*/
func (t Table[T]) GetCtx(ctx context.Context, id string) (Record[T], error) {
	record := Record[T]{Id: id, TableId: t.TableId, BaseId: t.BaseId}
	url := createRequestUrl(t.BaseId, t.TableId)
	return get(ctx, url, record)
}

/*
Get retrieves a single record with a given ID.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	record, err := table.Get("rec123")
*/
func (t Table[T]) Get(id string) (Record[T], error) {
	return t.GetCtx(context.Background(), id)
}

/*
UpdateCtx updates a set of records in the Airtable.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	err := table.UpdateCtx(ctx, records...)
*/
func (t Table[T]) UpdateCtx(ctx context.Context, records ...*Record[T]) error {
	url := createRequestUrl(t.BaseId, t.TableId)

	updateRequests := make([]updateRequest, len(records))

	for i, r := range records {
		fields, err := utils.StructJsonToMap(r.Fields, utils.WithIgnore())

		if err != nil {
			return &Error{Op: OpUpdate, Message: "failed to convert struct to map", Err: err}
		}
		updateRequests[i] = updateRequest{r.Id, fields, t.Options.Typecast}
	}

	return update(ctx, url, records, updateRequests...)
}

/*
Update updates a set of records in the Airtable.

Example:

		table := airtable.NewTable[MySchema]("baseId", "tableId")

		// retrieves 2 records
		records, err := table.WithLimit(2).List()

	   // update the records
	   for _, r := range records {
	      r.Fields.Name = "Updated Name"
	   }

	   err := table.Update(records...)
*/
func (t Table[T]) Update(records ...*Record[T]) error {
	return t.UpdateCtx(context.Background(), records...)
}

/*
ReplaceCtx performs a full replacement of records in the Airtable using PUT.
Unlike UpdateCtx (PATCH), ReplaceCtx will clear any fields not provided in the request.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	err := table.ReplaceCtx(ctx, records...)
*/
func (t Table[T]) ReplaceCtx(ctx context.Context, records ...*Record[T]) error {
	url := createRequestUrl(t.BaseId, t.TableId)

	replaceRequests := make([]replaceRequest, len(records))

	for i, r := range records {
		fields, err := utils.StructJsonToMap(r.Fields, utils.WithoutIgnore())

		if err != nil {
			return &Error{Op: OpReplace, Message: "failed to convert struct to map", Err: err}
		}
		replaceRequests[i] = replaceRequest{r.Id, fields, t.Options.Typecast}
	}

	return replace(ctx, url, records, replaceRequests...)
}

/*
Replace performs a full replacement of records in the Airtable using PUT.
Unlike Update (PATCH), Replace will clear any fields not provided in the request.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")

	records, err := table.WithLimit(2).List()

	// Replace entire record - fields not set will be cleared
	for _, r := range records {
	    r.Fields = MySchema{Name: "New Name"} // Age will be cleared
	}

	err := table.Replace(records...)
*/
func (t Table[T]) Replace(records ...*Record[T]) error {
	return t.ReplaceCtx(context.Background(), records...)
}

/*
CreateCtx creates a set of records in the Airtable.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	err := table.CreateCtx(ctx, records...)
*/
func (t Table[T]) CreateCtx(ctx context.Context, records ...*Record[T]) error {
	url := createRequestUrl(t.BaseId, t.TableId)

	createRequests := make([]createRequest, len(records))

	for i, r := range records {
		fields, err := utils.StructJsonToMap(r.Fields, utils.WithIgnore())

		if err != nil {
			return &Error{Op: OpCreate, Message: "failed to convert struct to map", Err: err}
		}

		createRequests[i] = createRequest{fields, t.Options.Typecast}
	}

	return insert(ctx, url, records, createRequests...)
}

/*
Create creates a set of records in the Airtable.

Example:

		table := airtable.NewTable[MySchema]("baseId", "tableId")

	   // create 2 records
	   records := table.NewRecords(2)

	   for _, r := range records {
	      r.Fields.Name = "New Name"
	   }

	   err := table.Create(records...)
*/
func (t Table[T]) Create(records ...*Record[T]) error {
	return t.CreateCtx(context.Background(), records...)
}

/*
DestroyCtx deletes a set of records in the Airtable and returns details of the destroyed records.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	destroyedRecords, err := table.DestroyCtx(ctx, records...)
*/
func (t Table[T]) DestroyCtx(ctx context.Context, records ...*Record[T]) ([]*destroyedRecord, error) {
	url := createRequestUrl(t.BaseId, t.TableId)
	resp, err := destroy(ctx, url, records...)

	for _, r := range resp {
		r.BaseId = t.BaseId
		r.TableId = t.TableId
	}

	return resp, err
}

/*
Destroy deletes a set of records in the Airtable and returns details of the destroyed records.

Example:

	      table := airtable.NewTable[MySchema]("baseId", "tableId")

	      // retrieves 2 records
			records, err := table.WithLimit(2).List()

	      // destroy the records
	      destroyedRecords, err := table.Destroy(records...)
*/
func (t Table[T]) Destroy(records ...*Record[T]) ([]*destroyedRecord, error) {
	return t.DestroyCtx(context.Background(), records...)
}

/*
FindCtx searches for records in the Airtable table based on a specific field and its value.
It accepts a context for cancellation and timeout control.

The method checks if the given field exists in the schema T. If the field doesn't exist, it returns an error.
Otherwise, it uses the Airtable API's filter functionality to retrieve matching records.

Note: The field parameter should match the JSON field name, not the Go struct field name.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.FindCtx(ctx, "name", "John")
*/
func (t Table[T]) FindCtx(ctx context.Context, field, value string) (Records[T], error) {
	var schema T
	fieldNames, err := utils.GetStructFieldJsonNames(schema)

	if err != nil {
		return nil, &Error{Op: OpFind, Message: "failed to get struct field names", Err: err}
	}

	if slices.Contains(fieldNames, field) {
		return t.WithFilter(filterFormula(field, value)).ListCtx(ctx)
	}

	return nil, &ValidationError{
		Op:      OpFind,
		Message: "field '" + field + "' not found in schema",
	}
}

/*
Find searches for records in the Airtable table based on a specific field and its value.

The method checks if the given field exists in the schema T. If the field doesn't exist, it returns an error.
Otherwise, it uses the Airtable API's filter functionality to retrieve matching records.

Note: The field parameter should match the JSON field name, not the Go struct field name.

Example:

	type MySchema struct {
	   Name string `json:"name"`
	   Age  int    `json:"age"`
	}

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.Find("name", "John")
*/
func (t Table[T]) Find(field, value string) (Records[T], error) {
	return t.FindCtx(context.Background(), field, value)
}

/*
GetFieldsCtx retrieves field metadata for this table from the Airtable Meta API.
It accepts a context for cancellation and timeout control.
Results are cached after the first call. Use RefreshFieldsCtx to force a refresh.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	fields, err := table.GetFieldsCtx(ctx)
*/
func (t *Table[T]) GetFieldsCtx(ctx context.Context) ([]Field, error) {
	// Check cache first
	if fields := cache.get(t.BaseId, t.TableId); fields != nil {
		return fields, nil
	}

	// Fetch from Meta API
	fields, err := fetchTableFields(ctx, t.BaseId, t.TableId)
	if err != nil {
		return nil, err
	}

	// Cache the result
	cache.set(t.BaseId, t.TableId, fields)
	return fields, nil
}

/*
GetFields retrieves field metadata for this table from the Airtable Meta API.
Results are cached after the first call. Use RefreshFields to force a refresh.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	fields, err := table.GetFields()
	for _, f := range fields {
	    fmt.Printf("Field: %s (ID: %s), Type: %s\n", f.Name, f.Id, f.Type)
	}
*/
func (t *Table[T]) GetFields() ([]Field, error) {
	return t.GetFieldsCtx(context.Background())
}

/*
GetFieldCtx retrieves a single field by name or ID.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	field, err := table.GetFieldCtx(ctx, "Name")
*/
func (t *Table[T]) GetFieldCtx(ctx context.Context, nameOrId string) (*Field, error) {
	fields, err := t.GetFieldsCtx(ctx)
	if err != nil {
		return nil, err
	}

	for _, f := range fields {
		if f.Name == nameOrId || f.Id == nameOrId {
			return &f, nil
		}
	}

	return nil, &APIError{
		Op:      OpGetFields,
		Type:    ErrTypeNotFound,
		Message: "field '" + nameOrId + "' not found",
	}
}

/*
GetField retrieves a single field by name or ID.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	field, err := table.GetField("Name")
	fmt.Printf("Field ID: %s, Type: %s\n", field.Id, field.Type)
*/
func (t *Table[T]) GetField(nameOrId string) (*Field, error) {
	return t.GetFieldCtx(context.Background(), nameOrId)
}

/*
RefreshFieldsCtx clears the cache and fetches fresh field metadata from the Meta API.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	fields, err := table.RefreshFieldsCtx(ctx)
*/
func (t *Table[T]) RefreshFieldsCtx(ctx context.Context) ([]Field, error) {
	cache.delete(t.BaseId, t.TableId)
	return t.GetFieldsCtx(ctx)
}

/*
RefreshFields clears the cache and fetches fresh field metadata from the Meta API.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	fields, err := table.RefreshFields() // Forces a fresh fetch
*/
func (t *Table[T]) RefreshFields() ([]Field, error) {
	return t.RefreshFieldsCtx(context.Background())
}

/*
	NewRecord returns a new Record instance for a given table.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	record := table.NewRecord()
*/
func (t Table[T]) NewRecord() *Record[T] {
   return &Record[T]{BaseId: t.BaseId, TableId: t.TableId, Typecast: t.Options.Typecast}
}

/*
	NewRecords returns a slice of new Records instances for a given table.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records := table.NewRecords(2)

	len(records) // 2
*/
func (t Table[T]) NewRecords(nbRecords int) Records[T] {

	records := make(Records[T], nbRecords)

	for i := 0; i < nbRecords; i++ {
      records[i] = &Record[T]{BaseId: t.BaseId, TableId: t.TableId, Typecast: t.Options.Typecast}
	}

	return records
}

/* createRequestUrl constructs the request URL based on a baseId and tableId. */
func createRequestUrl(baseId string, tableId string) string {
	u, err := url.JoinPath(config.EndpointUrl, baseId, tableId)
	if err != nil {
		return ""
	}
	return u
}
