package airtable

import (
	"errors"
	"os"
	"testing"

	. "github.com/Antfood/airgo/testutils/testutils"
)

type TestListRecordSchema struct {
	Title      string `json:"rec.Recording Title"`
	Descriptor string `json:"rec.Descriptor"`
}

func TestList(t *testing.T) {
	t.Run("Valid List", testValidList)
	t.Run("Invalid List", testInvalidList)
	t.Run("Paginated List", testPaginatedList)
	t.Run("getPageSize", testGetPageSize)
	t.Run("getMaxRecords", testGetMaxRecords)
	t.Run("newQuery", testNewQuery)
}

func testValidList(t *testing.T) {

	json, err := os.ReadFile("./testdata/list_recordings.json")
	Ok(t, err)

	client := newMockClient(200, json, nil)
	Configure(client, "mock_token")

	table := NewTable[TestListRecordSchema]("Base_id", "Table_id")
	resp, err := table.List()

	Assert(t, err == nil, "Should not return an error: %s", err)

	want := []TestListRecordSchema{

		{
			Title:      "Drifted (Extended Mix)",
			Descriptor: "Extended Mix",
		},
		{
			Title:      "Morning (Original)",
			Descriptor: "Original",
		},
		{
			Title:      "Sad Meredith (Original)",
			Descriptor: "Original",
		},
	}

	for i, r := range resp {
		Equals(t, want[i], r.Fields)
	}
}

func testInvalidList(t *testing.T) {
	msg := "Bad List"
	listErr := "airtable.List: Error making request: "

	client := newMockClient(500, []byte(""), errors.New(msg))
	Configure(client, "mock_token")

	table := NewTable[TestListRecordSchema]("Base_id", "Table_id")
	_, err := table.List()

	want := listErr + msg

	Assert(t, err != nil, "Invalid list should return an error")
	Assert(t, err.Error() == want, "Expected '%s' error, got '%s'", want, err.Error())
}

func testPaginatedList(t *testing.T) {
	json1, err := os.ReadFile("./testdata/list_paginate_recordings.json")
	Ok(t, err)
	json2, err := os.ReadFile("./testdata/list_recordings.json")
	Ok(t, err)

	client := newMockClientPaginate(200, [][]byte{json1, json2}, nil)
	Configure(client, "mock_token")

	table := NewTable[TestListRecordSchema]("Base_id", "Table_id")
	resp, err := table.List()

	want := []string{
		"Ride Of The Valkyries (Synthesizer)", "Acapulco (Original)",
		"Drifted (Extended Mix)", "Morning (Original)", "Sad Meredith (Original)",
	}

	Assert(t, err == nil, "Should not return an error: %s", err)
	Assert(t, len(resp) == 5, "Should return 5 records: %d", len(resp))

	for i, r := range resp {
		Assert(t, r.Fields.Title == want[i], "Expected '%s', got '%s'", want[i], r.Fields.Title)
	}
}

func testNewQuery(t *testing.T) {

	opts := Options{
		Sort:   Sorts{{Field: "some_field", Direction: "asc"}},
		Filter: "",
	}

	url := createRequestUrl("base_id", "table_id")

	q, err := newQuery[TestListRecordSchema](url, opts)
	Assert(t, err == nil, "Should not return an error: %s", err)

	have := q.Flush()
	want := "https://api.airtable.com/v0/base_id/table_id?fields%5B%5D=rec.Recording+Title&fields%5B%5D=rec.Descriptor&pageSize=100&maxRecords=100&sort%5B0%5D%5Bfield%5D=some_field&sort%5B0%5D%5Bdirection%5D=asc"

	Assert(t, want == have, "Expected '%s', got '%s'", want, have)

	/* with filter */
	opts = Options{
		Sort:   Sorts{{Field: "some_field", Direction: "asc"}},
		Filter: "{rec.Recording Title} = '2 Worlds (Original)'",
	}

	q, err = newQuery[TestListRecordSchema](url, opts)
	Assert(t, err == nil, "Should not return an error: %s", err)

	have = q.Flush()
	want = "https://api.airtable.com/v0/base_id/table_id?fields%5B%5D=rec.Recording+Title&fields%5B%5D=rec.Descriptor&pageSize=100&maxRecords=100&sort%5B0%5D%5Bfield%5D=some_field&sort%5B0%5D%5Bdirection%5D=asc&filterByFormula=%7Brec.Recording+Title%7D+%3D+%272+Worlds+%28Original%29%27"

	Assert(t, want == have, "Expected '%s', got '%s'", want, have)

	/* Double Sort & Limit */

	opts = Options{
		Limit: 10,
		Sort: Sorts{
			{
				Field:     "rec.Recording Title",
				Direction: "asc",
			},
			{
				Field:     "rec.Descriptor",
				Direction: "asc"},
		},
	}

	q, err = newQuery[TestListRecordSchema](url, opts)
	Assert(t, err == nil, "Should not return an error: '%v'", err)

	have = q.Flush()
	want = "https://api.airtable.com/v0/base_id/table_id?fields%5B%5D=rec.Recording+Title&fields%5B%5D=rec.Descriptor&pageSize=10&maxRecords=10&sort%5B0%5D%5Bfield%5D=rec.Recording+Title&sort%5B0%5D%5Bdirection%5D=asc&sort%5B1%5D%5Bfield%5D=rec.Descriptor&sort%5B1%5D%5Bdirection%5D=asc"

	Assert(t, want == have, "Expected '%s', got '%s'", want, have)
}

func testGetPageSize(t *testing.T) {
	r := getPageSize(10)
	Assert(t, r == 10, "Expected '10', got '%d'", r)

	r = getPageSize(0)
	Assert(t, r == maxPageSize, "Expected '%d', got '%d'", maxPageSize, r)

	r = getPageSize(maxPageSize + 1)
	Assert(t, r == maxPageSize, "Expected '%d', got '%d'", maxPageSize, r)
}

func testGetMaxRecords(t *testing.T) {
	r := getMaxRecords(10)
	Assert(t, r == 10, "Expected '10', got '%d'", r)

	r = getMaxRecords(0)
	Assert(t, r == maxPageSize, "Expected '%d', got '%d'", maxPageSize, r)
}
