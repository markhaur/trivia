# Build stage
FROM golang:1.19-alpine3.16 AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o trivia cmd/main.go

# Run stage
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/trivia .
COPY docker.env .

EXPOSE 8082
CMD ["/app/trivia"]