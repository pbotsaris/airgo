package airtable

import (
	"fmt"
	"net/http"

	"github.com/Antfood/airgo/retry"
	"github.com/Antfood/airgo/utils"
)

type listResp[T any] struct {
	Error   map[string]string
	Records []*Record[T]
	Offset  string
}

func (l listResp[T]) Err() map[string]string {
	return l.Error
}

func list[T any](url string, opts Options) (Records[T], error) {
	var records Records[T]

	if client == nil {
		return records, fmt.Errorf("airtable.List: Undefined client. airtable.Init before request.")
	}


   query, err := newQuery[T](url, opts)
	if err != nil {
		return records, fmt.Errorf("airtable.List: Error creating query: %s", err.Error())
	}

	for {
		var resp listResp[T]
		queryUrl := query.Flush()

		err := retry.Do(func() error {
			httpReq, err := newHttpRequest(http.MethodGet, queryUrl, nil)
			if err != nil {
				return fmt.Errorf("airtable.List: Failed to create http request: %v", err)
			}
			return makeRequest(client, httpReq, &resp)
		})

		if err != nil {
			return records, fmt.Errorf("airtable.List: %v", err)
		}

		records = append(records, resp.Records...)

		if resp.Offset == "" {
			break
		}

		query.AddOffset(resp.Offset)
	}

	return records, nil
}

func newQuery[T any](url string, opts Options) (*queryBuilder, error) {

	query := &queryBuilder{}
	var schema T

	fieldNames, err := utils.GetStructFieldJsonNames(schema)

	if err != nil {
		return query, err
	}

	query.NewWithUrl(url)
	query.AddFields(fieldNames)
	query.AddPageSize(getPageSize(opts.Limit))
	query.AddMaxRecords(getMaxRecords(opts.Limit))

	if !opts.Sort.Empty() {
		query.AddSort(opts.Sort)
	}

	if opts.Filter != "" {
		query.AddFilterByFormula(opts.Filter)
	}

	return query, nil
}

func getPageSize(limit int) int {

	if limit > 0 && limit < maxPageSize {
		return limit
	}
	return maxPageSize
}

func getMaxRecords(limit int) int {

	if limit > 0 {
		return limit
	}

	return maxPageSize
}
