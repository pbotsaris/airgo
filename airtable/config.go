package airtable

import "time"

// Constants that don't need to be configurable
const (
	bearer         = "Bearer "
	DateTimeLayout = "2006-01-02T15:04:05.000Z"
)

// Default configuration values
const (
	DefaultEndpointUrl   = "https://api.airtable.com/v0"
	DefaultMaxPageSize   = 100
	DefaultMaxUrlLength  = 15000
	DefaultRequestTimeout = 5 * time.Minute
)

/*
Config holds the configuration for the Airtable client.
*/
type Config struct {
	Token                string
	EndpointUrl          string
	MaxPageSize          int
	MaxUrlLength         int
	RequestTimeout       time.Duration
	NoRetryIfRateLimited bool
	CustomHeaders        map[string]string
}

var config = Config{
	EndpointUrl:    DefaultEndpointUrl,
	MaxPageSize:    DefaultMaxPageSize,
	MaxUrlLength:   DefaultMaxUrlLength,
	RequestTimeout: DefaultRequestTimeout,
}

var client Client = NewAirtableClient()

/*
SetToken sets the Airtable API token used for authentication.

Example:

	airtable.SetToken(os.Getenv("AIRTABLE_TOKEN"))
*/
func SetToken(airtableToken string) {
	config.Token = airtableToken
}

/*
Configure sets both the HTTP client and the Airtable API token.
Use this when you need a custom HTTP client (e.g., for testing or custom timeouts).

Example:

	client := airtable.NewAirtableClient()
	airtable.Configure(client, os.Getenv("AIRTABLE_TOKEN"))
*/
func Configure(c Client, airtableToken string) {
	config.Token = airtableToken
	client = c
}

/*
ConfigureWithOptions sets the full configuration for the Airtable client.
Any zero values in the provided config will use defaults.

Example:

	airtable.ConfigureWithOptions(airtable.Config{
	    Token:          os.Getenv("AIRTABLE_TOKEN"),
	    EndpointUrl:    "https://api.airtable.com/v0",
	    MaxPageSize:    50,
	    RequestTimeout: 2 * time.Minute,
	    CustomHeaders:  map[string]string{"X-Custom-Header": "value"},
	})
*/
func ConfigureWithOptions(cfg Config) {
	if cfg.Token != "" {
		config.Token = cfg.Token
	}
	if cfg.EndpointUrl != "" {
		config.EndpointUrl = cfg.EndpointUrl
	}
	if cfg.MaxPageSize > 0 {
		config.MaxPageSize = cfg.MaxPageSize
	}
	if cfg.MaxUrlLength > 0 {
		config.MaxUrlLength = cfg.MaxUrlLength
	}
	if cfg.RequestTimeout > 0 {
		config.RequestTimeout = cfg.RequestTimeout
	}
	config.NoRetryIfRateLimited = cfg.NoRetryIfRateLimited
	if cfg.CustomHeaders != nil {
		config.CustomHeaders = cfg.CustomHeaders
	}

	// Create a new client with the configured timeout
	client = NewAirtableClientWithTimeout(config.RequestTimeout)
}

/*
GetConfig returns a copy of the current configuration.
*/
func GetConfig() Config {
	return config
}
