# 构建阶段
FROM golang:1.23-alpine AS build

ARG APP=goinkblog
ARG VERSION=v1.0.0

# 设置Go代理
ENV GOPROXY="https://goproxy.cn"

# 设置工作目录
WORKDIR /app

# 复制项目文件
COPY . .

# 构建应用
RUN go build -ldflags "-X main.VERSION=${VERSION}" -o ${APP} ./

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 复制构建好的应用
COPY --from=build /app/goinkblog /app/
# 复制必要的配置和静态文件
COPY --from=build /app/configs /app/configs
COPY --from=build /app/static /app/static

# 暴露端口
EXPOSE 8080

# 启动命令
ENTRYPOINT ["./goinkblog", "start", "-d", "configs", "-c", "prod", "-s", "static"]
