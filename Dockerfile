# 第一阶段：构建阶段
FROM golang:1.20-alpine AS builder

WORKDIR /app

# 添加构建参数
ARG SERVER_PORT=3000

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 初始化模块
RUN go mod download && \
    go mod tidy

# 复制源代码
COPY . .

# 编译为静态二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 添加镜像元数据
LABEL org.opencontainers.image.title="Media Processor"
LABEL org.opencontainers.image.description="A Go service for processing media files, featuring file upload and OCR capabilities. Built with Gin framework and Aliyun OCR API, supporting single/batch file uploads and text recognition from images."
LABEL org.opencontainers.image.source="https://github.com/harrisonwang/media-processor"
LABEL org.opencontainers.image.licenses="MIT"

# 安装 CA 证书
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/server /app/

# 暴露动态端口
EXPOSE ${SERVER_PORT}
ENTRYPOINT ["/app/server"]
