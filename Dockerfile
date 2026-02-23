# Multi-stage build for ByteFreezer Fakedata
# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=unknown
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -o /bytefreezer-fakedata .

# Stage 2: Minimal runtime image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && rm -rf /var/cache/apk/*

RUN addgroup -g 1000 -S bytefreezer && \
    adduser -u 1000 -S bytefreezer -G bytefreezer -s /bin/sh -D

COPY --from=builder /bytefreezer-fakedata /bytefreezer-fakedata
RUN chmod +x /bytefreezer-fakedata

USER bytefreezer

ENTRYPOINT ["/bytefreezer-fakedata"]
CMD ["syslog", "--host", "proxy", "--port", "5514", "--rate", "10"]

ARG VERSION=unknown
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown
LABEL maintainer="ByteFreezer Team" \
      org.opencontainers.image.title="ByteFreezer Fakedata" \
      org.opencontainers.image.description="Synthetic data generator for testing ByteFreezer pipelines" \
      org.opencontainers.image.vendor="ByteFreezer" \
      org.opencontainers.image.source="https://github.com/bytefreezer/fakedata" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.revision="${GIT_COMMIT}"
