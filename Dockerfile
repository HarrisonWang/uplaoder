# 第一阶段：构建阶段
FROM golang:1.20-alpine AS builder

WORKDIR /app

# 只复制 go.mod，不再引用 go.sum
COPY go.mod ./

# 初始化模块并下载依赖
RUN go mod download
RUN go mod tidy

# 复制源代码
COPY . .

# 编译为静态二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# 第二阶段：运行阶段
FROM scratch

WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/server /app/

EXPOSE 3000
ENTRYPOINT ["/app/server"]
