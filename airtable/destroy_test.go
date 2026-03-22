package airtable

import (
	"fmt"
	"os"
	"testing"

	. "github.com/Antfood/airgo/testutils/testutils"
)

type testDestroySchema struct {
	Name string `json:"name"`
}

func TestDestroy(t *testing.T) {
	t.Run("testTableDestroy", testTableDestroy)
	t.Run("testRecordDestroy", testRecordDestroy)
	t.Run("testVerifyRecordId", testVerifyRecordId)
}

func testTableDestroy(t *testing.T) {

	baseId := "base_id"
	tableId := "table_id"

	jsonData, err := os.ReadFile("testdata/destroy_record.json")
	Ok(t, err)

	client := newMockClient(200, jsonData, nil)
	Configure(client, "mock_token")

	records := make([]*Record[testDestroySchema], 2)

	for i := range records {
		records[i] = &Record[testDestroySchema]{}
		records[i].Id = fmt.Sprintf("rec-%d", i)
		records[i].Fields.Name = fmt.Sprintf("name-%d", i)
	}

	table := NewTable[testDestroySchema](baseId, tableId)

	resp, err := table.Destroy(records...)
	Assert(t, err == nil, "Expected no error, got %v", err)

	want := make([]*destroyedRecord, 2)

	for i := range want {
		want[i] = &destroyedRecord{}
		want[i].BaseId = baseId
		want[i].TableId = tableId
		want[i].Id = fmt.Sprintf("rec-%d", i)
		want[i].Deleted = true
	}

	for i := range resp {
		Equals(t, want[i], resp[i])
	}

	// test error
	client = newMockClient(200, jsonData, fmt.Errorf("mock error"))
	Configure(client, "mock_token")

	_, err = table.Destroy(records...)

	wantError := "airtable.Destroy: Error making request: mock error"

	Assert(t, err != nil, "Expected error, got nil")
	Assert(t, err.Error() == wantError, "Expected error %s, got %v", wantError, err)
}

func testRecordDestroy(t *testing.T) {

	baseId := "base_id"
	tableId := "table_id"

	jsonData, err := os.ReadFile("testdata/destroy_one_record.json")
	Ok(t, err)
   client := newMockClient(200, jsonData, nil)

	Configure(client, "mock_token")

	record := NewRecord[testDestroySchema](baseId, tableId)
	record.Id = "rec-0"

	resp, err := record.Destroy()
	Assert(t, err == nil, "Expected no error, got %v", err)

	want := &destroyedRecord{}
	want.BaseId = baseId
	want.TableId = tableId
	want.Id = "rec-0"
	want.Deleted = true

	Equals(t, want, resp)
}

func testVerifyRecordId(t *testing.T) {

	ids := []string{"rec123", "rec456", "rec789"}
	names := []string{"name1", "name2", "name3"}
	records := make([]*Record[testDestroySchema], len(ids))

	for i, id := range ids {
		records[i] = &Record[testDestroySchema]{}
		records[i].Id = id
		records[i].Fields.Name = names[i]
	}

	err := verifyRecordId(records)
	Assert(t, err == nil, "Expected no error, got %v", err)

}
