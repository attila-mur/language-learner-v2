# Stage 1: Build the Go binary
FROM golang:1.22-alpine AS builder

# Install build dependencies for CGO + SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy all source code
COPY backend/ .

# Download dependencies and generate go.sum if needed, then build
RUN go mod tidy && \
    CGO_ENABLED=1 GOOS=linux go build -o app .

# Stage 2: Runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/app ./app

# Copy frontend files
COPY frontend/ ./frontend/

# Copy seed data
COPY backend/data/ ./data/

# Create data directory for the database
RUN mkdir -p /app/data

EXPOSE 8080

ENV DATABASE_PATH=/app/data/quiz.db
ENV FRONTEND_PATH=/app/frontend
ENV DATA_PATH=/app/data/words.json
ENV PORT=8080

CMD ["./app"]
