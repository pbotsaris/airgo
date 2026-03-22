package airtable

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

/* Client is an interface for the http clients used by this package */
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

/* AirtableClient is a wrapper for the http.Client */
type AirtableClient struct {
	Client *http.Client
}

func (c AirtableClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}

/* NewAirtableClient returns a new AirtableClient with default timeout */
func NewAirtableClient() *AirtableClient {
	return &AirtableClient{
		Client: &http.Client{
			Timeout: DefaultRequestTimeout,
		},
	}
}

/* NewAirtableClientWithTimeout returns a new AirtableClient with a custom timeout */
func NewAirtableClientWithTimeout(timeout time.Duration) *AirtableClient {
	return &AirtableClient{
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

/* mockClient is a mock http client for testing */
type mockClient struct {
	Response *http.Response
	Err      error
}

func (c *mockClient) Do(req *http.Request) (*http.Response, error) {
	return c.Response, c.Err
}

/* mockClientPaginate is a mock http client for testing paginationkj */
type mockClientPaginate struct {
	StatusCode int
	Err        error
	Data       [][]byte
	Index      int
}

func (c *mockClientPaginate) Do(req *http.Request) (*http.Response, error) {

	if c.Index >= len(c.Data) {
		panic(fmt.Sprintf("MockClientPaginate: Index '%d' out of range", c.Index))
	}

	next := c.Data[c.Index]
	reader := bytes.NewReader(next)

	mockResp := &http.Response{
		StatusCode: c.StatusCode,
		Body:       io.NopCloser(reader),
	}

	c.Index++

	return mockResp, c.Err
}

/* newMockClient returns a mock client that will return the provided body */
func newMockClient(s int, b []byte, err error) *mockClient {
	reader := bytes.NewReader(b)

	mockResp := &http.Response{
		StatusCode: s,
		Body:       io.NopCloser(reader),
	}

	return &mockClient{Response: mockResp, Err: err}
}

/* newMockClientPaginate returns a mock client that will return the provided bodies in order for each request */
func newMockClientPaginate(statusCode int, data [][]byte, err error) *mockClientPaginate {

	if len(data) < 2 {
		panic(fmt.Sprintf("MockClientPaginate requires at least 2 bodies. Only '%d' provided", len(data)))
	}

	return &mockClientPaginate{Err: err, Data: data, Index: 0, StatusCode: statusCode}
}
