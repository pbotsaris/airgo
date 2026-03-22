package airtable

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Antfood/airgo/retry"
)

type ResponseBody[T any] struct {
	Records Records[T] `json:"records"`
}

func update[T any](ctx context.Context, url string, records Records[T], req ...updateRequest) error {
	var typecast bool

	if len(req) > 0 {
		typecast = req[0].Typecast
	}

	return upsert(ctx, url, records, typecast, req...)
}

func insert[T any](ctx context.Context, url string, records Records[T], req ...createRequest) error {
	var typecast bool

	if len(req) > 0 {
		typecast = req[0].Typecast
	}

	return upsert(ctx, url, records, typecast, req...)
}

func replace[T any](ctx context.Context, url string, records Records[T], req ...replaceRequest) error {
	var typecast bool

	if len(req) > 0 {
		typecast = req[0].Typecast
	}

	return upsert(ctx, url, records, typecast, req...)
}

func upsert[T any, R createRequest | updateRequest | replaceRequest](ctx context.Context, url string, records Records[T], typecast bool, req ...R) error {

	op := getOperation(req)

	if client == nil {
		return NewConfigError(op, "client not configured; call SetToken or Configure first")
	}

	reqBody := airtableBody[R]{
		Records:  req,
		Typecast: typecast,
	}

	respBody := ResponseBody[T]{
		Records: records,
	}

	jsonBody, err := json.Marshal(reqBody)

	if err != nil {
		return &Error{Op: op, Message: "failed to marshal record", Err: err}
	}

	return retry.DoCtx(ctx, func() error {
		httpReq, err := newHttpRequest(ctx, getMethod(req), url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return &Error{Op: op, Message: "failed to create http request", Err: err}
		}

		resp, err := client.Do(httpReq)
		if err != nil {
			return &Error{Op: op, Message: "failed to make request", Err: err}
		}
		defer resp.Body.Close()

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			return &retry.HTTPError{StatusCode: resp.StatusCode}
		}

		if resp.StatusCode != http.StatusOK {
			return handleError(resp, op)
		}

		respBodyData, err := io.ReadAll(resp.Body)
		if err != nil {
			return &Error{Op: op, Message: "failed to read response body", Err: err}
		}

		if err := json.Unmarshal(respBodyData, &respBody); err != nil {
			return &Error{Op: op, Message: "failed to unmarshal response", Err: err}
		}

		return nil
	})
}


func getOperation(records any) Operation {
	switch records.(type) {
	case []createRequest:
		return OpCreate
	case []updateRequest:
		return OpUpdate
	case []replaceRequest:
		return OpReplace
	default:
		return OpCreate
	}
}

func getMethod(records any) string {
	switch records.(type) {
	case []createRequest:
		return http.MethodPost
	case []updateRequest:
		return http.MethodPatch
	case []replaceRequest:
		return http.MethodPut
	default:
		return http.MethodPost
	}
}

func handleError(resp *http.Response, op Operation) error {
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Error{Op: op, StatusCode: resp.StatusCode, Message: "failed to read error response", Err: err}
	}

	return ParseAPIError(op, resp.StatusCode, bodyData)
}
