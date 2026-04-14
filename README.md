# Go REST API Example

## 项目简介
生产级 Go REST API 示例，基于 Gin 框架 + MongoDB，实现电商订单管理的 CRUD 接口。采用 Clean Architecture 分层设计，包含完整的错误处理、日志、测试。

## 快速启动

### Docker 启动（推荐）

```bash
# 克隆项目
git clone https://github.com/11DingKing/solo-zj-00061-20260415
cd solo-zj-00061-20260415

# 启动所有服务
docker compose up -d

# 查看运行状态
docker compose ps
```

### 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 后端 API | http://localhost:8080 | Go API 服务 |
| MongoDB | localhost:27022 | 数据库 |

### 停止服务

```bash
docker compose down
```

## 项目结构
- `internal/handlers/` - HTTP 处理器
- `internal/models/` - 数据模型
- `internal/db/` - 数据库操作
- `internal/server/` - 服务器配置
- `pkg/` - 公共包（日志、MongoDB 连接）

## 来源
- 原始来源: https://github.com/rameshsunkara/go-rest-api-example
- GitHub（上传）: https://github.com/11DingKing/solo-zj-00061-20260415
