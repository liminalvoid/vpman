# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY src/ .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /server /server
USER 65534:65534
EXPOSE 8080
ENTRYPOINT [ "/server" ]
