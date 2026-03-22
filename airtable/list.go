package airtable

import (
	"bytes"
	"context"
	"encoding/json"
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

// listRequestBody represents the POST body for large queries
type listRequestBody struct {
	Fields          []string        `json:"fields,omitempty"`
	FilterByFormula string          `json:"filterByFormula,omitempty"`
	MaxRecords      int             `json:"maxRecords,omitempty"`
	PageSize        int             `json:"pageSize,omitempty"`
	Sort            []listSortField `json:"sort,omitempty"`
	View            string          `json:"view,omitempty"`
	CellFormat      string          `json:"cellFormat,omitempty"`
	TimeZone        string          `json:"timeZone,omitempty"`
	UserLocale      string          `json:"userLocale,omitempty"`
	RecordMetadata  []string        `json:"recordMetadata,omitempty"`
	Offset          string          `json:"offset,omitempty"`
}

type listSortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction,omitempty"`
}

func list[T any](ctx context.Context, baseUrl string, opts Options) (Records[T], error) {
	var records Records[T]

	if client == nil {
		return records, fmt.Errorf("airtable.List: Undefined client. airtable.Init before request.")
	}

	// Build initial query to check URL length
	query, err := newQuery[T](baseUrl, opts)
	if err != nil {
		return records, fmt.Errorf("airtable.List: Error creating query: %s", err.Error())
	}

	queryUrl := query.Flush()
	usePOST := len(queryUrl) > config.MaxUrlLength

	// Get field names for POST request body
	var fieldNames []string
	if usePOST {
		if len(opts.Fields) > 0 {
			fieldNames = opts.Fields
		} else {
			var schema T
			fieldNames, err = utils.GetStructFieldJsonNames(schema)
			if err != nil {
				return records, fmt.Errorf("airtable.List: Error getting field names: %s", err.Error())
			}
		}
	}

	var offset string
	for {
		// Check context before each page request
		if err := ctx.Err(); err != nil {
			return records, err
		}

		var resp listResp[T]

		err := retry.DoCtx(ctx, func() error {
			var httpReq *http.Request
			var err error

			if usePOST {
				httpReq, err = createPOSTListRequest(ctx, baseUrl, opts, fieldNames, offset)
			} else {
				if offset != "" {
					query.AddOffset(offset)
					queryUrl = query.Flush()
				}
				httpReq, err = newHttpRequest(ctx, http.MethodGet, queryUrl, nil)
			}

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

		offset = resp.Offset
	}

	return records, nil
}

func createPOSTListRequest(ctx context.Context, baseUrl string, opts Options, fieldNames []string, offset string) (*http.Request, error) {
	body := listRequestBody{
		Fields:          fieldNames,
		FilterByFormula: opts.Filter,
		PageSize:        getPageSize(opts.Limit),
		View:            opts.View,
		CellFormat:      opts.CellFormat,
		TimeZone:        opts.TimeZone,
		UserLocale:      opts.UserLocale,
		RecordMetadata:  opts.RecordMetadata,
		Offset:          offset,
	}

	if opts.MaxRecords > 0 {
		body.MaxRecords = opts.MaxRecords
	} else {
		body.MaxRecords = getMaxRecords(opts.Limit)
	}

	// Convert sorts
	for _, s := range opts.Sort {
		body.Sort = append(body.Sort, listSortField(s))
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// POST to /listRecords endpoint
	listUrl := baseUrl + "/listRecords"
	return newHttpRequest(ctx, http.MethodPost, listUrl, bytes.NewBuffer(jsonBody))
}

func newQuery[T any](url string, opts Options) (*queryBuilder, error) {
	query := &queryBuilder{}

	// Use explicit fields if provided, otherwise derive from schema
	var fieldNames []string
	if len(opts.Fields) > 0 {
		fieldNames = opts.Fields
	} else {
		var schema T
		var err error
		fieldNames, err = utils.GetStructFieldJsonNames(schema)
		if err != nil {
			return query, err
		}
	}

	query.NewWithUrl(url)
	query.AddFields(fieldNames)
	query.AddPageSize(getPageSize(opts.Limit))

	// Use MaxRecords if specified, otherwise fall back to Limit
	if opts.MaxRecords > 0 {
		query.AddMaxRecords(opts.MaxRecords)
	} else {
		query.AddMaxRecords(getMaxRecords(opts.Limit))
	}

	if !opts.Sort.Empty() {
		query.AddSort(opts.Sort)
	}

	if opts.Filter != "" {
		query.AddFilterByFormula(opts.Filter)
	}

	// New query parameters
	query.AddView(opts.View)
	query.AddCellFormat(opts.CellFormat)
	query.AddTimeZone(opts.TimeZone)
	query.AddUserLocale(opts.UserLocale)
	query.AddRecordMetadata(opts.RecordMetadata)

	return query, nil
}

func getPageSize(limit int) int {
	if limit > 0 && limit < config.MaxPageSize {
		return limit
	}
	return config.MaxPageSize
}

func getMaxRecords(limit int) int {
	if limit > 0 {
		return limit
	}
	return config.MaxPageSize
}
