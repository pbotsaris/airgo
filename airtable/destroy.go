package airtable

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Antfood/airgo/retry"
	"github.com/Antfood/airgo/utils"
)

type destroyedRecord struct {
   BaseId  string `json:"baseId"`
   TableId string `json:"tableId"`
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

type destroyResp struct {
	Error            map[string]string
	DestroyedRecords []*destroyedRecord `json:"records"`
}

func (d destroyResp) Err() map[string]string {
	return d.Error
}

func destroy[T any](ctx context.Context, deleteUrl string, records ...*Record[T]) ([]*destroyedRecord, error) {

	var destroyed []*destroyedRecord
	var resp destroyResp

	if client == nil {
		return destroyed, fmt.Errorf("airtable.destroy: Undefined client. Use airtable.Init before request")
	}

	if err := verifyRecordId(records); err != nil {
		return destroyed, err
	}

	recordIds := utils.Map(records, func(r *Record[T]) string { return r.Id })

	query := &queryBuilder{}
	query.NewWithUrl(deleteUrl)
	query.AddRecordIds(recordIds...)
	queryUrl := query.Flush()

	err := retry.DoCtx(ctx, func() error {
		httpReq, err := newHttpRequest(ctx, http.MethodDelete, queryUrl, nil)
		if err != nil {
			return fmt.Errorf("airtable.destroy: Failed to create http request: %v", err)
		}
		return makeRequest(client, httpReq, &resp)
	})

	if err != nil {
		return destroyed, fmt.Errorf("airtable.Destroy: %v", err)
	}

	return resp.DestroyedRecords, nil
}

func verifyRecordId[T any](records []*Record[T]) error {
	for _, record := range records {
		if record.Id == "" {
			return fmt.Errorf("airtable.destroy: Undefined record id: %v", record)
		}
	}
	return nil
}
