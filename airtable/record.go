package airtable

import (
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
	Id      string `json:"id"`
	TableId string
	BaseId  string
	Error   struct {
		Message string `json:"message"`
	} `json:"error"`
	Fields T `json:"fields"`
   Typecast bool 
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
	url := fmt.Sprintf("%s/%s/%s", baseUrl, r.BaseId, r.TableId)

	fields, err := utils.StructJsonToMap(r.Fields, utils.WithIgnore())

	if err != nil {
		return fmt.Errorf("airtable.Record.Save: %s", err.Error())
	}

	/* If the record has no Id, it's a new record */

	if r.Id == "" {
      return insert(url, Records[T]{r}, createRequest{Fields: fields, Typecast: r.Typecast})
	}

   return update(url, Records[T]{r}, updateRequest{Id: r.Id, Fields: fields, Typecast: r.Typecast})
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
	url := createRequestUrl(r.BaseId, r.TableId)

	records, err := destroy(url, r)

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
