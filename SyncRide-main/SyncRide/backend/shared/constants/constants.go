package constants

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
	EnvTest        Environment = "test"
)

// HTTP Headers
const (
	HeaderRequestID     = "X-Request-ID"
	HeaderTraceID       = "X-Trace-ID"
	HeaderUserID        = "X-User-ID"
	HeaderUserRole      = "X-User-Role"
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
)

// Context Keys
type ContextKey string

const (
	CtxKeyRequestID ContextKey = "request_id"
	CtxKeyTraceID   ContextKey = "trace_id"
	CtxKeyUserID    ContextKey = "user_id"
	CtxKeyUserRole  ContextKey = "user_role"
	CtxKeyLogger    ContextKey = "logger"
)

// Service Names
const (
	ServiceAPIGateway        = "api-gateway"
	ServiceAuth              = "auth-service"
	ServicePassenger         = "passenger-service"
	ServiceDriver            = "driver-service"
	ServiceRide              = "ride-service"
	ServiceTrip              = "trip-service"
	ServicePricing           = "pricing-service"
	ServicePayment           = "payment-service"
	ServiceNotification      = "notification-service"
	ServiceMaps              = "maps-service"
	ServiceAnalytics         = "analytics-service"
)

// Error Codes
const (
	ErrCodeInternal           = "ERR_INTERNAL_SERVER"
	ErrCodeUnauthorized       = "ERR_UNAUTHORIZED"
	ErrCodeForbidden          = "ERR_FORBIDDEN"
	ErrCodeNotFound           = "ERR_NOT_FOUND"
	ErrCodeBadRequest         = "ERR_BAD_REQUEST"
	ErrCodeValidation         = "ERR_VALIDATION_FAILED"
	ErrCodeConflict           = "ERR_CONFLICT"
	ErrCodeTooManyRequests    = "ERR_TOO_MANY_REQUESTS"
	ErrCodeServiceUnavailable = "ERR_SERVICE_UNAVAILABLE"
)
