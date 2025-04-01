# GoInk Blog 后端

GoInk Blog 是一个前后端分离的个人博客系统，提供文章管理、用户认证、评论系统等功能，并集成 AI 增强功能，提升内容创作体验。

## 技术栈

- **框架**: [Gin](https://github.com/gin-gonic/gin) v1.10.0
- **ORM**: [GORM](https://gorm.io/) v1.25.12
- **权限**: [Casbin](https://github.com/casbin/casbin) v2.103.0
- **日志**: [Zap](https://github.com/uber-go/zap) v1.27.0
- **依赖注入**: [Wire](https://github.com/google/wire) v0.6.0
- **认证**: [JWT](https://github.com/golang-jwt/jwt) v3.2.2
- **API文档**: [Swagger](https://github.com/swaggo/swag) v1.16.4
- **命令行**: [Urfave/CLI](https://github.com/urfave/cli) v2.27.6
- **数据库**: MySQL v8+, Redis v6+
- **Go版本**: Go 1.23.6
- **中间件**: CORS, 限流器, 身份认证, 链路追踪, 请求日志记录等

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
│   └── pic/             # 图片资源
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
- **系统监控** — 日志查询与分析

### 🤖 AI增强
- **智能创作** — 内容润色、自动生成摘要
- **智能推荐** — 标题建议、相关标签推荐
- **灵活配置** — 支持多种 AI 模型(OpenAI/Ollama)

## 开发与部署

### 系统要求

- Go 1.21+ (开发使用 1.23.6)
- MySQL 8.0+
- Redis 6.0+
- 支持 Ollama 本地 LLM 模型用于 AI 增强功能

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

# 启动服务（通过脚本，后台运行）
make start

# 停止服务
make stop

# 重启服务
make restart
```

### API文档

启动应用后，访问以下地址查看后端服务信息（端口可在 `configs` 目录中配置）：

```
http://localhost:52443/
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
- `-d`：配置文件目录，默认为 `configs`
- `-c`：配置环境，默认为 `dev`
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

## AI功能配置

GoInk Blog 目前支持 local（本地 Ollama）和 openai 两种AI模型提供商，可在配置文件中设置：

```json
"ai": {
  "provider": "local",
  "api_key": "",
  "endpoint": "http://localhost:11434/api/generate",
  "model": "gemma3:12b",
  "temperature": 0.7
}
```

## 许可证

MIT
