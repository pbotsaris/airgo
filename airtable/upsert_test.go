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
	t.Run("testGetOperation", testGetOperation)
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

func testGetOperation(t *testing.T) {
	op := getOperation([]createRequest{})
	Assert(t, op == OpCreate, "Expected '%s', got '%s'", OpCreate, op)

	op = getOperation([]updateRequest{})
	Assert(t, op == OpUpdate, "Expected '%s', got '%s'", OpUpdate, op)

	op = getOperation([]replaceRequest{})
	Assert(t, op == OpReplace, "Expected '%s', got '%s'", OpReplace, op)
}

func testGetMethod(t *testing.T) {
	method := getMethod([]createRequest{})
	Assert(t, method == http.MethodPost, "Expected '%s', got '%s'", http.MethodPost, method)

	method = getMethod([]updateRequest{})
	Assert(t, method == http.MethodPatch, "Expected '%s', got '%s'", http.MethodPatch, method)

	method = getMethod([]replaceRequest{})
	Assert(t, method == http.MethodPut, "Expected '%s', got '%s'", http.MethodPut, method)
}

func testHandleError(t *testing.T) {
	msg := "Not Found"
	errType := "NOT_FOUND"

	// Create a proper Airtable error response
	errBody := map[string]any{
		"error": map[string]any{
			"type":    errType,
			"message": msg,
		},
	}

	jsonBody, err := json.Marshal(errBody)
	Ok(t, err)

	client := newMockClient(
		404,
		jsonBody,
		errors.New(msg))

	req, err := http.NewRequest(http.MethodPost, "http://example.com", nil)
	Ok(t, err)

	resp, _ := client.Do(req)
	Assert(t, resp.StatusCode == 404, "Expected '%d', got '%d'", 404, resp.StatusCode)

	err = handleError(resp, OpCreate)
	Assert(t, err != nil, "Expected error, got nil")

	// Check that it's an APIError with the right fields
	var apiErr *APIError
	Assert(t, errors.As(err, &apiErr), "Expected error to be APIError")
	Assert(t, apiErr.StatusCode == 404, "Expected status 404, got %d", apiErr.StatusCode)
	Assert(t, string(apiErr.Type) == errType, "Expected type '%s', got '%s'", errType, apiErr.Type)
	Assert(t, apiErr.Message == msg, "Expected message '%s', got '%s'", msg, apiErr.Message)
}

func getMockRecord(t *testing.T, json []byte) Records[testRecordSchema] {
	client := newMockClient(200, json, nil)
	Configure(client, "mock_token")
	table := NewTable[testRecordSchema]("base_id", "table_id")

	records, err := table.List()
	Ok(t, err)

	return records
}
