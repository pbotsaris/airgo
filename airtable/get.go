package airtable

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Antfood/airgo/retry"
)

func get[T any](ctx context.Context, getUrl string, record Record[T]) (Record[T], error) {

	if client == nil {
		return record, fmt.Errorf("airtable.Get: Undefined client. Use airtable.Init before request")
	}

	if record.Id == "" {
		return record, fmt.Errorf("airtable.Get: Undefined record id")
	}

	url := fmt.Sprintf("%s/%s", getUrl, record.Id)

	err := retry.DoCtx(ctx, func() error {
		httpReq, err := newHttpRequest(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		res, err := client.Do(httpReq)
		if err != nil {
			return fmt.Errorf("airtable.Get: %s", err.Error())
		}
		defer res.Body.Close()

		if res.StatusCode == 429 || res.StatusCode >= 500 {
			return &retry.HTTPError{StatusCode: res.StatusCode}
		}

		if err := json.NewDecoder(res.Body).Decode(&record); err != nil {
			return fmt.Errorf("airtable.Get: %s", err.Error())
		}

		if record.Error.Message != "" {
			return fmt.Errorf("airtable.Get: %s", record.Error)
		}

		return nil
	})

	return record, err
}
