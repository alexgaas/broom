# ---- BUILD IMAGE -----
FROM golang:1.25 AS build

WORKDIR /app

# Copy bloom package first (for the replace directive)
COPY bloom/ /bloom/

# Copy cache service go.mod and go.sum
COPY bloom_example/cache/go.mod bloom_example/cache/go.sum ./

# Fix the replace directive for Docker build
RUN sed -i 's|../../bloom|/bloom|g' go.mod

RUN go mod download

COPY bloom_example/cache/ .

RUN env GOOS=linux GOARCH=amd64 go build -o cache .

# ---- RUNTIME IMAGE ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /app/cache ./cache

EXPOSE 8080

ENTRYPOINT ["./cache", "server", "start"]
