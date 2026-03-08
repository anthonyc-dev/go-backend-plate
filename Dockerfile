#Build the binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -extldflags=-static" \
    -o app .

#Production image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/app .
# COPY --from=builder /app/.env.dev .env.prod

EXPOSE 8080

ENV TZ=UTC

CMD ["./app"]
