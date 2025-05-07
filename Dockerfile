# Stage 1: build
FROM golang:1.21-alpine AS builder

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY GO/go.mod GO/go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY GO/. .

# Build the Go app
RUN go build -o mygame main.go

# Stage 2: final image
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the compiled binary from builder
COPY --from=builder /app/mygame .

# Expose the server port (if server.go listens on 8080)
EXPOSE 8080

# Run the server
CMD ["./mygame", "server"]
