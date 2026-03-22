package airtable

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/Antfood/airgo/retry"
)

// metaBasePath is appended to EndpointUrl to form the Meta API URL
const metaBasePath = "/meta/bases"

/*
Field represents an Airtable field definition from the Meta API.
Use Table.GetFields() or Table.GetField() to retrieve field metadata.
*/

type Field struct {
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Description string         `json:"description,omitempty"`
	Options     map[string]any `json:"options,omitempty"`
}

/*
View represents a saved view in an Airtable table.
*/
type View struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

/*
TableMeta represents table metadata from the Meta API.
*/
type TableMeta struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Fields      []Field `json:"fields"`
	Views       []View  `json:"views"`
}

type metaResponse struct {
	Tables []TableMeta       `json:"tables"`
	Error  map[string]string `json:"error,omitempty"`
}

func (m metaResponse) Err() map[string]string {
	return m.Error
}

// fieldCache stores cached field metadata per table
type fieldCache struct {
	mu     sync.RWMutex
	fields map[string][]Field // key: "baseId:tableId"
}

var cache = &fieldCache{
	fields: make(map[string][]Field),
}

func cacheKey(baseId, tableId string) string {
	return baseId + ":" + tableId
}

// get retrieves cached fields or nil if not cached
func (c *fieldCache) get(baseId, tableId string) []Field {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fields[cacheKey(baseId, tableId)]
}

// set stores fields in cache
func (c *fieldCache) set(baseId, tableId string, fields []Field) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fields[cacheKey(baseId, tableId)] = fields
}

// delete removes fields from cache
func (c *fieldCache) delete(baseId, tableId string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.fields, cacheKey(baseId, tableId))
}

/*
ClearFieldCache clears all cached field metadata.
Mostly used for testing
*/
func ClearFieldCache() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.fields = make(map[string][]Field)
}

func fetchTableFields(ctx context.Context, baseId, tableId string) ([]Field, error) {
	if client == nil {
		return nil, NewConfigError(OpGetFields, "client not configured; call SetToken or Configure first")
	}

	url := config.EndpointUrl + metaBasePath + "/" + baseId + "/tables"

	var resp metaResponse

	err := retry.DoCtx(ctx, func() error {
		httpReq, err := newHttpRequest(ctx, http.MethodGet, url, nil)
		if err != nil {
			return &Error{Op: OpGetFields, Message: "failed to create http request", Err: err}
		}

		res, err := client.Do(httpReq)
		if err != nil {
			return &Error{Op: OpGetFields, Message: "failed to make request", Err: err}
		}
		defer res.Body.Close()

		if res.StatusCode == 429 || res.StatusCode >= 500 {
			return &retry.HTTPError{StatusCode: res.StatusCode}
		}

		bodyData, err := io.ReadAll(res.Body)
		if err != nil {
			return &Error{Op: OpGetFields, Message: "failed to read response body", Err: err}
		}

		if err := json.Unmarshal(bodyData, &resp); err != nil {
			return &Error{Op: OpGetFields, Message: "failed to decode response", Err: err}
		}

		if len(resp.Error) > 0 {
			return ParseAPIError(OpGetFields, res.StatusCode, bodyData)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, table := range resp.Tables {
		if table.Id == tableId || table.Name == tableId {
			return table.Fields, nil
		}
	}

	return nil, &APIError{
		Op:      OpGetFields,
		Type:    ErrTypeNotFound,
		Message: "table '" + tableId + "' not found in base '" + baseId + "'",
	}
}
