# Monitor 系统

这是一个基于 Go 语言开发的监控与管理系统，提供服务状态监控、任务管理、导航管理、命令执行等功能。

## 功能特性

- **服务与端点监控**：支持定时对指定的 HTTP 端点进行健康检查，记录响应状态和延迟。
- **任务管理 (Todo)**：支持项目制的待办事项管理。
- **导航管理**：支持自定义导航链接及其显示顺序。
- **命令执行**：支持通过模板或直接输入执行系统命令（支持 sudo）。
- **代码项目管理**：管理本地或远程代码项目的元数据。
- **账号密码管理**：安全的存储和管理账号信息。
- **身份验证**：基于 JWT 的登录验证。

## 环境变量配置

在运行项目前，请确保配置了以下环境变量：

| 变量名 | 说明 | 默认值 / 示例 |
| :--- | :--- | :--- |
| `MONGO_URI` | 完整的 MongoDB 连接字符串（优先使用） | `mongodb://user:pass@localhost:27017` |
| `MONGO_HOST` | MongoDB 主机地址（未提供 URI 时使用） | `localhost:27017` |
| `MONGO_USERNAME` | MongoDB 用户名 | - |
| `MONGO_PASSWORD` | MongoDB 密码 | - |
| `JWT_SECRET` | 用于签名 JWT Token 的密钥 | - |
| `ADMIN_USERNAME` | 管理员登录用户名 | - |
| `ADMIN_PASSWORD` | 管理员登录密码 | - |
| `SUDO_PASSWORD` | 执行 sudo 命令时需要的密码 | - |
| `GIN_MODE` | 设置为 `release` 开启生产模式 | `debug` |
| `ALLOWED_ORIGIN` | CORS 允许的来源（仅在 release 模式生效） | `*` 或特定域名 |
| `PORT` | 服务启动端口 | `8080` |

## 快速开始

### 1. 依赖项

- Go 1.24+
- MongoDB
  - 数据库名：`health_check`
  - 必须包含一个名为 `check_results` 的 Time Series 集合（项目启动时会检查）。

### 2. 运行

```bash
# 安装依赖
go mod download

# 运行服务
go run main.go
```

### 3. API 端点

- 登录：`POST /login`
- 健康检查：`GET /health`
- 业务 API（需鉴权）：`/api/...`

## 技术栈

- 框架：[Gin](https://github.com/gin-gonic/gin)
- 数据库驱动：[mongo-go-driver](https://go.mongodb.org/mongo-driver)
- 认证：[JWT (golang-jwt)](https://github.com/golang-jwt/jwt)
