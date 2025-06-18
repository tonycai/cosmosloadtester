package loadtest

import (
	"fmt"
	"net/url"
	"time"
	
	"github.com/informalsystems/tm-load-test/pkg/loadtest"
)

// TransactorFactory creates the appropriate transactor based on endpoint protocol
type TransactorFactory struct{}

// NewTransactorFactory creates a new transactor factory
func NewTransactorFactory() *TransactorFactory {
	return &TransactorFactory{}
}

// CreateTransactor creates either a WebSocket or HTTP transactor based on the endpoint URL
func (tf *TransactorFactory) CreateTransactor(remoteAddr string, config *loadtest.Config) (TransactorInterface, error) {
	u, err := url.Parse(remoteAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	switch u.Scheme {
	case "ws", "wss":
		// Use original WebSocket transactor for WebSocket endpoints
		return loadtest.NewTransactor(remoteAddr, config)
	case "http", "https":
		// Use simple hybrid transactor for HTTP(S) endpoints
		return NewHybridTransactor(remoteAddr, config)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s (supported: ws://, wss://, http://, https://)", u.Scheme)
	}
}

// TransactorInterface defines the common interface for all transactor types
type TransactorInterface interface {
	SetProgressCallback(id int, interval time.Duration, callback func(int, int, int64))
	Start()
	Cancel()
	Wait() error
	GetTxCount() int
	GetTxBytes() int64
	GetTxRate() float64
}

// Ensure original Transactor implements the interface
var _ TransactorInterface = (*loadtest.Transactor)(nil)

// Ensure SimpleHybridTransactor implements the interface  
var _ TransactorInterface = (*SimpleHybridTransactor)(nil)