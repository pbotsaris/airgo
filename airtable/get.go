package airtable

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Antfood/airgo/retry"
)

func get[T any](ctx context.Context, getUrl string, record Record[T]) (Record[T], error) {

	if client == nil {
		return record, NewConfigError(OpGet, "client not configured; call SetToken or Configure first")
	}

	if record.Id == "" {
		return record, &ValidationError{
			Op:      OpGet,
			Message: "record ID required",
			Err:     ErrMissingRecordID,
		}
	}

	url := getUrl + "/" + record.Id

	err := retry.DoCtx(ctx, func() error {
		httpReq, err := newHttpRequest(ctx, http.MethodGet, url, nil)
		if err != nil {
			return &Error{Op: OpGet, Message: "failed to create http request", Err: err}
		}

		res, err := client.Do(httpReq)
		if err != nil {
			return &Error{Op: OpGet, Message: "failed to make request", Err: err}
		}
		defer res.Body.Close()

		if res.StatusCode == 429 || res.StatusCode >= 500 {
			return &retry.HTTPError{StatusCode: res.StatusCode}
		}

		bodyData, err := io.ReadAll(res.Body)
		if err != nil {
			return &Error{Op: OpGet, Message: "failed to read response body", Err: err}
		}

		if err := json.Unmarshal(bodyData, &record); err != nil {
			return &Error{Op: OpGet, Message: "failed to decode response", Err: err}
		}

		if record.Error.Message != "" {
			return ParseAPIError(OpGet, res.StatusCode, bodyData)
		}

		return nil
	})

	return record, err
}
