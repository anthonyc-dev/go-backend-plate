# Build the binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev git

COPY go.mod go.sum ./
COPY vendor ./vendor 

COPY . .

# Use vendor modules, avoid fetching from internet
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor \
    -ldflags="-s -w -extldflags=-static" \
    -o app .

# Production image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

ENV TZ=UTC

CMD ["./app"]