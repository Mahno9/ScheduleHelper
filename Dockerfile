# Multi-stage build

# Stage 1: Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./frontend/
WORKDIR /app/frontend
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Backend
FROM golang:alpine AS backend-builder
WORKDIR /app/backend
COPY backend/go.mod ./
# COPY backend/go.sum ./
# RUN go mod download
COPY backend/ ./
RUN go build -o /app/server main.go

# Stage 3: Final image
FROM alpine:latest
WORKDIR /app
COPY --from=backend-builder /app/server /app/server
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

EXPOSE 8080
CMD ["/app/server"]