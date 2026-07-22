# --- Stage 1: Build React Web Frontend ---
FROM node:20-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# --- Stage 2: Build Go Server ---
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# --- Stage 3: Minimal Production Image ---
FROM alpine:3.19
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy compiled Go server binary
COPY --from=go-builder /app/server ./server

# Copy built static frontend assets to be served or embedded if needed
COPY --from=web-builder /app/web/dist ./web/dist

# Expose HTTP port
EXPOSE 8080

# Persistent Git repository volume directory
VOLUME ["/app/data/repository"]

ENV PORT=8080 \
    ENV=production \
    DATA_DIR=/app/data/repository

ENTRYPOINT ["./server"]
