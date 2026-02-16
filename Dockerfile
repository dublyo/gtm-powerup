FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gtm-powerup .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /gtm-powerup /gtm-powerup

EXPOSE 8081
ENTRYPOINT ["/gtm-powerup"]
