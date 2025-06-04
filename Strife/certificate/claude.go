package interceptors

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// Logger instance for payment system logs
var logger *log.Logger

func init() {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
	}

	// Set up log file
	logFile, err := os.OpenFile(
		filepath.Join("logs", "payment_system.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		logger = log.New(os.Stdout, "[PAYMENT-GATEWAY] ", log.LstdFlags)
	} else {
		// Configure logger to write to both file and stdout
		multiWriter := log.MultiWriter(logFile, os.Stdout)
		logger = log.New(multiWriter, "[PAYMENT-GATEWAY] ", log.LstdFlags)
	}
}

// extractRequestInfo attempts to extract relevant transaction information from the request
func extractRequestInfo(req interface{}) map[string]interface{} {
	info := make(map[string]interface{})

	// Use reflection to examine request fields
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return info
	}

	// Look for common payment-related fields
	paymentFields := []string{"Amount", "SenderID", "RecipientID", "AccountID", "TransactionID", "Balance"}
	for _, field := range paymentFields {
		f := v.FieldByName(field)
		if f.IsValid() && f.CanInterface() {
			info[field] = f.Interface()
		}
	}

	return info
}

// getClientInfo extracts client information from the context
func getClientInfo(ctx context.Context) map[string]string {
	clientInfo := make(map[string]string)

	// Extract peer information (client IP address)
	if p, ok := peer.FromContext(ctx); ok {
		clientInfo["address"] = p.Addr.String()
	}

	// Extract metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgent := md.Get("user-agent"); len(userAgent) > 0 {
			clientInfo["user-agent"] = userAgent[0]
		}
		if authorization := md.Get("authorization"); len(authorization) > 0 {
			// Don't log the actual token, just note if it's present
			clientInfo["authenticated"] = "true"
		}
	}

	return clientInfo
}

// UnaryServerInterceptor returns a gRPC interceptor that logs request/response information
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Generate transaction ID for tracing
		transactionID := uuid.New().String()

		// Add transaction ID to context metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "transaction-id", transactionID)

		// Get method name from info
		methodName := info.FullMethod

		// Log request start
		startTime := time.Now()
		logger.Printf("[%s] Request started: %s", transactionID, methodName)

		// Log client information
		clientInfo := getClientInfo(ctx)
		for k, v := range clientInfo {
			logger.Printf("[%s] Client %s: %s", transactionID, k, v)
		}

		// For payment/transaction methods, log more details
		if strings.Contains(methodName, "Payment") || strings.Contains(methodName, "Transaction") {
			requestInfo := extractRequestInfo(req)
			for k, v := range requestInfo {
				logger.Printf("[%s] %s: %v", transactionID, k, v)
			}
		}

		// Execute the handler
		resp, err := handler(ctx, req)

		// Calculate execution time
		duration := time.Since(startTime)

		// Log response information
		if err != nil {
			st, _ := status.FromError(err)
			logger.Printf("[%s] Request failed: %s - code: %s, message: %s (took %v)",
				transactionID, methodName, st.Code(), st.Message(), duration)
		} else {
			logger.Printf("[%s] Request successful: %s (took %v)",
				transactionID, methodName, duration)

			// Log response details if possible
			responseInfo := extractRequestInfo(resp)
			for k, v := range responseInfo {
				logger.Printf("[%s] Response %s: %v", transactionID, k, v)
			}
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a gRPC interceptor for streaming RPCs
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Generate transaction ID for tracing
		transactionID := uuid.New().String()

		// Log stream start
		startTime := time.Now()
		logger.Printf("[%s] Stream started: %s", transactionID, info.FullMethod)

		// Get client information from context
		clientInfo := getClientInfo(ss.Context())
		for k, v := range clientInfo {
			logger.Printf("[%s] Client %s: %s", transactionID, k, v)
		}

		// Wrap the server stream to intercept messages
		wrappedStream := &wrappedServerStream{
			ServerStream:  ss,
			transactionID: transactionID,
		}

		// Execute the handler
		err := handler(srv, wrappedStream)

		// Calculate execution time
		duration := time.Since(startTime)

		// Log stream end
		if err != nil {
			st, _ := status.FromError(err)
			logger.Printf("[%s] Stream failed: %s - code: %s, message: %s (took %v)",
				transactionID, info.FullMethod, st.Code(), st.Message(), duration)
		} else {
			logger.Printf("[%s] Stream completed: %s (took %v)",
				transactionID, info.FullMethod, duration)
		}

		return err
	}
}

// wrappedServerStream wraps grpc.ServerStream to intercept the RecvMsg and SendMsg methods
type wrappedServerStream struct {
	grpc.ServerStream
	transactionID string
	msgCount      int
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)

	if err == nil {
		w.msgCount++
		// Log received message
		logger.Printf("[%s] Stream received message #%d: %s",
			w.transactionID, w.msgCount, reflect.TypeOf(m).String())

		// For payment/transaction messages, log more details
		methodName := w.ServerStream.Context().Value("grpc_method")
		if methodName != nil {
			methodStr := methodName.(string)
			if strings.Contains(methodStr, "Payment") || strings.Contains(methodStr, "Transaction") {
				requestInfo := extractRequestInfo(m)
				for k, v := range requestInfo {
					logger.Printf("[%s] Message %s: %v", w.transactionID, k, v)
				}
			}
		}
	}

	return err
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
	// Log sent message
	logger.Printf("[%s] Stream sending message: %s",
		w.transactionID, reflect.TypeOf(m).String())

	return w.ServerStream.SendMsg(m)
}

// UnaryClientInterceptor returns a gRPC client interceptor for unary RPCs
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Generate transaction ID for tracing
		transactionID := uuid.New().String()

		// Add transaction ID to outgoing context
		ctx = metadata.AppendToOutgoingContext(ctx, "transaction-id", transactionID)

		// Log request start
		startTime := time.Now()
		logger.Printf("[CLIENT][%s] Request started: %s", transactionID, method)

		// For payment/transaction methods, log more details
		if strings.Contains(method, "Payment") || strings.Contains(method, "Transaction") {
			requestInfo := extractRequestInfo(req)
			for k, v := range requestInfo {
				logger.Printf("[CLIENT][%s] %s: %v", transactionID, k, v)
			}
		}

		// Execute the invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Calculate execution time
		duration := time.Since(startTime)

		// Log response information
		if err != nil {
			st, _ := status.FromError(err)
			logger.Printf("[CLIENT][%s] Request failed: %s - code: %s, message: %s (took %v)",
				transactionID, method, st.Code(), st.Message(), duration)
		} else {
			logger.Printf("[CLIENT][%s] Request successful: %s (took %v)",
				transactionID, method, duration)

			// Log response details if possible
			responseInfo := extractRequestInfo(reply)
			for k, v := range responseInfo {
				logger.Printf("[CLIENT][%s] Response %s: %v", transactionID, k, v)
			}
		}

		return err
	}
}

// StreamClientInterceptor returns a gRPC client interceptor for streaming RPCs
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// Generate transaction ID for tracing
		transactionID := uuid.New().String()

		// Add transaction ID to outgoing context
		ctx = metadata.AppendToOutgoingContext(ctx, "transaction-id", transactionID)

		// Log stream start
		startTime := time.Now()
		logger.Printf("[CLIENT][%s] Stream started: %s", transactionID, method)

		// Execute the streamer
		stream, err := streamer(ctx, desc, cc, method, opts...)

		if err != nil {
			st, _ := status.FromError(err)
			logger.Printf("[CLIENT][%s] Stream creation failed: %s - code: %s, message: %s",
				transactionID, method, st.Code(), st.Message())
			return nil, err
		}

		// Return wrapped client stream
		return &wrappedClientStream{
			ClientStream:  stream,
			transactionID: transactionID,
			method:        method,
			startTime:     startTime,
		}, nil
	}
}

// wrappedClientStream wraps grpc.ClientStream to intercept the RecvMsg and SendMsg methods
type wrappedClientStream struct {
	grpc.ClientStream
	transactionID string
	method        string
	startTime     time.Time
	msgCount      int
}

func (w *wrappedClientStream) RecvMsg(m interface{}) error {
	err := w.ClientStream.RecvMsg(m)

	if err == nil {
		w.msgCount++
		// Log received message
		logger.Printf("[CLIENT][%s] Stream received message #%d: %s",
			w.transactionID, w.msgCount, reflect.TypeOf(m).String())

		// For payment/transaction messages, log more details
		if strings.Contains(w.method, "Payment") || strings.Contains(w.method, "Transaction") {
			responseInfo := extractRequestInfo(m)
			for k, v := range responseInfo {
				logger.Printf("[CLIENT][%s] Response %s: %v", w.transactionID, k, v)
			}
		}
	} else if err != nil && err.Error() != "EOF" {
		st, _ := status.FromError(err)
		logger.Printf("[CLIENT][%s] Stream receive error: code: %s, message: %s",
			w.transactionID, st.Code(), st.Message())
	}

	return err
}

func (w *wrappedClientStream) SendMsg(m interface{}) error {
	// Log sent message
	logger.Printf("[CLIENT][%s] Stream sending message: %s",
		w.transactionID, reflect.TypeOf(m).String())

	// For payment/transaction messages, log more details
	if strings.Contains(w.method, "Payment") || strings.Contains(w.method, "Transaction") {
		requestInfo := extractRequestInfo(m)
		for k, v := range requestInfo {
			logger.Printf("[CLIENT][%s] Message %s: %v", w.transactionID, k, v)
		}
	}

	return w.ClientStream.SendMsg(m)
}

func (w *wrappedClientStream) CloseSend() error {
	logger.Printf("[CLIENT][%s] Stream closed: %s (took %v)",
		w.transactionID, w.method, time.Since(w.startTime))
	return w.ClientStream.CloseSend()
}

// Helper function to check if an error should trigger a retry
func ShouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// Get status code
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	// These status codes typically indicate transient errors
	retryableCodes := []codes.Code{
		codes.Unavailable,       // Server is currently unavailable
		codes.DeadlineExceeded,  // Request deadline exceeded
		codes.Aborted,           // Concurrency conflict
		codes.Internal,          // Server internal error
		codes.ResourceExhausted, // Rate limiting
	}

	for _, code := range retryableCodes {
		if st.Code() == code {
			return true
		}
	}

	return false
}

// RetryInterceptor implements retry logic for failed requests
func RetryUnaryInterceptor(maxRetries int, backoffDuration time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Generate transaction ID for tracing
		transactionID := uuid.New().String()

		var lastErr error

		// Try the call up to maxRetries times
		for attempt := 0; attempt <= maxRetries; attempt++ {
			// If this is a retry, log it
			if attempt > 0 {
				logger.Printf("[RETRY][%s] Attempt %d/%d for method %s",
					transactionID, attempt, maxRetries, method)

				// Wait before retrying with exponential backoff
				retryDelay := backoffDuration * time.Duration(1<<uint(attempt-1))
				logger.Printf("[RETRY][%s] Waiting %v before retry",
					transactionID, retryDelay)

				select {
				case <-time.After(retryDelay):
					// Continue with retry
				case <-ctx.Done():
					// Context canceled or timed out
					return fmt.Errorf("retry canceled: %w", ctx.Err())
				}
			}

			// Make the actual call
			err := invoker(ctx, method, req, reply, cc, opts...)

			// If call succeeded or shouldn't be retried, return result
			if err == nil || !ShouldRetry(err) {
				return err
			}

			// Store the error for potential later return
			lastErr = err

			st, _ := status.FromError(err)
			logger.Printf("[RETRY][%s] Request failed with retryable error: %s - code: %s",
				transactionID, method, st.Code())
		}

		// All retries failed
		logger.Printf("[RETRY][%s] All %d retry attempts failed for method %s",
			transactionID, maxRetries, method)

		return lastErr
	}
}
