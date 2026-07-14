# ======================================
# 阶段1: 构建前端 (Node.js)
# ======================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# ======================================
# 阶段2: 构建后端 (Go)
# ======================================
FROM golang:1.21-alpine AS backend-builder

WORKDIR /build/backend

# 先复制前端构建产物到 dist 目录（供 go:embed 使用）
COPY --from=frontend-builder /build/frontend/dist ./dist

# 复制 Go 模块和源码
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# 纯 Go SQLite，无需 CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o baby-care-tracker .

# ======================================
# 阶段3: 运行镜像
# ======================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# 非 root 用户
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -s /bin/sh -D appuser

WORKDIR /app

# 复制后端可执行文件
COPY --from=backend-builder /build/backend/baby-care-tracker .

# 数据目录
RUN mkdir -p /app/data && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

ENV PORT=8080
ENV DATA_DIR=/app/data

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

ENTRYPOINT ["./baby-care-tracker"]
