package airtable

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	. "github.com/Antfood/airgo/testutils/testutils"
)

type testSaveRecordSchema struct {
	Descriptor string `json:"rec.Descriptor"`
	Title      string `json:"rec.Recording Title"`
}

func TestUpsert(t *testing.T) {
	t.Run("testCreateReturnReturnId", testCreateReturnReturnId)
	t.Run("testUpdateMultipleRecords", testUpdateMultipleRecords)
	t.Run("testGetMsgPrefix", testGetPrefix)
	t.Run("testGetMethod", testGetMethod)
	t.Run("testHandleError", testHandleError)
}

func testCreateReturnReturnId(t *testing.T) {
	saveJson, err := os.ReadFile("testdata/save_recording.json")
	Ok(t, err)

	client := newMockClient(200, saveJson, nil)
	Configure(client, "mock_token")

	record := NewRecord[testSaveRecordSchema]("base_id", "table_id")

	err = record.Save()
	Ok(t, err)

	want := "rec00Z2SwAoCiPZuc"
	Assert(t, record.Id == want, "Expected create to return id '%s', got '%s'", want, record.Id)
}

func testUpdateMultipleRecords(t *testing.T) {
	jsonData, err := os.ReadFile("testdata/list_recordings.json")
	Ok(t, err)

	saveJson, err := os.ReadFile("testdata/save_recording.json")
	Ok(t, err)

	records := getMockRecord(t, jsonData)

	client := newMockClient(200, saveJson, nil)
	Configure(client, "mock_token")

	table := NewTable[testRecordSchema]("base_id", "table_id")

	err = table.Update(records...)
	Assert(t, err == nil, "Expected no error, got '%s'", err)
}

func testGetPrefix(t *testing.T) {
	msg := getMsgPrefix([]createRequest{})
	Assert(t, msg == createPrefix, "Expected '%s', got '%s'", createPrefix, msg)

	msg = getMsgPrefix([]updateRequest{})
	Assert(t, msg == updatePrefix, "Expected '%s', got '%s'", createPrefix, msg)
}

func testGetMethod(t *testing.T) {
	method := getMethod([]createRequest{})
	Assert(t, method == http.MethodPost, "Expected '%s', got '%s'", http.MethodPost, method)

	method = getMethod([]updateRequest{})
	Assert(t, method == http.MethodPatch, "Expected '%s', got '%s'", http.MethodPatch, method)
}

func TestGetVerb(t *testing.T) {
	verb := getVerb(createPrefix)
	Assert(t, verb == "create", "Expected '%s', got '%s'", "create", verb)

	verb = getVerb(updatePrefix)
	Assert(t, verb == "update", "Expected '%s', got '%s'", "update", verb)
}

func testHandleError(t *testing.T) {
	msg := "Not Found"
	inBody := testErrorBody{}
	inBody.Error.Message = msg

	jsonBody, err := json.Marshal(inBody)
	Ok(t, err)

	client := newMockClient(
		404,
		jsonBody,
		errors.New(msg))

	req, err := http.NewRequest(http.MethodPost, "http://example.com", nil)
	Ok(t, err)

	resp, _ := client.Do(req)
	Assert(t, resp.StatusCode == 404, "Expected '%d', got '%d'", 404, resp.StatusCode)

	err = handleError(resp, createPrefix)
	Assert(t, err != nil, "Expected error, got nil")

	errMsg := "airtable.Create: Failed to create with status: '404' and message: 'Not Found'"
	Assert(t, err.Error() == errMsg, "Expected '%s', got '%s'", errMsg, err.Error())
}

func getMockRecord(t *testing.T, json []byte) Records[testRecordSchema] {
	client := newMockClient(200, json, nil)
	Configure(client, "mock_token")
	table := NewTable[testRecordSchema]("base_id", "table_id")

	records, err := table.List()
	Ok(t, err)

	return records
}
