# Step 1: Build React SPA Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Step 2: Build Go Backend Binary
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Step 3: Production Runtime Container
FROM alpine:3.19
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /app

# Copy Go binary & static frontend bundle
COPY --from=backend-builder /app/server /app/server
COPY --from=frontend-builder /app/web/dist /app/web/dist

# Create repository data directory
RUN mkdir -p /app/data/repository && git config --global user.name "Chorus" && git config --global user.email "system@chorus.local"

ENV PORT=8080
ENV ENV=production
ENV DATA_DIR=/app/data/repository

EXPOSE 8080

CMD ["/app/server"]
