FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git for go mod
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app for Linux AMD64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/api

# Final image
FROM alpine:latest

WORKDIR /app

COPY .firebase ./.firebase
COPY .env.development ./.env.development
COPY .env.production ./.env.production

COPY --from=builder /app/main .

CMD ["./main"]