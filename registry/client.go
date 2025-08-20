// Package registry provides a client for interacting with the Terraform Registry API.
package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultBaseURL is the default base URL for the Terraform Registry API
	DefaultBaseURL = "https://registry.terraform.io"

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default maximum number of retries
	DefaultMaxRetries = 3

	// DefaultUserAgent is the default user agent string
	DefaultUserAgent = "terraform-registry-client/1.0"
)

var (
	// ErrClientNotInitialized is returned when the client is not properly initialized
	ErrClientNotInitialized = errors.New("client not initialized")

	// ErrInvalidConfiguration is returned when the client configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid client configuration")
)

// Client represents a Terraform Registry API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	userAgent  string
	apiToken   string // For future private registry support

	// Rate limiting
	rateLimiter *RateLimiter

	// Service clients
	Providers ProvidersServiceInterface
	Modules   ModulesServiceInterface
	Policies  PoliciesServiceInterface

	// Configuration
	config *ClientConfig

	// Ensure thread safety
	mu sync.RWMutex
}

// ClientConfig holds the configuration for the client
type ClientConfig struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	UserAgent  string
	APIToken   string
	Logger     *logrus.Logger

	// Rate limiting configuration
	RateLimitRequests int
	RateLimitPeriod   time.Duration

	// HTTP client configuration
	HTTPClient *http.Client

	// Retry configuration
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration

	// Circuit breaker configuration
	CircuitBreakerThreshold   int
	CircuitBreakerTimeout     time.Duration
	CircuitBreakerMaxRequests int
}

// DefaultClientConfig returns a default client configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		BaseURL:                   DefaultBaseURL,
		Timeout:                   DefaultTimeout,
		MaxRetries:                DefaultMaxRetries,
		UserAgent:                 DefaultUserAgent,
		RateLimitRequests:         100,
		RateLimitPeriod:           time.Minute,
		RetryWaitMin:              1 * time.Second,
		RetryWaitMax:              30 * time.Second,
		CircuitBreakerThreshold:   5,
		CircuitBreakerTimeout:     60 * time.Second,
		CircuitBreakerMaxRequests: 1,
		Logger:                    logrus.New(),
	}
}

// ClientOption is a function that configures a Client
type ClientOption func(*ClientConfig)

// WithBaseURL sets a custom base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *ClientConfig) {
		c.BaseURL = baseURL
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *ClientConfig) {
		c.HTTPClient = httpClient
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *logrus.Logger) ClientOption {
	return func(c *ClientConfig) {
		c.Logger = logger
	}
}

// WithTimeout sets a custom timeout for the HTTP client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = timeout
	}
}

// WithUserAgent sets a custom user agent
func WithUserAgent(userAgent string) ClientOption {
	return func(c *ClientConfig) {
		c.UserAgent = userAgent
	}
}

// WithAPIToken sets an API token for authentication
func WithAPIToken(token string) ClientOption {
	return func(c *ClientConfig) {
		c.APIToken = token
	}
}

// WithRateLimit configures rate limiting
func WithRateLimit(requests int, period time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.RateLimitRequests = requests
		c.RateLimitPeriod = period
	}
}

// NewClient creates a new Terraform Registry API client
func NewClient(opts ...ClientOption) (*Client, error) {
	config := DefaultClientConfig()

	// Apply options
	for _, opt := range opts {
		opt(config)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfiguration, err)
	}

	client := &Client{
		baseURL:   config.BaseURL,
		logger:    config.Logger,
		userAgent: config.UserAgent,
		apiToken:  config.APIToken,
		config:    config,
	}

	// Create HTTP client if not provided
	if config.HTTPClient == nil {
		httpClient, err := newDefaultHTTPClient(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP client: %w", err)
		}
		client.httpClient = httpClient
	} else {
		client.httpClient = config.HTTPClient
	}

	// Initialize rate limiter
	client.rateLimiter = NewRateLimiter(config.RateLimitRequests, config.RateLimitPeriod)

	// Initialize service clients
	client.Providers = &ProvidersService{client: client}
	client.Modules = &ModulesService{client: client}
	client.Policies = &PoliciesService{client: client}

	return client, nil
}

// validateConfig validates the client configuration
func validateConfig(config *ClientConfig) error {
	if config.BaseURL == "" {
		return errors.New("base URL cannot be empty")
	}

	if _, err := url.Parse(config.BaseURL); err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	if config.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	if config.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}

	if config.RateLimitRequests <= 0 {
		return errors.New("rate limit requests must be positive")
	}

	if config.RateLimitPeriod <= 0 {
		return errors.New("rate limit period must be positive")
	}

	return nil
}

// newDefaultHTTPClient creates a default HTTP client with retry logic
func newDefaultHTTPClient(config *ClientConfig) (*http.Client, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = config.Logger

	transport := cleanhttp.DefaultPooledTransport()
	transport.Proxy = http.ProxyFromEnvironment
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 10

	retryClient.HTTPClient = &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}
	retryClient.RetryMax = config.MaxRetries
	retryClient.RetryWaitMin = config.RetryWaitMin
	retryClient.RetryWaitMax = config.RetryWaitMax

	// Custom backoff for rate limiting
	retryClient.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			if resetAfter := resp.Header.Get("x-ratelimit-reset"); resetAfter != "" {
				var resetTime int64
				if _, err := fmt.Sscanf(resetAfter, "%d", &resetTime); err == nil {
					waitTime := time.Until(time.Unix(resetTime, 0))
					config.Logger.Debugf("Rate limited, waiting %v until reset", waitTime)
					return waitTime
				}
			}
		}
		return retryablehttp.DefaultBackoff(min, max, attemptNum, resp)
	}

	// Custom retry policy
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			// Always retry on network errors
			return true, nil
		}

		if resp != nil {
			if resp.StatusCode == http.StatusTooManyRequests {
				return true, nil
			}

			// Retry on 5xx errors
			if resp.StatusCode >= 500 {
				return true, nil
			}
		}

		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	return retryClient.StandardClient(), nil
}

// get performs a GET request to the specified path
func (c *Client) get(ctx context.Context, path string, version string, result interface{}) error {
	return c.request(ctx, "GET", path, version, nil, result)
}

// request performs an HTTP request
func (c *Client) request(ctx context.Context, method, path, version string, body io.Reader, result interface{}) error {
	// Check rate limit
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit error: %w", err)
	}

	req, err := c.newRequest(ctx, method, path, version, body)
	if err != nil {
		return err
	}

	return c.do(req, result)
}

// newRequest creates a new HTTP request
func (c *Client) newRequest(ctx context.Context, method, path, version string, body io.Reader) (*http.Request, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	u, err := url.Parse(fmt.Sprintf("%s/%s/%s", c.baseURL, version, path))
	if err != nil {
		return nil, &RequestError{
			Method: method,
			URL:    fmt.Sprintf("%s/%s/%s", c.baseURL, version, path),
			Err:    fmt.Errorf("error parsing URL: %w", err),
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, &RequestError{
			Method: method,
			URL:    u.String(),
			Err:    fmt.Errorf("error creating request: %w", err),
		}
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add authentication if available
	if c.apiToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiToken))
	}

	return req, nil
}

// do performs the HTTP request and decodes the response
func (c *Client) do(req *http.Request, result interface{}) error {
	c.logger.WithFields(logrus.Fields{
		"method": req.Method,
		"url":    req.URL.String(),
	}).Debug("Sending request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &RequestError{
			Method: req.Method,
			URL:    req.URL.String(),
			Err:    fmt.Errorf("error performing request: %w", err),
		}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ResponseError{
			StatusCode: resp.StatusCode,
			Err:        fmt.Errorf("error reading response body: %w", err),
		}
	}

	c.logger.WithFields(logrus.Fields{
		"status": resp.StatusCode,
		"length": len(body),
	}).Debug("Received response")

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
			Headers:    resp.Header,
		}

		// Try to parse error response
		var errResp struct {
			Message string `json:"message"`
			Errors  []struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"errors"`
		}

		if err := json.Unmarshal(body, &errResp); err == nil {
			if errResp.Message != "" {
				apiErr.Message = errResp.Message
			}
			if len(errResp.Errors) > 0 {
				apiErr.Message = errResp.Errors[0].Message
			}
		}

		return apiErr
	}

	// Decode response if result is provided
	if result != nil && len(body) > 0 {
		if err := json.Unmarshal(body, result); err != nil {
			return &ResponseError{
				StatusCode: resp.StatusCode,
				Err:        fmt.Errorf("error decoding response: %w", err),
			}
		}
	}

	return nil
}

// SetBaseURL updates the base URL for the client
func (c *Client) SetBaseURL(baseURL string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, err := url.Parse(baseURL); err != nil {
		return fmt.Errorf("invalid base URL: %w", err)
	}

	c.baseURL = baseURL
	return nil
}

// GetBaseURL returns the current base URL
func (c *Client) GetBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.baseURL
}

// GetRateLimiter returns the client's rate limiter
func (c *Client) GetRateLimiter() *RateLimiter {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rateLimiter
}
