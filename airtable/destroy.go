package airtable

import (
	"context"
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
		return destroyed, NewConfigError(OpDestroy, "client not configured; call SetToken or Configure first")
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
			return &Error{Op: OpDestroy, Message: "failed to create http request", Err: err}
		}
		return makeRequest(client, httpReq, &resp, OpDestroy)
	})

	if err != nil {
		return destroyed, err
	}

	return resp.DestroyedRecords, nil
}

func verifyRecordId[T any](records []*Record[T]) error {
	for _, record := range records {
		if record.Id == "" {
			return &ValidationError{
				Op:      OpDestroy,
				Message: "record ID required",
				Err:     ErrMissingRecordID,
			}
		}
	}
	return nil
}
