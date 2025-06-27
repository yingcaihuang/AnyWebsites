FROM golang:1.23-alpine AS builder

WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 整理依赖
RUN go mod tidy

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

# 运行阶段
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

# 复制web目录（模板和静态文件）
COPY --from=builder /app/web ./web

# 复制docs目录（API文档）
COPY --from=builder /app/docs ./docs

# 创建上传目录
RUN mkdir -p uploads

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./main"]
