# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dnstester ./cmd/dnstester

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/dnstester .

# Create non-root user
RUN addgroup -g 1000 dnstester && \
    adduser -D -u 1000 -G dnstester dnstester && \
    chown -R dnstester:dnstester /app

USER dnstester

EXPOSE 8080

ENTRYPOINT ["./dnstester"]

