# 🍼 宝宝护理记录 — 项目规格说明

## 1. 项目概述

**项目名称:** Baby Care Tracker（宝宝护理记录）
**类型:** 全栈 Web 应用（移动端优先 SPA）
**核心功能:** 记录婴儿喂奶、尿布更换，支持多用户家庭共享
**目标用户:** 有婴儿的家庭成员（0-3岁宝宝）

## 2. 技术架构

| 层 | 技术 | 说明 |
|----|------|------|
| 前端 | Vue 3 + Vite 5 + TypeScript + TailwindCSS | 移动端 SPA |
| 状态管理 | Pinia | JWT 持久化、宝宝管理 |
| 后端 | Go 1.21 + Gin | REST API + WebSocket |
| 数据库 | SQLite（modernc.org/sqlite） | WAL 模式，纯 Go 无 CGO |
| 实时同步 | WebSocket | 记录变化实时广播 |
| 部署 | Docker + Docker Compose | 单镜像一键部署 |

## 3. 功能清单

### 3.1 认证模块
- ✅ 用户注册（用户名 + 密码 ≥6位）
- ✅ 用户登录（返回 JWT token，7天有效）
- ✅ JWT 中间件保护所有业务 API

### 3.2 宝宝管理
- ✅ 创建宝宝档案（姓名、出生日期、性别）
- ✅ 编辑宝宝信息
- ✅ 删除宝宝（级联删除所有记录）
- ✅ 宝宝列表（支持多宝宝）

### 3.3 喂奶记录
- ✅ 类型：亲喂（breast）、母乳瓶喂（bottle）、配方奶（formula）
- ✅ 亲喂：记录时长（分钟）+ 方向（左/右/双侧）
- ✅ 瓶喂/配方奶：记录奶量（ml）+ 品牌
- ✅ 备注字段

### 3.4 尿布记录
- ✅ 类型：小便（pee）、大便（poop）、混合（both）
- ✅ 备注字段

### 3.5 统计
- ✅ 今日喂奶次数、尿布次数、总奶量
- ✅ 最近一次喂奶时间、尿布时间
- ✅ 最近喂奶记录（用于预填表单）

### 3.6 时间线
- ✅ 按宝宝筛选所有记录（喂奶+尿布合并）
- ✅ 按类型筛选
- ✅ 按日期筛选
- ✅ 按时间倒序排列

### 3.7 实时同步
- ✅ WebSocket 连接
- ✅ 记录创建/删除实时推送
- ✅ 宝宝档案变更推送
- ✅ 前端 Toast 通知

## 4. API 路由

```
公开:
  POST /api/auth/register      注册
  POST /api/auth/login         登录
  GET  /api/health             健康检查

认证 (需 Bearer Token):
  GET    /api/me                                当前用户
  GET    /api/babies                            宝宝列表
  POST   /api/babies                            创建宝宝
  PUT    /api/babies/:id                        更新宝宝
  DELETE /api/babies/:id                        删除宝宝
  GET    /api/babies/:id/stats                  今日统计
  GET    /api/babies/:id/latest-feeding         最近喂奶
  GET    /api/babies/:id/records                记录时间线
  POST   /api/babies/:id/feeding                记录喂奶
  POST   /api/babies/:id/diaper                 记录尿布
  PUT    /api/records/:id                       更新记录
  DELETE /api/records/:id                       删除记录

WebSocket:
  WS /ws                                          实时推送
```

## 5. 数据模型

### users
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| username | TEXT UNIQUE | |
| password_hash | TEXT | bcrypt |
| created_at | DATETIME | |

### babies
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| user_id | INTEGER FK | |
| name | TEXT | |
| birth_date | DATE | |
| gender | TEXT | male/female/other |
| avatar_color | TEXT | 头像颜色 |
| created_at | DATETIME | |

### feeding_records
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| baby_id | INTEGER FK | |
| user_id | INTEGER FK | |
| type | TEXT | breast/bottle/formula |
| duration_minutes | INTEGER | |
| amount_ml | INTEGER | |
| side | TEXT | left/right/both |
| brand | TEXT | |
| note | TEXT | |
| occurred_at | DATETIME | |
| created_at | DATETIME | |

### diaper_records
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| baby_id | INTEGER FK | |
| user_id | INTEGER FK | |
| type | TEXT | pee/poop/both |
| note | TEXT | |
| occurred_at | DATETIME | |
| created_at | DATETIME | |

## 6. 前端页面

| 页面 | 路由 | 说明 |
|------|------|------|
| 登录页 | /login | 登录 + 注册切换 |
| 首页 | / | 统计卡片 + 最近记录 |
| 时间线 | /timeline | 全部记录时间线 |
| 宝宝档案 | /profile | 宝宝列表 + 管理 |
| 宝宝表单 | /baby/new, /baby/:id/edit | 添加/编辑宝宝 |
| 记录表单 | /record/new | 喂奶/尿布记录表单 |

## 7. 部署方式

```bash
# Docker 一键启动
docker-compose up -d
open http://localhost:8080
```

## 8. 数据库

- 路径: `./data/app.db`（Docker 中为 `/app/data/app.db`）
- 模式: WAL（支持并发读）
- 外键: 启用（`FOREIGN KEY ON DELETE CASCADE`）
- 备份: 复制 `app.db` 文件即可
