package grpc

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/syncride/backend/shared/constants"
	"github.com/syncride/backend/shared/logger"
	"github.com/syncride/backend/shared/utils"
)

// LoggingUnaryServerInterceptor logs gRPC calls with latency and status
func LoggingUnaryServerInterceptor(l *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		md, ok := metadata.FromIncomingContext(ctx)
		reqID := ""
		if ok {
			if vals := md.Get(constants.HeaderRequestID); len(vals) > 0 {
				reqID = vals[0]
			}
		}
		if reqID == "" {
			reqID = utils.GenerateUUID()
		}

		ctx = context.WithValue(ctx, constants.CtxKeyRequestID, reqID)

		resp, err := handler(ctx, req)

		latency := time.Since(start)
		st, _ := status.FromError(err)

		l.Info("gRPC Request",
			zap.String("method", info.FullMethod),
			zap.String("request_id", reqID),
			zap.String("code", st.Code().String()),
			zap.Duration("latency", latency),
		)

		return resp, err
	}
}

// RecoveryUnaryServerInterceptor catches gRPC panics
func RecoveryUnaryServerInterceptor(l *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				l.Error("gRPC Panic Recovered",
					zap.String("method", info.FullMethod),
					zap.Any("error", r),
				)
				err = status.Errorf(codes.Internal, "Internal server panic: %v", r)
			}
		}()
		return handler(ctx, req)
	}
}

// ClientMetadataUnaryInterceptor injects RequestID into outbound gRPC requests
func ClientMetadataUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		reqID, ok := ctx.Value(constants.CtxKeyRequestID).(string)
		if !ok || reqID == "" {
			reqID = utils.GenerateUUID()
		}

		ctx = metadata.AppendToOutgoingContext(ctx, constants.HeaderRequestID, reqID)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
