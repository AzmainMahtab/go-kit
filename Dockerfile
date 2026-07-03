# syntax=docker/dockerfile:1

# ============================================================
# BUILDER STAGE
# ============================================================
FROM golang:1.26-alpine AS builder

# Install build dependencies for CGO-free static binary.
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy dependency manifests first for better layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy source code.
COPY . .

# Build the API binary. CGO_ENABLED=0 produces a static binary suitable for
# alpine scratch images.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/api ./cmd/api

# Install Goose for migrations inside the final image.
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# ============================================================
# PRODUCTION STAGE
# ============================================================
FROM alpine:3.21 AS production

# Install runtime dependencies.
RUN apk add --no-cache ca-certificates postgresql-client

# Copy binaries.
COPY --from=builder /bin/api /usr/local/bin/api
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# Copy migrations and entrypoint.
COPY migrations /migrations
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Create a non-root user for security.
RUN addgroup -g 1000 -S appuser && adduser -u 1000 -S appuser -G appuser
USER appuser

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# ============================================================
# DEVELOPMENT STAGE
# ============================================================
FROM builder AS development

# Install Air for hot reload.
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Air will run the app from the mounted source volume.
CMD ["air", "-c", ".air.toml"]
