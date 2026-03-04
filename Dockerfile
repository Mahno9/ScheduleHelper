# ScheduleHelper — Multistage Docker build
# Stage 1: Build Svelte frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --silent
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app/backend
# CGO not needed with modernc sqlite
ENV CGO_ENABLED=0
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN go build -ldflags="-s -w" -o /schedulehelper .

# Stage 3: Final minimal image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /schedulehelper /app/schedulehelper
COPY --from=frontend-builder /app/backend/static /app/static

ENV PORT=8080
ENV DB_PATH=/data/schedulehelper.db
ENV STATIC_DIR=/app/static

EXPOSE 8080
VOLUME /data

ENTRYPOINT ["/app/schedulehelper"]
