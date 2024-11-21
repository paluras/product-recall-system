# Dockerfile
FROM golang:1.23-alpine

WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app

# Use multi-stage build for smaller final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and static assets from builder
COPY --from=0 /app/main .
COPY --from=0 /app/ui ./ui
COPY --from=0 /app/scripts ./scripts

# Expose port
EXPOSE 54321

# Command to run
CMD ["./main"]