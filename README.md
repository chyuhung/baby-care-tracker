# 🍼 宝宝护理记录

一款轻量级婴儿护理记录工具，支持多用户、实时同步、Docker 一键部署。

## ✨ 功能特点

- 📝 **喂奶记录** - 支持亲喂、母乳瓶喂、配方奶，记录时长、奶量、方向
- 🧷 **尿布记录** - 记录小便、大便、混合，附备注
- 👶 **多宝宝支持** - 一个家庭可管理多个宝宝档案
- 👨‍👩‍👧 **多用户协作** - 家庭成员共享数据，实时同步
- 📱 **移动端优先** - 专为手机设计，大按钮、单手操作
- 💾 **数据持久化** - SQLite 数据库，数据永不丢失
- 🐳 **Docker 一键部署** - `docker-compose up` 即用

## 🚀 快速开始

### 方式一：Docker 部署（推荐）

```bash
# 进入项目目录
cd baby-care-tracker

# 一键启动
docker-compose up -d

# 访问
open http://localhost:8080
```

### 方式二：本地开发

**前置条件：** Node.js 18+、Go 1.21+

```bash
# 1. 启动后端
cd backend
go run .

# 2. 另起终端，启动前端
cd frontend
npm install
npm run dev
```

## 📂 项目结构

```
baby-care-tracker/
├── backend/                  # Go 后端
│   ├── main.go              # 入口 + 路由 + JWT中间件
│   ├── models/              # 数据模型
│   ├── handlers/            # API 处理器
│   │   ├── auth.go          # 认证
│   │   ├── baby.go          # 宝宝管理
│   │   ├── records.go       # 记录CRUD
│   │   └── ws.go            # WebSocket 实时同步
│   └── database/
│       └── sqlite.go        # SQLite 数据库初始化
├── frontend/                 # Vue 3 前端
│   ├── src/
│   │   ├── pages/           # 页面组件
│   │   ├── components/       # 可复用组件
│   │   ├── stores/          # Pinia 状态管理
│   │   └── api/             # API 封装
│   └── ...
├── docker-compose.yml       # Docker 编排
├── Dockerfile               # 多阶段构建
└── SPEC.md                  # 项目规格说明书
```

## 🔧 配置说明

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `PORT` | `8080` | 服务端口 |
| `DATA_DIR` | `/app/data` | SQLite 数据目录 |
| `JWT_SECRET` | （内置） | JWT 签名密钥，生产环境建议修改 |

## 📱 使用流程

1. **注册账号** - 输入用户名和密码（密码≥6位）
2. **添加宝宝** - 输入宝宝姓名、出生日期、选择性别
3. **开始记录** - 点击右下角「+」按钮，选择喂奶或尿布
4. **家庭共享** - 其他家庭成员注册相同应用的账号即可（目前为家庭共享模式，数据按宝宝归属隔离）

## 🔒 数据安全

- JWT 认证，7天有效期
- SQLite WAL 模式，支持并发读
- 数据存储在本地 Docker volume，宿主机可直接访问

## 📦 备份与迁移

```bash
# 备份数据文件
cp data/app.db data/app.db.backup-$(date +%Y%m%d)

# 迁移：将 app.db 文件复制到新服务器的 data 目录即可
```

## 🛠️ 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3 + Vite 5 + TailwindCSS + Pinia |
| 后端 | Go + Gin + SQLite |
| 数据库 | SQLite (go-sqlite3) |
| 实时同步 | WebSocket |
| 部署 | Docker + Docker Compose |
