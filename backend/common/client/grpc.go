package client

import (
	"MathOverflow/common/utils"
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	connMap   = make(map[string]*grpc.ClientConn)
	connMutex sync.Mutex

	breakerMap   = make(map[string]*utils.CircuitBreaker)
	breakerMutex sync.Mutex
)

func getBreaker(target string) *utils.CircuitBreaker {
	breakerMutex.Lock()
	defer breakerMutex.Unlock()

	if cb, ok := breakerMap[target]; ok {
		return cb
	}
	cb := utils.NewCircuitBreaker(5, 5*time.Second)
	breakerMap[target] = cb
	return cb
}

// unaryClientTraceInterceptor propagates trace ID from context to outgoing gRPC metadata,
// and adds basic timeout + retry + circuit-breaker behavior around the call.
func unaryClientTraceInterceptor(ctx context.Context, method string, req interface{},
	reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

	// Trace propagation
	traceID := utils.TraceIDFromContext(ctx)
	if traceID != "" {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		md.Set("x-request-id", traceID)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	target := cc.Target()
	cb := getBreaker(target)
	if !cb.Allow() {
		return status.Error(codes.Unavailable, "circuit breaker is open")
	}

	const (
		maxRetries     = 2
		perCallTimeout = 2 * time.Second
		initialBackoff = 100 * time.Millisecond
	)

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		callCtx := ctx
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(ctx, perCallTimeout)
			defer cancel()
		}

		err := invoker(callCtx, method, req, reply, cc, opts...)
		if err == nil {
			cb.OnSuccess()
			return nil
		}
		lastErr = err

		st, ok := status.FromError(err)
		if !ok {
			cb.OnFailure()
			return err
		}

		switch st.Code() {
		case codes.Unavailable, codes.DeadlineExceeded, codes.Canceled:
			cb.OnFailure()
			if attempt < maxRetries {
				time.Sleep(initialBackoff * time.Duration(attempt+1))
				continue
			}
			return err
		default:
			cb.OnFailure()
			return err
		}
	}

	return lastErr
}

func GetGRPCConn(target string) (*grpc.ClientConn, error) {
	connMutex.Lock()
	defer connMutex.Unlock()

	// 如果已存在，直接复用
	if c, ok := connMap[target]; ok {
		return c, nil
	}

	// 不存在则创建新的连接。
	// 使用 round_robin 负载均衡策略，配合 DNS 或自定义解析器，
	// 在同一逻辑地址后面挂多实例时，客户端可以自动在多个实例之间分摊流量。
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(unaryClientTraceInterceptor),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		return nil, err
	}

	connMap[target] = conn
	return conn, nil
}
