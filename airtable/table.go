package airtable

import (
	"fmt"
	"net/url"
	"slices"

	"github.com/Antfood/airgo/utils"
)

var defaultOptions = Options{
	Limit:    10,
	Filter:   "",
	Sort:     Sorts{},
	Typecast: false,
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
	url := createRequestUrl(t.BaseId, t.TableId)
	records, err := list[T](url, t.Options)

	for _, r := range records {
		r.BaseId = t.BaseId
		r.TableId = t.TableId
	}

	t.Options = defaultOptions
	return records, err
}

/*
Get retrieves a single record with a given ID.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	record, err := table.Get("rec123")
*/
func (t Table[T]) Get(id string) (Record[T], error) {
	record := Record[T]{Id: id, TableId: t.TableId, BaseId: t.BaseId}
	url := createRequestUrl(t.BaseId, t.TableId)
	return get(url, record)
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
	url := createRequestUrl(t.BaseId, t.TableId)

	updateRequests := make([]updateRequest, len(records))

	for i, r := range records {
		fields, err := utils.StructJsonToMap(r.Fields, utils.WithIgnore())

		if err != nil {
			return fmt.Errorf("airtable.Update: Error converting struct to map: %v", err)
		}
		updateRequests[i] = updateRequest{r.Id, fields, t.Options.Typecast}
	}

	return update(url, records, updateRequests...)
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
	url := createRequestUrl(t.BaseId, t.TableId)

	createRequests := make([]createRequest, len(records))

	for i, r := range records {
		fields, err := utils.StructJsonToMap(r.Fields, utils.WithIgnore())

		if err != nil {
			return fmt.Errorf("airtable.Update: Error converting struct to map: %v", err)
		}

		createRequests[i] = createRequest{fields, t.Options.Typecast}
	}

	return insert(url, records, createRequests...)
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
	url := createRequestUrl(t.BaseId, t.TableId)
	resp, err := destroy(url, records...)

	for _, r := range resp {
		r.BaseId = t.BaseId
		r.TableId = t.TableId
	}

	return resp, err
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
	var schema T
	fieldNames, err := utils.GetStructFieldJsonNames(schema)

	if err != nil {
		return nil, fmt.Errorf("airtable.Find: Error getting struct field json names: %v", err)
	}

	if slices.Contains(fieldNames, field) {
			return t.WithFilter(fmt.Sprintf("{%s} = '%s'", field, value)).List()
		}

	return nil, fmt.Errorf("airtable.Find: Field '%s' not found in schema %v", field, schema)
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
	u, err := url.JoinPath(baseUrl, baseId, tableId)
	if err != nil {
		return ""
	}
	return u
}
