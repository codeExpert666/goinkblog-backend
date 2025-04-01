# GoInk Blog 后端

GoInk Blog 是一个前后端分离的个人博客系统，提供文章管理、用户认证、评论系统等功能，并集成AI增强功能，提升内容创作体验。

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

## 功能列表

- **用户与认证**
  - 用户注册、登录、登出
  - 用户信息管理：个人资料更新、头像上传
  - 基于JWT的认证系统
  - 基于Casbin的RBAC权限控制
    - 用户角色管理：添加/移除角色
    - 权限策略管理：添加/移除策略
    - 角色查询：获取所有角色、用户的角色、具有指定角色的用户
    - 权限验证与策略重载
  - 图形验证码生成与验证

- **博客核心功能**
  - 文章管理
    - 检索与搜索：关键词、分类、标签过滤，多种排序方式
    - 文章操作：创建、更新、删除、发布/草稿状态切换
    - 文章内容：标题、内容、摘要、分类、标签、封面图片上传
  - 文章交互
    - 点赞/取消点赞文章
    - 收藏/取消收藏文章
    - 用户浏览历史记录
    - 获取用户点赞、收藏、评论过的文章列表
  - 分类管理
    - 分类列表获取（分页/全部）
    - 分类详情查看
    - 分类创建、更新、删除
  - 标签管理
    - 标签列表获取（分页/全部/热门）
    - 标签详情查看
    - 标签创建、更新、删除
  
- **评论系统**
  - 发表评论：对文章进行评论
  - 回复评论：支持对现有评论的回复，形成评论树
  - 获取评论：获取文章的所有评论
  - 评论管理：获取、删除用户评论

- **统计功能**
  - 文章数据统计
  - 内容分布统计
  - 网站访问统计
  - 用户活跃度分析
  - 系统日志查询
  
- **AI增强功能**
  - AI配置管理：设置AI提供商、API密钥、端点、模型和参数
  - 文章内容润色：利用AI优化文章语言表达和结构
  - 自动生成标题建议：基于文章内容智能推荐标题选项
  - 自动生成内容摘要：提取文章要点生成摘要
  - 智能推荐标签：分析文章内容推荐相关标签

## 开发与部署

### 系统要求

- Go 1.21+ (开发使用 1.23.6)
- MySQL 8.0+
- Redis 6.0+
- 支持本地LLM模型（如Ollama）用于AI增强功能

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

GoInk Blog 目前支持本地（Ollama）和 OpenAI 两种AI模型提供商，可在配置文件中设置：

```json
"ai": {
  "provider": "local",        // 提供商类型：local、openai
  "api_key": "",              // API密钥（如有需要）
  "endpoint": "http://localhost:11434/api/generate", // API端点
  "model": "gemma3:12b",      // 使用的模型名称
  "temperature": 0.7          // 生成温度（创造性）
}
```

## 许可证

MIT
