.PHONY: all clean build run gen wire swagger start stop restart test

# 定义应用名称
APP_NAME = goinkblog

# GO命令
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

# 定义版本
VERSION = v1.0.0

# 定义构建参数
LDFLAGS=-ldflags "-X main.VERSION = $(VERSION)"

# 定义SWAGGER参数
SWAGGER_ARGS = -o ./internal/swagger

all: clean gen build run

# 清理构建文件
clean:
	@echo "Cleaning..."
	@rm -f $(APP_NAME)
	@echo "Cleaned!"

# 构建应用
build:
	@echo "Building..."
	@$(GOBUILD) $(LDFLAGS) -o $(APP_NAME) ./
	@echo "Build complete!"

# 生成Wire依赖注入
wire:
	@echo "Generating wire..."
	@cd ./internal/wirex && wire
	@echo "Wire generated!"

# 生成Swagger文档
swagger:
	@echo "Generating swagger docs..."
	@swag init $(SWAGGER_ARGS)
	@cp ./internal/swagger/swagger.json ./static/openapi/swagger.json
	@echo "Swagger docs generated!"

# 生成所有文件
gen: wire swagger

# 运行应用
run:
	@echo "Running..."
	@$(GOCMD) run ./main.go start -d configs -c dev -s static -daemon

# 测试应用
test:
	@echo "Testing..."
	@$(GOTEST) -v ./...

# 启动应用
start:
	@echo "Starting..."
	@bash ./scripts/start.sh

# 停止应用
stop:
	@echo "Stopping..."
	@bash ./scripts/stop.sh

# 重启应用
restart:
	@echo "Restarting..."
	@bash ./scripts/restart.sh

# 初始化项目
init:
	@echo "Initializing project..."
	@$(GOGET) -u github.com/google/wire/cmd/wire
	@$(GOGET) -u github.com/swaggo/swag/cmd/swag
	@echo "Project initialized!"

# 格式化代码
fmt:
	@echo "Formatting code..."
	@$(GOCMD) fmt ./...
	@echo "Code formatted!"

# 代码检查
lint:
	@echo "Linting code..."
	@golangci-lint run ./...
	@echo "Lint complete!"
