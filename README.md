# GoInk Blog 后端

GoInk Blog 是一个前后端分离的个人博客系统，提供文章管理、用户认证、评论系统等功能，并集成多模型 AI 增强功能，提升内容创作体验。系统采用模块化设计，具备完善的权限控制和数据统计分析能力。

[前端项目基于 Ant Design React 实现](https://github.com/codeExpert666/goinkblog-frontend)

## 技术栈

- **框架**: [Gin](https://github.com/gin-gonic/gin) v1.10.0
- **ORM**: [GORM](https://gorm.io/) v1.25.12
- **权限**: [Casbin](https://github.com/casbin/casbin) v2.103.0
- **日志**: [Zap](https://github.com/uber-go/zap) v1.27.0
- **依赖注入**: [Wire](https://github.com/google/wire) v0.6.0
- **认证**: [JWT](https://github.com/golang-jwt/jwt) v3.2.2
- **API文档**: [Swagger](https://github.com/swaggo/swag) v1.16.4
- **命令行**: [Urfave/CLI](https://github.com/urfave/cli) v2.27.6
- **缓存**: [Redis](https://github.com/redis/go-redis) v9.7.1
- **限流**: [Redis Rate](https://github.com/go-redis/redis_rate) v10.0.1
- **系统监控**: [Gopsutil](https://github.com/shirou/gopsutil) v3.24.5
- **数据库**: MySQL v8+, Redis v6+
- **Go版本**: Go 1.23.6s
- **中间件**: CORS v1.7.3, 限流器, 身份认证, 链路追踪, 请求日志记录等

## 项目结构

```
goinkblog-backend/
├── cmd/                 # 应用程序命令
│   ├── start.go         # 启动命令
│   ├── stop.go          # 停止命令
│   └── version.go       # 版本命令
├── configs/             # 配置文件目录
│   ├── dev/             # 开发环境配置
│   ├── rbac_model.conf  # RBAC模型配置
│   └── rbac_policy.csv  # RBAC策略配置
├── internal/            # 内部包目录
│   ├── bootstrap/       # 初始化逻辑
│   ├── config/          # 配置逻辑
│   ├── mods/            # 业务模块
│   │   ├── ai/          # AI增强模块
│   │   ├── auth/        # 认证模块
│   │   ├── blog/        # 博客模块
│   │   ├── comment/     # 评论模块
│   │   └── stat/        # 统计模块
│   ├── swagger/         # Swagger文档
│   └── wirex/           # 依赖注入
├── pkg/                 # 公共包目录
│   ├── ai/              # 大模型客户端
│   ├── cachex/          # 缓存实现
│   ├── errors/          # 错误处理
│   ├── gormx/           # GORM扩展
│   ├── json/            # JSON处理工具
│   ├── jwtx/            # JWT工具
│   ├── logging/         # 日志工具
│   ├── middleware/      # HTTP中间件
│   └── util/            # 通用工具
├── scripts/             # 脚本文件
│   ├── restart.sh       # 重启脚本
│   ├── start.sh         # 启动脚本
│   └── stop.sh          # 停止脚本
├── static/              # 静态资源
│   ├── openapi/         # OpenAPI文档
│   ├── pic/             # 图片资源
│   └── index.html       # 单页应用
├── test/                # 测试文件
├── go.mod               # Go模块文件
├── go.sum               # Go依赖校验
├── main.go              # 主程序入口
├── Makefile             # 构建脚本
└── README.md            # 项目说明
```

## ✨ 功能特性

### 🔐 用户与权限
- **身份认证** — JWT 登录系统、图形验证码
- **用户管理** — 注册、个人资料、头像上传
- **权限控制** — 基于 Casbin 的 RBAC 权限系统

### 📝 内容管理
- **文章系统**
  - 创建/编辑/发布文章，支持草稿模式
  - 多维度搜索：关键词、分类、标签筛选等
  - 丰富媒体：支持封面图片上传
- **分类/标签** — 完整的分类与标签管理
- **用户互动** — 点赞、收藏、浏览历史记录

### 💬 社区功能
- **评论系统** — 支持文章评论与嵌套回复
- **用户互动** — 评论管理、互动历史

### 📊 数据分析
- **内容统计** — 文章数据、内容分布分析
- **用户分析** — 访问统计、活跃度分析
- **系统监控** — 实时跟踪CPU、内存、磁盘IO使用情况
- **性能分析** — 监测数据库性能和缓存状态
- **日志管理** — 提供系统日志查询功能

### 🤖 AI增强
- **智能创作** — 内容润色、自动生成摘要（支持流式传输）
- **智能推荐** — 标题建议、相关标签推荐
- **多模型支持** — 支持多种 AI 模型(OpenAI/Ollama)
- **模型管理** — 动态加载模型、负载均衡
- **请求限流** — 基于令牌桶的 AI 请求限流
  
### 🛠️ 管理功能（管理员）
- **权限管理** — Casbin 权限系统，支持角色分配、权限策略配置、权限验证
- **分类管理** — 文章分类的创建、编辑、删除，支持分类统计
- **标签管理** — 文章标签的创建、编辑、删除，支持标签统计
- **评论管理** — 评论查询与审核，支持评论统计
- **AI模型管理** — 模型配置添加、权重调整、限流幅度调节、模型性能监控


## 开发与部署

### 系统要求

- Go 1.21+ (开发使用 1.23.6)
- MySQL 8.0+
- Redis 6.0+
- 支持 Ollama 本地 LLM 模型用于 AI 增强功能
- 支持 OpenAI API 或其他兼容 API 接入

### 开发工具

- [swag](https://github.com/swaggo/swag): API文档生成
- [wire](https://github.com/google/wire): 依赖注入代码生成
- [golangci-lint](https://github.com/golangci/golangci-lint): 代码质量检查

### 项目初始化

```bash
# 初始化项目依赖
make init

# 格式化代码
make fmt

# 代码质量检查
make lint
```

### 生成代码

```bash
# 生成依赖注入代码
make wire

# 生成Swagger文档
make swagger

# 生成所有代码
make gen
```

### 构建与运行

```bash
# 构建应用
make build

# 运行开发环境
make run

# 启动服务（构建+运行）
make start

# 停止服务
make stop

# 重启服务
make restart
```

### API文档

启动应用后，访问以下地址查看后端服务信息（端口可在 `configs` 目录中配置）：

```
http://localhost:8080/
```

### 部署指南

#### 本地部署

1. 构建应用
```bash
make build
```

2. 启动应用
```bash
./goinkblog start -d configs -c dev -s static -daemon
```

参数说明：
- `-d`：工作目录，默认为 `configs`
- `-c`：配置目录 **（相对于工作目录）**，默认为 `dev`
- `-s`：静态文件目录，默认为 `static`
- `-daemon`：开启守护进程模式

#### 生产环境

生产环境部署建议：

1. 创建生产环境配置目录 `configs/prod/`
2. 在该目录中定制生产环境配置文件
3. 使用以下命令启动:
```bash
./goinkblog start -d configs -c prod -s static -daemon
```

#### Docker部署

1. 构建Docker镜像
```bash
# 构建镜像，默认版本为v1.0.0
docker build -t goinkblog:v1.0.0 .

# 指定应用版本构建
docker build --build-arg VERSION=v1.1.0 -t goinkblog:v1.1.0 .
```

2. 运行Docker容器
```bash
# 使用开发环境配置启动（默认）
docker run -d --name goinkblog -p 8080:8080 goinkblog:v1.0.0

# 使用生产环境配置启动
docker run -d --name goinkblog -p 8080:8080 -e CONFIG_DIR=prod goinkblog:v1.0.0

# 挂载外部配置目录和静态资源目录
docker run -d --name goinkblog \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/static:/app/static \
  goinkblog:v1.0.0
```

## AI功能配置

GoInk Blog 支持多种AI模型提供商，采用模型选择器动态管理多个模型，可在配置文件中设置：

```json
"ai": {
  "models": [
    {
      "provider": "local",
      "api_key": "ollama",
      "endpoint": "http://localhost:11434/v1/chat/completions",
      "model_name": "gemma3:12b",
      "temperature": 0.7,
      "timeout": 90,
      "active": true,
      "description": "Ollama model",
      "rpm": 10,
      "weight": 100
    }
  ],
  "selector": {
    "load_models_interval": 5,
    "update_weight_interval": 2
  }
}
```

配置说明：
- `models`: 支持配置多个AI模型
  - `provider`: 模型提供商，目前支持 "local"(Ollama) 和 "openai"
  - `api_key`: API密钥
  - `endpoint`: API端点
  - `model_name`: 模型名称
  - `temperature`: 温度参数，控制生成文本的随机性
  - `timeout`: 请求超时时间（秒）
  - `active`: 是否启用该模型
  - `description`: 模型描述
  - `rpm`: 每分钟请求限制
  - `weight`: 模型权重，用于负载均衡
- `selector`: 模型选择器配置
  - `load_models_interval`: 从数据库加载模型间隔（单位为分钟）
  - `update_weight_interval`: 权重更新间隔（用于模型负载均衡，单位为分钟）

## 许可证

MIT
