package airtable

import (
	"context"
	"fmt"

	"github.com/Antfood/airgo/utils"
)

/*
NewRecord creates and returns a new instance of Record for a given baseId and tableId.
*/
func NewRecord[T any](baseId string, tableId string) *Record[T] {
	return &Record[T]{BaseId: baseId, TableId: tableId}
}

/*
Record represents a single record within an Airtable table. The structure of the record's fields is defined by the generic type T.

For instance, if your Airtable table has fields 'Name' and 'Age', you might define T as:

	type MySchema struct {
	   Name string `json:"name"`
	   Age  int    `json:"age"`
	}

The record then becomes Record[MySchema].
*/
type Record[T any] struct {
	Id          string `json:"id"`
	CreatedTime string `json:"createdTime,omitempty"`
	TableId     string `json:"-"`
	BaseId      string `json:"-"`
	Error       struct {
		Message string `json:"message"`
	} `json:"error"`
	Fields   T    `json:"fields"`
	Typecast bool `json:"-"`
	// CommentCount is populated when WithRecordMetadata("commentCount") is used
	CommentCount *int `json:"commentCount,omitempty"`
}

/*
SaveCtx updates the existing record in the Airtable.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	record.Fields.Name = "Updated Name"
	err = record.SaveCtx(ctx)
*/
func (r *Record[T]) SaveCtx(ctx context.Context) error {
	url := fmt.Sprintf("%s/%s/%s", config.EndpointUrl, r.BaseId, r.TableId)

	fields, err := utils.StructJsonToMap(r.Fields, utils.WithIgnore())

	if err != nil {
		return fmt.Errorf("airtable.Record.Save: %s", err.Error())
	}

	/* If the record has no Id, it's a new record */

	if r.Id == "" {
		return insert(ctx, url, Records[T]{r}, createRequest{Fields: fields, Typecast: r.Typecast})
	}

	return update(ctx, url, Records[T]{r}, updateRequest{Id: r.Id, Fields: fields, Typecast: r.Typecast})
}

/*
Save updates the existing record in the Airtable.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	record, err := table.Get("rec123")

	// Update record's fields
	record.Fields.Name = "Updated Name"

	// Save updated record
	err = record.Save()
*/
func (r *Record[T]) Save() error {
	return r.SaveCtx(context.Background())
}

/*
DestroyCtx deletes the current record from the Airtable and returns details of the destroyed record.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	destroyed, err := record.DestroyCtx(ctx)
*/
func (r *Record[T]) DestroyCtx(ctx context.Context) (*destroyedRecord, error) {
	url := createRequestUrl(r.BaseId, r.TableId)

	records, err := destroy(ctx, url, r)

	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("airtable.Record.Destroy: Delete returned no records")
	}

	records[0].Id = r.Id
	records[0].TableId = r.TableId
	records[0].BaseId = r.BaseId

	return records[0], err
}

/*
Destroy deletes the current record from the Airtable and returns details of the destroyed record.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	record, err := table.Get("rec123")

	// Delete the record
	destroyed, err := record.Destroy()
*/
func (r *Record[T]) Destroy() (*destroyedRecord, error) {
	return r.DestroyCtx(context.Background())
}

/*
ReplaceCtx performs a full replacement of the record in Airtable using PUT.
Unlike SaveCtx (which uses PATCH for updates), ReplaceCtx will clear any fields not provided.
It accepts a context for cancellation and timeout control.

Example:

	ctx := context.Background()
	record.Fields = MySchema{Name: "New Name"} // Age will be cleared
	err = record.ReplaceCtx(ctx)
*/
func (r *Record[T]) ReplaceCtx(ctx context.Context) error {
	if r.Id == "" {
		return fmt.Errorf("airtable.Record.Replace: Cannot replace record without Id")
	}

	url := fmt.Sprintf("%s/%s/%s", config.EndpointUrl, r.BaseId, r.TableId)

	fields, err := utils.StructJsonToMap(r.Fields, utils.WithoutIgnore())
	if err != nil {
		return fmt.Errorf("airtable.Record.Replace: %s", err.Error())
	}

	return replace(ctx, url, Records[T]{r}, replaceRequest{Id: r.Id, Fields: fields, Typecast: r.Typecast})
}

/*
Replace performs a full replacement of the record in Airtable using PUT.
Unlike Save (which uses PATCH for updates), Replace will clear any fields not provided.

Example:

	table := airtable.NewTable[MySchema]("baseId", "tableId")
	record, err := table.Get("rec123")

	// Replace entire record - fields not set will be cleared
	record.Fields = MySchema{Name: "New Name"} // Age will be cleared

	err = record.Replace()
*/
func (r *Record[T]) Replace() error {
	return r.ReplaceCtx(context.Background())
}

/*
WithId sets the Id of the current record and returns the record.

Example:

   table := airtable.NewTable[MySchema]("baseId", "tableId")
   record := table.NewRecord().WithId("rec123")

*/

func (r *Record[T]) WithId(id string) *Record[T] {
	r.Id = id
	return r
}

/*
Records is a slice of Record instances. This structure represents multiple records fetched from or to be processed in Airtable.

Example:

	// Assuming MySchema is defined as before
	table := airtable.NewTable[MySchema]("baseId", "tableId")
	records, err := table.WithLimit(10).List()

	// Now records is of type Records[MySchema]
*/
type Records[T any] []*Record[T]
