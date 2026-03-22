package airtable

import (
	"os"
	"testing"

	. "github.com/Antfood/airgo/testutils/testutils"
)

type testTableSchema struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestTable(t *testing.T) {
	json, err := os.ReadFile("testdata/table_list_records.json")
	Ok(t, err)

	client := newMockClient(200, json, nil)
	Configure(client, "mock_key")

	t.Run("NewTable", testNewTable)
	t.Run("TableWithOptionsResetsAfterRequest", testTableWithOptionsResetsAfterRequest)
	t.Run("TestWithLimitFilterSort", testLimitFilterSort)
	t.Run("TestTableNewRecord", testTableNewRecord)

   
   client = newMockClient(200, json, nil)
	Configure(client, "mock_key")
	t.Run("TestTableFind", testTableFind)

}

func testNewTable(t *testing.T) {
	table := NewTable[testTableSchema]("base_id", "table_id")

	Assert(t, table.BaseId == "base_id", "Expected 'base_id, got '%s'", table.BaseId)
	Assert(t, table.TableId == "table_id", "Expected 'table_id', got '%s'", table.TableId)
	Equals(t, table.Options, defaultOptions)
}

func testTableWithOptionsResetsAfterRequest(t *testing.T) {

	table := NewTable[testTableSchema]("base_id", "table_id")

	sort := Sorts{
		{Field: "name", Direction: "asc"},
		{Field: "age", Direction: "asc"},
	}

	want := Options{Limit: 20,
		Filter: "{name} = 'John'",
		Sort:   sort,
	}

	table = table.WithOptions(want)
	Equals(t, table.Options, want)

	// ignore result as it doesn't matter for this test
	_, err := table.List()
	Ok(t, err)

	// options always resets to default after a request
	Equals(t, table.Options, defaultOptions)
}

func testLimitFilterSort(t *testing.T) {

	table := NewTable[testTableSchema]("base_id", "table_id")
	limit := 200
	filter := "{name} = 'John'"
	sort := Sorts{{"name", "desc"}, {"age", "asc"}}

	want := defaultOptions
	want.Limit = limit

	table = table.WithLimit(limit)
	Equals(t, table.Options, want)

	want.Filter = filter
	table = table.WithFilter(filter)
	Equals(t, table.Options, want)

	want.Sort = sort
	table = table.WithSort(sort)
	Equals(t, table.Options, want)

}

func testTableNewRecord(t *testing.T) {

	table := NewTable[testTableSchema]("base_id", "table_id")

	record := table.NewRecord()

	Assert(t, record != nil, "Expected record created from a table to not be nil")
	Assert(t, record.BaseId == table.BaseId, "Expected record to have base id '%s', got '%s'", table.BaseId, record.BaseId)
	Assert(t, record.TableId == table.TableId, "Expected record to have table id '%s', got '%s'", table.TableId, record.TableId)

	records := table.NewRecords(5)
	Assert(t, len(records) == 5, "Expected 5 records, got %d", len(records))

	for _, record := range records {
		Assert(t, record != nil, "Expected record created from a table to not be nil")
		Assert(t, record.BaseId == table.BaseId, "Expected record to have base id '%s', got '%s'", table.BaseId, record.BaseId)
		Assert(t, record.TableId == table.TableId, "Expected record to have table id '%s', got '%s'", table.TableId, record.TableId)
	}
}

func testTableFind(t *testing.T) {

	table := NewTable[testTableSchema]("base_id", "table_id")

	records, err := table.Find("name", "roger")

	Assert(t, err == nil, "Expected no error, got %s", err)
	Assert(t, len(records) > 0, "Expected at least one record, got %d", len(records))
	Assert(t, records[0].Fields.Name == "roger", "Expected record name to be 'roger', got '%s'", records[0].Fields.Name)
	Assert(t, records[0].Fields.Age == 20, "Expected record age to be 20, got %d", records[0].Fields.Age)

}
