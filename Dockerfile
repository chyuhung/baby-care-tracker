# ======================================
# 阶段1: 构建前端 (Node.js)
# ======================================
FROM node:20-alpine AS frontend-builder

WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm config set registry https://registry.npmmirror.com && \
    npm ci

COPY frontend/ ./
RUN node node_modules/vite/bin/vite.js build

# ======================================
# 阶段2: 构建后端 (Go)
# ======================================
FROM golang:1.21-alpine AS backend-builder

WORKDIR /build/backend

# 设置 Go 代理（国内加速）
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# 先下载 Go 模块（只依赖 go.mod/go.sum，缓存命中率高）
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制前端构建产物（供 go:embed 使用）
COPY --from=frontend-builder /build/frontend/dist ./dist

# 复制 Go 源码
COPY backend/ ./

# 纯 Go SQLite，无需 CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o baby-care-tracker .

# ======================================
# 阶段3: 运行镜像
# ======================================
FROM alpine:3.19

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 复制后端可执行文件
COPY --from=backend-builder /build/backend/baby-care-tracker .

# 数据目录
RUN mkdir -p /app/data

EXPOSE 8080

ENV PORT=8080
ENV DATA_DIR=/app/data

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

ENTRYPOINT ["./baby-care-tracker"]
