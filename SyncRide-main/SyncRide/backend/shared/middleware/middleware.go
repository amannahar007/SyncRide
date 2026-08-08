package middleware

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/syncride/backend/shared/constants"
	"github.com/syncride/backend/shared/errors"
	"github.com/syncride/backend/shared/jwt"
	"github.com/syncride/backend/shared/logger"
	"github.com/syncride/backend/shared/utils"
)

// responseWriterInterceptor captures HTTP status code and body byte size for logging/metrics
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
	bytesCount int64
}

func (rwi *responseWriterInterceptor) WriteHeader(code int) {
	rwi.statusCode = code
	rwi.ResponseWriter.WriteHeader(code)
}

func (rwi *responseWriterInterceptor) Write(b []byte) (int, error) {
	if rwi.statusCode == 0 {
		rwi.statusCode = http.StatusOK
	}
	n, err := rwi.ResponseWriter.Write(b)
	rwi.bytesCount += int64(n)
	return n, err
}

// RequestID Middleware
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(constants.HeaderRequestID)
		if reqID == "" {
			reqID = utils.GenerateUUID()
		}

		traceID := r.Header.Get(constants.HeaderTraceID)
		if traceID == "" {
			traceID = utils.GenerateTraceID()
		}

		w.Header().Set(constants.HeaderRequestID, reqID)
		w.Header().Set(constants.HeaderTraceID, traceID)

		ctx := context.WithValue(r.Context(), constants.CtxKeyRequestID, reqID)
		ctx = context.WithValue(ctx, constants.CtxKeyTraceID, traceID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ZapLogger Middleware
func ZapLogger(l *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rwi := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rwi, r)

			latency := time.Since(start)
			reqID, _ := r.Context().Value(constants.CtxKeyRequestID).(string)
			traceID, _ := r.Context().Value(constants.CtxKeyTraceID).(string)

			l.LogHTTPRequest(r.Method, r.URL.Path, reqID, traceID, rwi.statusCode, latency, rwi.bytesCount)
		})
	}
}

// Recovery Middleware
func Recovery(l *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rerr := recover(); rerr != nil {
					stackTrace := string(debug.Stack())
					reqID, _ := r.Context().Value(constants.CtxKeyRequestID).(string)

					l.Error("HTTP Server Panic Recovered",
						zap.Any("error", rerr),
						zap.String("request_id", reqID),
						zap.String("path", r.URL.Path),
						zap.String("stack", stackTrace),
					)

					appErr := errors.NewInternalError("An unexpected server panic occurred", fmt.Errorf("%v", rerr))
					errors.WriteHTTPError(w, r.URL.Path, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORS Middleware
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check allowed origin
			allow := false
			if len(allowedOrigins) == 0 || allowedOrigins[0] == "*" {
				allow = true
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				for _, o := range allowedOrigins {
					if o == origin {
						allow = true
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			if allow {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Trace-ID")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Gzip Middleware
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()

		gzw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(gzw, r)
	})
}

// Authentication Middleware
func JWTAuth(tm *jwt.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(constants.HeaderAuthorization)
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				errors.WriteHTTPError(w, r.URL.Path, errors.NewUnauthorizedError("Missing or invalid Authorization header"))
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := tm.ValidateToken(tokenStr)
			if err != nil {
				errors.WriteHTTPError(w, r.URL.Path, errors.NewUnauthorizedError(err.Error()))
				return
			}

			ctx := context.WithValue(r.Context(), constants.CtxKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, constants.CtxKeyUserRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
