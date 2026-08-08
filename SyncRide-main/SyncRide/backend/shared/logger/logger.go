package logger

import (
	"context"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/syncride/backend/shared/constants"
)

var (
	globalLogger *Logger
	once         sync.Once
)

type Logger struct {
	zap *zap.Logger
}

// InitLogger initializes global Zap JSON logger for a service
func InitLogger(serviceName string, logLevel string) (*Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	zapLog := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.Fields(
		zap.String("service", serviceName),
	))

	l := &Logger{zap: zapLog}
	once.Do(func() {
		globalLogger = l
	})

	return l, nil
}

func GetLogger() *Logger {
	if globalLogger == nil {
		l, _ := InitLogger("syncride-service", "info")
		return l
	}
	return globalLogger
}

func (l *Logger) Sync() error {
	return l.zap.Sync()
}

func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{zap: l.zap.With(fields...)}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	var fields []zap.Field

	if reqID, ok := ctx.Value(constants.CtxKeyRequestID).(string); ok && reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}
	if traceID, ok := ctx.Value(constants.CtxKeyTraceID).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if userID, ok := ctx.Value(constants.CtxKeyUserID).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if len(fields) == 0 {
		return l
	}

	return &Logger{zap: l.zap.With(fields...)}
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// LogHTTPRequest logs standard HTTP access log
func (l *Logger) LogHTTPRequest(method, path, reqID, traceID string, status int, latency time.Duration, bytesWritten int64) {
	l.zap.Info("HTTP Request",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("request_id", reqID),
		zap.String("trace_id", traceID),
		zap.Int("status", status),
		zap.Duration("latency", latency),
		zap.Int64("bytes_written", bytesWritten),
	)
}
