# 🍼 宝宝护理记录

一款轻量级婴儿护理记录工具，支持家庭多人共享、实时同步、PWA 离线可用、Docker 一键部署。

## ✨ 功能特点

- 🍼 **喂奶记录** - 母乳亲喂 / 母乳瓶喂 / 配方奶，记录时长、奶量、喂养侧、品牌
- 🧷 **尿布记录** - 小便 / 大便 / 混合，附备注
- 😴 **睡眠记录** - 一键开始/结束计时，支持编辑开始与结束时间
- 🌡️ **体温记录** - 5 种测量位置，≥37.5°C 发烧提醒
- 🌳 **户外活动** - 开始/结束计时，记录户外时长
- 📊 **趋势统计** - 喂奶/尿布/睡眠/体温/户外 7 天与 30 天聚合图表
- ⏱️ **统一时间线** - 全部记录混排，按类型筛选，支持编辑与删除
- 👶 **多宝宝支持** - 一个家庭可管理多个宝宝档案，首页一键切换
- 👨‍👩‍👧 **家庭共享** - 邀请码加入家庭，成员实时同步所有记录
- 📱 **移动端优先** - 专为手机设计，iOS 风格界面、整页下拉刷新、PWA 可添加到主屏幕
- 🗺️ **时区自适应** - 按客户端时区显示与聚合，数据库统一存储 UTC
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
# 1. 启动后端（:8080）
cd backend
go run .

# 2. 另起终端，启动前端 dev server（:5173，已配置 /api 与 /ws 代理）
cd frontend
npm install
npm run dev
```

浏览器访问 `http://localhost:5173` 开发调试。

### 生产构建

```bash
cd frontend
npm run build
```

后端启动时会优先加载 `frontend/dist`（或内嵌 `dist`），`http://localhost:8080` 直接访问完整应用。

## 📂 项目结构

```
baby-care-tracker/
├── backend/                  # Go 后端
│   ├── main.go              # 入口 + 路由 + JWT 中间件 + SPA 托管
│   ├── models/              # 数据模型与请求/响应 DTO
│   ├── database/
│   │   └── sqlite.go        # SQLite 初始化（modernc.org/sqlite，纯 Go 无 CGO）
│   └── handlers/
│       ├── auth.go          # 注册 / 登录 / JWT
│       ├── baby.go          # 宝宝 CRUD + 统计 + 趋势聚合
│       ├── records.go       # 统一时间线 / 记录更新删除 / 最近记录
│       ├── sleep.go         # 睡眠 开始/结束/进行中
│       ├── temperature.go   # 体温记录
│       ├── outdoor.go       # 户外 开始/结束/进行中
│       ├── family.go        # 家庭 / 邀请码
│       ├── helpers.go       # 时区与时间解析工具
│       └── ws.go            # WebSocket 实时广播
├── frontend/                 # Vue 3 前端
│   ├── src/
│   │   ├── pages/           # 首页 / 时间线 / 趋势 / 我的 / 各类记录表单
│   │   ├── components/       # PullRefresh、RecordCard、Segmented 等
│   │   ├── stores/          # Pinia 状态管理（app / auth）
│   │   └── api/             # API 封装与类型定义
│   ├── public/              # PWA（sw.js / manifest.webmanifest / 图标）
│   └── index.html
├── docker-compose.yml       # Docker 编排
├── Dockerfile               # 多阶段构建（前端 + Go + 运行镜像）
└── SPEC.md                  # 项目规格说明书
```

## 🔧 配置说明

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `PORT` | `8080` | 服务端口 |
| `DATA_DIR` | `/app/data` | SQLite 数据目录 |
| `JWT_SECRET` | （内置） | 生产环境建议修改 `backend/handlers/auth.go` 中的签名密钥 |

> JWT 有效期 7 天；SQLite 以 UTC 存储时间，前端通过 `X-Timezone-Offset` 请求头上报时区，由后端按用户时区做当日统计与趋势聚合并本地显示。

## 📱 使用流程

1. **注册账号** - 输入用户名和密码（密码≥6位）
2. **添加宝宝** - 「我的」页进入「宝宝档案」，填写姓名、出生日期、性别
3. **开始记录** - 底部导航「记录」页四宫格卡片，一行按钮快速添加喂奶 / 尿布 / 体温，睡眠与户外可直接开始/结束计时
4. **查看与分析** - 「时间线」按类型筛选全部记录；「趋势」查看 7/30 天图表
5. **家庭共享** - 「我的」页复制邀请码，家人注册后输入邀请码即可加入，多人实时同步
6. **离线使用** - 添加 PWA 到手机主屏幕（iOS Safari：「分享 → 添加到主屏幕」），支持整页下拉刷新

## 🔒 数据安全

- JWT 认证，7 天有效期
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
| 前端 | Vue 3 + Vite 5 + TailwindCSS + Pinia + Vue Router |
| 后端 | Go + Gin |
| 数据库 | SQLite (modernc.org/sqlite，纯 Go) |
| 实时同步 | WebSocket |
| 部署 | Docker + Docker Compose |
| 移动端 | PWA（service worker + manifest，可添加到 iOS 主屏幕） |