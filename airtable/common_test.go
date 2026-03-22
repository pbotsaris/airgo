package airtable

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	. "github.com/Antfood/airgo/testutils/testutils"
)

func TestNewHttpRequest(t *testing.T) {

	json, err := os.ReadFile("./testdata/get_recording.json")
	Ok(t, err)
	url := "http://www.example.com"
	contentType := "application/json"

	httpReq, err := newHttpRequest(context.Background(), http.MethodGet, url, bytes.NewBuffer(json))
	Ok(t, err)
	Assert(t, http.MethodGet == httpReq.Method, "Expected '%s', got '%s'", http.MethodGet, httpReq.Method)
	Assert(t, httpReq.URL.String() == url, "Expected '%s', got '%s'", url, httpReq.URL.String())
	Assert(t, httpReq.Header.Get("Content-Type") == contentType, "Expected '%s', got '%s'", contentType, httpReq.Header.Get("Content-Type"))
	Assert(t, httpReq.Header.Get("Authorization") == bearer+config.Token, "Expected '%s', got '%s'", bearer+config.Token, httpReq.Header.Get("Authorization"))

	b, err := io.ReadAll(httpReq.Body)
	Ok(t, err)

	j, err := io.ReadAll(bytes.NewBuffer(json))
	Ok(t, err)

	Equals(t, b, j)
}

func TestMakeRequest(t *testing.T) {
	jsonData, err := os.ReadFile("./testdata/list_recordings.json")
	Ok(t, err)

	client := newMockClient(200, jsonData, nil)

	httpReq, err := newHttpRequest(context.Background(), http.MethodGet, "https://mock.com", bytes.NewBuffer(jsonData))
	Ok(t, err)

	resp := new(listResp[TestListRecordSchema])

	err = makeRequest(client, httpReq, resp)
	Assert(t, err == nil, "Expected no error, got '%s'", err)

	client = newMockClient(404, jsonData, fmt.Errorf("404 Not Found"))

	err = makeRequest(client, httpReq, resp)
	Assert(t, err != nil, "Expected error, got '%s'", err)

	want := "Error making request: 404 Not Found"
	Assert(t, err.Error() == want, "Expected '%s', got '%s'", want, err.Error())
}
