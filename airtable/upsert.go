package airtable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Antfood/airgo/retry"
)

const (
	updatePrefix = "airtable.Update:"
	createPrefix = "airtable.Create:"
)

type ResponseBody[T any] struct {
	Records Records[T] `json:"records"`
}

func update[T any](url string, records Records[T], req ...updateRequest) error {
  var typecast bool

  if len(req) > 0 {
     typecast = req[0].Typecast
  }
   
	return upsert(url, records, typecast, req...)
}

func insert[T any](url string, records Records[T], req ...createRequest) error {
 var typecast bool

  if len(req) > 0 {
     typecast = req[0].Typecast
  }

	return upsert(url, records, typecast, req...)
}

func upsert[T any, R createRequest | updateRequest](url string, records Records[T], typecast bool, req ...R) error {

	prefix := getMsgPrefix(req)

	if client == nil {
		return fmt.Errorf("%s: Undefined client. airtable.Init before request", prefix)
	}

	reqBody := airtableBody[R]{
		Records: req,
      Typecast: typecast,
	}

	respBody := ResponseBody[T]{
		Records: records,
	}

	jsonBody, err := json.Marshal(reqBody)

	if err != nil {
		return fmt.Errorf("%s Could not Marshal record:  %s", prefix, err.Error())
	}

	return retry.Do(func() error {
		httpReq, err := newHttpRequest(getMethod(req), url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("%s Could not create http request:  %v", prefix, err)
		}

		resp, err := client.Do(httpReq)
		if err != nil {
			return fmt.Errorf("%s failed to make request:  %v", prefix, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			return &retry.HTTPError{StatusCode: resp.StatusCode}
		}

		if resp.StatusCode != http.StatusOK {
			return handleError(resp, prefix)
		}

		respBodyData, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%s Could not read response body:  %v", prefix, err)
		}

		if err := json.Unmarshal(respBodyData, &respBody); err != nil {
			return fmt.Errorf("%s Could not unmarshal response:  %v", prefix, err)
		}

		return nil
	})
}

func getMsgPrefix(records any) string {
	switch records.(type) {
	case []createRequest:
		return createPrefix
	case []updateRequest:
		return updatePrefix
	default:
		return createPrefix
	}
}

func getMethod(records interface{}) string {
	switch records.(type) {
	case []createRequest:
		return http.MethodPost
	case []updateRequest:
		return http.MethodPatch
	default:
		return http.MethodPost
	}
}

func handleError(resp *http.Response, prefix string) error {
	verb := getVerb(prefix)

	var atErrorResponse map[string]interface{}

	bodyData, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("%s.handleError: Could not read response body:  %v", prefix, err)
	}

	if err := json.Unmarshal(bodyData, &atErrorResponse); err != nil {
		return fmt.Errorf("%s.handleError: Failed to %s. Could not unmarshal response '%s': '%v'",
			prefix, verb, string(bodyData), err)
	}

	return fmt.Errorf("%s Failed to %s with status: '%d' and message: '%s'",
		prefix,
		verb,
		resp.StatusCode,
		getErrorMessage(atErrorResponse))
}

func getErrorMessage(err interface{}) string {
    switch errorType := err.(type) {

    case map[string]interface{}:
        if errorMsg, ok := errorType["message"].(string); ok {
            return errorMsg
        }

        if nestedError, ok := errorType["error"]; ok {
            return getErrorMessage(nestedError)
        }

        return fmt.Sprintf("Unknown error format: %v", errorType)

    case string:
        return errorType

    case error:
        return errorType.Error()

    default:
        return fmt.Sprintf("Unhandled error type %T", err)
    }
}

func getVerb(prefix string) string {
	if prefix == updatePrefix {
		return "update"
	}

	return "create"
}
