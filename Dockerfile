# 第一阶段：构建阶段
FROM golang:1.20-alpine AS builder

WORKDIR /app

# 将 go.mod 和 go.sum 复制进来并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译为静态二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

# 第二阶段：运行阶段
FROM scratch

WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/server /app/

# 如果需要预先创建目录，可在此进行
# 但通常通过启动时挂载或在运行时创建即可
# RUN mkdir -p /app/upload

ENV UPLOAD_PATH=/app/images
ENV BASE_URL=https://voxsay.com/upload/

EXPOSE 3000
ENTRYPOINT ["/app/server"]
