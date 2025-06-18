package recovery

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/orijtech/cosmosloadtester/pkg/errors"
	"github.com/orijtech/cosmosloadtester/pkg/logger"
)

// RecoveryHandler handles panic recovery
type RecoveryHandler struct {
	logger logger.Logger
}

// NewRecoveryHandler creates a new recovery handler
func NewRecoveryHandler(log logger.Logger) *RecoveryHandler {
	return &RecoveryHandler{
		logger: log,
	}
}

// RecoverWithError recovers from a panic and returns an error
func (r *RecoveryHandler) RecoverWithError() error {
	if rec := recover(); rec != nil {
		err := r.handlePanic(rec)
		r.logger.WithError(err).Error("Panic recovered")
		return err
	}
	return nil
}

// RecoverWithCallback recovers from a panic and calls a callback
func (r *RecoveryHandler) RecoverWithCallback(callback func(error)) {
	if rec := recover(); rec != nil {
		err := r.handlePanic(rec)
		r.logger.WithError(err).Error("Panic recovered")
		if callback != nil {
			callback(err)
		}
	}
}

// SafeGo runs a function in a goroutine with panic recovery
func (r *RecoveryHandler) SafeGo(fn func()) {
	go func() {
		defer r.RecoverWithCallback(nil)
		fn()
	}()
}

// SafeGoWithContext runs a function in a goroutine with context and panic recovery
func (r *RecoveryHandler) SafeGoWithContext(ctx context.Context, fn func(context.Context)) {
	go func() {
		defer r.RecoverWithCallback(nil)
		fn(ctx)
	}()
}

// SafeExecute executes a function with panic recovery
func (r *RecoveryHandler) SafeExecute(fn func() error) error {
	defer func() {
		if rec := recover(); rec != nil {
			err := r.handlePanic(rec)
			r.logger.WithError(err).Error("Panic during execution")
		}
	}()
	
	return fn()
}

// SafeExecuteWithRetry executes a function with panic recovery and retry logic
func (r *RecoveryHandler) SafeExecuteWithRetry(fn func() error, maxRetries int, delay time.Duration) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := r.SafeExecute(fn)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if error is recoverable
		if !errors.IsRecoverable(err) {
			r.logger.WithError(err).Warn("Non-recoverable error, not retrying")
			return err
		}
		
		if attempt < maxRetries {
			r.logger.WithFields(logger.Fields{
				"attempt": attempt + 1,
				"max_retries": maxRetries,
				"delay": delay.String(),
			}).WithError(err).Warn("Retrying after error")
			
			time.Sleep(delay)
		}
	}
	
	return lastErr
}

// handlePanic converts a panic to a structured error
func (r *RecoveryHandler) handlePanic(rec interface{}) error {
	stack := debug.Stack()
	
	var message string
	switch v := rec.(type) {
	case error:
		message = v.Error()
	case string:
		message = v
	default:
		message = fmt.Sprintf("panic: %v", v)
	}
	
	return errors.NewInternalError(errors.ErrCodeInternalError, message).
		WithDetails(string(stack)).
		WithContext("panic_value", rec).
		WithContext("stack_trace", string(stack))
}

// Global recovery handler
var globalRecoveryHandler *RecoveryHandler

// SetGlobalRecoveryHandler sets the global recovery handler
func SetGlobalRecoveryHandler(handler *RecoveryHandler) {
	globalRecoveryHandler = handler
}

// GetGlobalRecoveryHandler returns the global recovery handler
func GetGlobalRecoveryHandler() *RecoveryHandler {
	if globalRecoveryHandler == nil {
		globalRecoveryHandler = NewRecoveryHandler(logger.GetGlobalLogger())
	}
	return globalRecoveryHandler
}

// Convenience functions using global recovery handler

// Recover recovers from a panic and returns an error
func Recover() error {
	return GetGlobalRecoveryHandler().RecoverWithError()
}

// RecoverWithCallback recovers from a panic and calls a callback
func RecoverWithCallback(callback func(error)) {
	GetGlobalRecoveryHandler().RecoverWithCallback(callback)
}

// SafeGo runs a function in a goroutine with panic recovery
func SafeGo(fn func()) {
	GetGlobalRecoveryHandler().SafeGo(fn)
}

// SafeGoWithContext runs a function in a goroutine with context and panic recovery
func SafeGoWithContext(ctx context.Context, fn func(context.Context)) {
	GetGlobalRecoveryHandler().SafeGoWithContext(ctx, fn)
}

// SafeExecute executes a function with panic recovery
func SafeExecute(fn func() error) error {
	return GetGlobalRecoveryHandler().SafeExecute(fn)
}

// SafeExecuteWithRetry executes a function with panic recovery and retry logic
func SafeExecuteWithRetry(fn func() error, maxRetries int, delay time.Duration) error {
	return GetGlobalRecoveryHandler().SafeExecuteWithRetry(fn, maxRetries, delay)
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay" yaml:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay" yaml:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor" yaml:"backoff_factor"`
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  1 * time.Second,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
	}
}

// ExponentialBackoffRetry executes a function with exponential backoff retry
func (r *RecoveryHandler) ExponentialBackoffRetry(fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}
	
	var lastErr error
	delay := config.InitialDelay
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		err := r.SafeExecute(fn)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Check if error is recoverable
		if !errors.IsRecoverable(err) {
			r.logger.WithError(err).Warn("Non-recoverable error, not retrying")
			return err
		}
		
		if attempt < config.MaxRetries {
			r.logger.WithFields(logger.Fields{
				"attempt": attempt + 1,
				"max_retries": config.MaxRetries,
				"delay": delay.String(),
			}).WithError(err).Warn("Retrying with exponential backoff")
			
			time.Sleep(delay)
			
			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}
	
	return lastErr
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failureCount    int
	lastFailureTime time.Time
	state           CircuitBreakerState
	logger          logger.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, log logger.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitBreakerClosed,
		logger:       log,
	}
}

// Execute executes a function through the circuit breaker
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if cb.state == CircuitBreakerOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.logger.Info("Circuit breaker transitioning to half-open state")
		} else {
			return errors.NewConnectionError("CIRCUIT_BREAKER_OPEN", "Circuit breaker is open")
		}
	}
	
	err := fn()
	
	if err != nil {
		cb.onFailure()
		return err
	}
	
	cb.onSuccess()
	return nil
}

// onFailure handles a failure
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	
	if cb.failureCount >= cb.maxFailures {
		cb.state = CircuitBreakerOpen
		cb.logger.WithFields(logger.Fields{
			"failure_count": cb.failureCount,
			"max_failures": cb.maxFailures,
		}).Warn("Circuit breaker opened due to failures")
	}
}

// onSuccess handles a success
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0
	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerClosed
		cb.logger.Info("Circuit breaker closed after successful execution")
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.state = CircuitBreakerClosed
	cb.failureCount = 0
	cb.logger.Info("Circuit breaker manually reset")
}

// HealthChecker provides health checking functionality
type HealthChecker struct {
	checks map[string]func() error
	logger logger.Logger
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(log logger.Logger) *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]func() error),
		logger: log,
	}
}

// AddCheck adds a health check
func (hc *HealthChecker) AddCheck(name string, check func() error) {
	hc.checks[name] = check
}

// CheckHealth performs all health checks
func (hc *HealthChecker) CheckHealth() map[string]error {
	results := make(map[string]error)
	
	for name, check := range hc.checks {
		err := SafeExecute(check)
		results[name] = err
		
		if err != nil {
			hc.logger.WithFields(logger.Fields{
				"check": name,
			}).WithError(err).Warn("Health check failed")
		} else {
			hc.logger.WithFields(logger.Fields{
				"check": name,
			}).Debug("Health check passed")
		}
	}
	
	return results
}

// IsHealthy returns true if all health checks pass
func (hc *HealthChecker) IsHealthy() bool {
	results := hc.CheckHealth()
	for _, err := range results {
		if err != nil {
			return false
		}
	}
	return true
}

// ErrorCollector collects and aggregates errors
type ErrorCollector struct {
	errors []error
	logger logger.Logger
}

// NewErrorCollector creates a new error collector
func NewErrorCollector(log logger.Logger) *ErrorCollector {
	return &ErrorCollector{
		errors: make([]error, 0),
		logger: log,
	}
}

// Add adds an error to the collector
func (ec *ErrorCollector) Add(err error) {
	if err != nil {
		ec.errors = append(ec.errors, err)
		ec.logger.WithError(err).Debug("Error added to collector")
	}
}

// HasErrors returns true if there are any errors
func (ec *ErrorCollector) HasErrors() bool {
	return len(ec.errors) > 0
}

// GetErrors returns all collected errors
func (ec *ErrorCollector) GetErrors() []error {
	return ec.errors
}

// GetFirstError returns the first error or nil
func (ec *ErrorCollector) GetFirstError() error {
	if len(ec.errors) > 0 {
		return ec.errors[0]
	}
	return nil
}

// Clear clears all collected errors
func (ec *ErrorCollector) Clear() {
	ec.errors = ec.errors[:0]
}

// ToMultiError converts collected errors to a single multi-error
func (ec *ErrorCollector) ToMultiError() error {
	if len(ec.errors) == 0 {
		return nil
	}
	
	if len(ec.errors) == 1 {
		return ec.errors[0]
	}
	
	return &MultiError{errors: ec.errors}
}

// MultiError represents multiple errors
type MultiError struct {
	errors []error
}

// Error implements the error interface
func (me *MultiError) Error() string {
	if len(me.errors) == 0 {
		return "no errors"
	}
	
	if len(me.errors) == 1 {
		return me.errors[0].Error()
	}
	
	return fmt.Sprintf("multiple errors occurred: %d errors", len(me.errors))
}

// Errors returns all errors
func (me *MultiError) Errors() []error {
	return me.errors
}

// Unwrap returns the first error for error unwrapping
func (me *MultiError) Unwrap() error {
	if len(me.errors) > 0 {
		return me.errors[0]
	}
	return nil
} 