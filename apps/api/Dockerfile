# syntax=docker/dockerfile:1.7

# Build the Go API binary.
FROM golang:1.25-alpine AS build

WORKDIR /src/apps/api

# Copy module files first so dependency downloads can be cached.
COPY apps/api/go.mod ./
RUN go mod download

# Copy the rest of the API sources and build a static binary.
COPY apps/api/ ./
RUN CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# Minimal runtime image.
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 10001 app

USER app
COPY --from=build /out/server /usr/local/bin/server

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/server"]
