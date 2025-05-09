# 构建阶段
FROM golang:1.23-alpine AS build

# 安装必要的工具
RUN apk add --no-cache git make

# 设置Go代理
ENV GOPROXY="https://goproxy.cn"

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 安装wire和swag
RUN go install github.com/google/wire/cmd/wire@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 复制项目文件
COPY . .

# 生成依赖注入代码和Swagger文档
RUN make gen

# 构建应用
RUN make build

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 复制构建好的应用
COPY --from=build /app/goinkblog /app/
# 复制必要的配置和静态文件
COPY --from=build /app/configs /app/configs
COPY --from=build /app/static /app/static

# 暴露端口
EXPOSE 8080

# 启动命令
ENTRYPOINT ["./goinkblog", "start", "-d", "configs", "-c", "prod", "-s", "static"]
