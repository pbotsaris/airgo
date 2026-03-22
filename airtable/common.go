package airtable

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Antfood/airgo/retry"
)

/*
Options represents the parameters that will be provided to airtable upon request.

	Limit:          The number of records per page (max 100). Default is 100.
	MaxRecords:     The maximum total number of records to return across all pages. 0 means no limit.
	Filter:         A formula used to filter records. See https://support.airtable.com/docs/formula-field-reference
	Sort:           A slice of Sort structs. airtable.List will sort by the first Sort in the slice, then the second, and so on.
	Typecast:       If true, Airtable will try to convert string values to the appropriate cell type.
	Fields:         Only return data for the specified field names. If empty, returns all fields from schema.
	View:           The name or ID of a view to use for filtering/sorting.
	CellFormat:     The format for cell values: "json" (default) or "string".
	TimeZone:       The time zone to use for formatting dates (e.g., "America/New_York").
	UserLocale:     The locale to use for formatting (e.g., "en-US").
	RecordMetadata: Additional metadata to return (e.g., "commentCount").
*/
type Options struct {
	Limit          int
	MaxRecords     int
	Filter         string
	Sort           Sorts
	Typecast       bool
	Fields         []string
	View           string
	CellFormat     string
	TimeZone       string
	UserLocale     string
	RecordMetadata []string
}

/*
Sort is a structure for specifying how you want to sort the records returned from airtable.List.

	Field:     The field you want to sort by.
	Direction: The direction you want to sort by. Must be either "asc" or "desc".
*/
type Sort struct {
	Field     string
	Direction string
}

/*
A Slice of Sorts. airtable.List will sort by the first Sort in the slice, then the second, and so on.
*/
type Sorts []Sort

func (s Sorts) Empty() bool {
	return len(s) == 0
}

/* Private */

type updateRequest struct {
	Id       string         `json:"id"`
	Fields   map[string]any `json:"fields"`
	Typecast bool           `json:"-"`
}

type createRequest struct {
	Fields   map[string]any `json:"fields"`
	Typecast bool           `json:"-"`
}

type replaceRequest struct {
	Id       string         `json:"id"`
	Fields   map[string]any `json:"fields"`
	Typecast bool           `json:"-"`
}

type airtableBody[T updateRequest | createRequest | replaceRequest] struct {
	Records  []T  `json:"records"`
	Typecast bool `json:"typecast"`
}

type Responder interface {
	Err() map[string]string
}

type testErrorBody struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func newHttpRequest(ctx context.Context, verb string, url string, body io.Reader) (*http.Request, error) {
	httpReq, err := http.NewRequestWithContext(ctx, verb, url, body)

	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", bearer+config.Token)
	httpReq.Header.Set("Content-Type", "application/json")

	// Add custom headers from config
	for key, value := range config.CustomHeaders {
		httpReq.Header.Set(key, value)
	}

	return httpReq, nil
}

/* */

func makeRequest(client Client, httpReq *http.Request, r Responder) error {

	res, err := client.Do(httpReq)

	if err != nil {
		return fmt.Errorf("Error making request: %s", err.Error())
	}

	defer res.Body.Close()

	if res.StatusCode == 429 || res.StatusCode >= 500 {
		return &retry.HTTPError{StatusCode: res.StatusCode}
	}

	if err := json.NewDecoder(res.Body).Decode(r); err != nil {
		return fmt.Errorf("Error decoding response: %s", err.Error())
	}

	errMsg, ok := r.Err()["message"]

	if ok { // airtable has returned an error
		return fmt.Errorf("Airtable returned an error: '%s'", errMsg)
	}

	return nil
}
