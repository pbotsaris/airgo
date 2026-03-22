package airtable

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Antfood/airgo/retry"
)

/*
Options represents the parameters that will be provided to airtable upon request.

	Limit:   The maximum number of records returned in each request. Must be less than or equal to 100. Default is 100
	Filter:  A formula used to filter records. See https://support.airtable.com/docs/formula-field-reference for more info
	Sort:    A slice of Sort structs. airtable.List will sort by the first Sort in the slice, then the second, and so on
*/
type Options struct {
	Limit    int
	Filter   string
	Sort     Sorts
	Typecast bool
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
	Id       string                 `json:"id"` // must have and id
	Fields   map[string]any `json:"fields"`
	Typecast bool                   `json:"-"`
}

type createRequest struct {
	Fields   map[string]any `json:"fields"`
	Typecast bool                   `json:"-"`
}

type airtableBody[T updateRequest | createRequest] struct {
	Records []T `json:"records"`
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

func newHttpRequest(verb string, url string, body io.Reader) (*http.Request, error) {
	httpReq, err := http.NewRequest(verb, url, body)

	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", bearer+token)
	httpReq.Header.Set("Content-Type", "application/json")

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
