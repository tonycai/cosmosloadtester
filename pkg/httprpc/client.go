package httprpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTPRPCClient provides HTTP RPC functionality as an alternative to WebSocket
type HTTPRPCClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	mutex      sync.RWMutex
	requestID  int64
}

// NewHTTPRPCClient creates a new HTTP RPC client
func NewHTTPRPCClient(endpoint string) (*HTTPRPCClient, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Convert HTTPS to HTTP RPC endpoint
	var baseURL string
	switch u.Scheme {
	case "https":
		baseURL = fmt.Sprintf("https://%s:%s", u.Hostname(), u.Port())
		if u.Port() == "" {
			baseURL = fmt.Sprintf("https://%s", u.Hostname())
		}
	case "http":
		baseURL = fmt.Sprintf("http://%s:%s", u.Hostname(), u.Port())
		if u.Port() == "" {
			baseURL = fmt.Sprintf("http://%s", u.Hostname())
		}
	default:
		return nil, fmt.Errorf("unsupported protocol: %s (http:// and https:// required for HTTP RPC)", u.Scheme)
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	logger := logrus.WithField("component", fmt.Sprintf("http-rpc[%s]", baseURL)).Logger

	return &HTTPRPCClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
		requestID:  1,
	}, nil
}

// BroadcastTx sends a transaction via HTTP RPC
func (c *HTTPRPCClient) BroadcastTx(method string, txBytes []byte) (*BroadcastTxResponse, error) {
	c.mutex.Lock()
	reqID := c.requestID
	c.requestID++
	c.mutex.Unlock()

	// Create JSON-RPC request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  method,
		Params: map[string]interface{}{
			"tx": txBytes,
		},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP POST request
	url := c.baseURL + "/"
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %s (status %d)", resp.Status, resp.StatusCode)
	}

	// Parse response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var rpcResponse JSONRPCResponse
	if err := json.Unmarshal(responseBody, &rpcResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResponse.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code %d)", rpcResponse.Error.Message, rpcResponse.Error.Code)
	}

	// Parse the result based on the broadcast method
	var result BroadcastTxResponse
	resultBytes, err := json.Marshal(rpcResponse.Result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal broadcast result: %w", err)
	}

	return &result, nil
}

// Close cleans up the HTTP client
func (c *HTTPRPCClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int64                  `json:"id"`
	Result  interface{}            `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// BroadcastTxResponse represents the response from a broadcast_tx call
type BroadcastTxResponse struct {
	Code      int    `json:"code"`
	Data      string `json:"data"`
	Log       string `json:"log"`
	Hash      string `json:"hash"`
	Codespace string `json:"codespace,omitempty"`
}

// HealthCheck verifies the HTTP RPC endpoint is accessible
func (c *HTTPRPCClient) HealthCheck() error {
	c.mutex.Lock()
	reqID := c.requestID
	c.requestID++
	c.mutex.Unlock()

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  "status",
		Params:  map[string]interface{}{},
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal health check request: %w", err)
	}

	url := c.baseURL + "/"
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("health check HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("health check failed: %s (status %d)", resp.Status, resp.StatusCode)
	}

	c.logger.Info("HTTP RPC health check passed")
	return nil
}