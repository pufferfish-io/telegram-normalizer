# syntax=docker/dockerfile:1.7

FROM golang:1.24.4 AS builder
WORKDIR /app

# Build metadata (optional)
ARG VERSION="dev"
ARG COMMIT="local"
ARG BUILT_AT

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o app ./cmd/telegram-normalizer


FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# Build metadata labels
ARG VERSION
ARG COMMIT
ARG BUILT_AT
LABEL org.opencontainers.image.title="telegram-normalizer" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILT_AT}"

COPY --from=builder /app/app /app/app

# Expose default app port if needed (no server currently)
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]

