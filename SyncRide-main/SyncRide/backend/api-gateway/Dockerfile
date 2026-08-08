# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git tzdata ca-certificates

# We use Go Workspaces (go.work) so we need to copy the entire workspace to build properly
# Copy go.work and all module directories
COPY go.work .
COPY api-gateway/ ./api-gateway/
COPY shared/ ./shared/

# Build the API Gateway binary
WORKDIR /app/api-gateway
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/api-gateway ./cmd/main.go

# Run Stage (Distroless / Minimal Alpine)
FROM alpine:latest

WORKDIR /app

# Add required certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user for security
RUN adduser -D -g '' appuser
USER appuser

# Copy the binary from the builder stage
COPY --from=builder /go/bin/api-gateway /app/api-gateway

# Expose standard gateway port
EXPOSE 8080
EXPOSE 9090

# Entrypoint
ENTRYPOINT ["/app/api-gateway"]
