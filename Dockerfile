FROM golang:1.24-alpine AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -ldflags="-s -w" \
    -o /usr/local/bin/orchestrator \
    ./cmd

FROM alpine:latest

RUN apk add --no-cache ca-certificates curl unzip

RUN curl -sL \
      https://github.com/redpanda-data/redpanda/releases/latest/download/rpk-linux-amd64.zip \
    -o /tmp/rpk.zip \
 && unzip -j /tmp/rpk.zip rpk \
 && mv rpk /usr/local/bin/rpk \
 && rm /tmp/rpk.zip

COPY --from=builder /usr/local/bin/orchestrator /usr/local/bin/orchestrator

EXPOSE 8080

ENTRYPOINT ["orchestrator"]
