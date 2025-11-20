FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
# COPY go.sum ./ # Will be created later
# RUN go mod download

COPY . .

RUN go build -o secure-fm main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/secure-fm .
# Create sandbox directory
RUN mkdir -p /app/sandbox

# Install sqlite client if needed for debugging, but we use postgres
# RUN apk add --no-cache postgresql-client

CMD ["./secure-fm"]
