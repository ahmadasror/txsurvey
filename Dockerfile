# syntax=docker/dockerfile:1

# --- Stage 1: build the SPA (no-op until frontend/ exists) ---
FROM node:22-alpine AS fe
WORKDIR /fe
# The build context always has a frontend/ dir once Phase 1 scaffolds it.
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build   # -> /fe/dist

# --- Stage 2: build the Go server, embedding the SPA + migrations ---
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=fe /fe/dist ./internal/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -tags embedspa -o /bin/server ./cmd/server

# --- Stage 3: runtime ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D appuser
COPY --from=builder /bin/server /bin/server
USER appuser
EXPOSE 8080
ENTRYPOINT ["/bin/server"]
