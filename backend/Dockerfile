# multi-stage build: build Go binary
FROM golang:1.21-alpine as builder
RUN apk add --no-cache git build-base
WORKDIR /src
COPY . .
RUN go build -o /out/admin ./cmd/admin-main

# runtime stage: tiny, non-root user
FROM alpine:3.19
RUN adduser -D -g '' appuser
COPY --from=builder /out/admin /usr/local/bin/admin
WORKDIR /app
RUN chown appuser:appuser /usr/local/bin/admin
USER appuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/admin"]