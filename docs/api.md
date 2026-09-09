# Portal Go API

与原 Node Portal 兼容。业务路由同时挂在 **`/`** 与 **`/portal/`** 下；线上 Nginx 通常走 `/portal/…`。下文路径均写 `/portal` 前缀。

管理后台静态资源：`GET /manager/*`（非 API）。

---

## 约定

### 鉴权

| 来源 | 说明 |
|------|------|
| Header `Diary-Token` | 登录返回的 `password` 字段（bcrypt 哈希字符串），也可用 query `?token=` |
| Header `Diary-Uid` | 用户 `uid`（必填；缺省会提示升级后重登） |

服务端按 `users.password = token AND uid = ?` 校验，**不是 JWT**。  
`group_id == 1` 为管理员。多数接口在 handler 内自行鉴权，无全局中间件。

### 响应体

几乎所有业务接口 HTTP 状态码为 **200**，用 `success` 区分成败：

```json
{ "success": true, "message": "请求成功", "data": {} }
```

例外：

- `GET /`、`GET /portal/`：健康检查 `{ status, message, title }`
- `GET /portal/init/`：返回 HTML
- 部分五笔空 stub：空 body

### 尾斜杠

全局关闭了 Gin 的尾斜杠 301。根路径类接口（如 `user-config`、`system-config`）已同时注册有无 `/` 两种路径；其余路径请按文档写死（一般**不要**尾斜杠）。

### Base URL 示例

| 环境 | 示例 |
|------|------|
| 生产 | `https://kylebing.cn/portal` |
| 本地 | `http://localhost:3000/portal` |

---

## 一览

| 模块 | 前缀 | 说明 |
|------|------|------|
| [用户](#用户-user) | `/user` | 注册登录、资料 |
| [用户配置](#用户配置-user-config) | `/user-config` | 主题 / 默认分类 / 编辑器 |
| [系统配置](#系统配置-system-config) | `/system-config` | 演示账号、七牛、和风等 |
| [邀请码](#邀请码-invitation) | `/invitation` | 注册邀请 |
| [日记](#日记-diary) | `/diary` | 核心日记 CRUD |
| [日记分类](#日记分类-diary-category) | `/diary-category` | 分类字典 |
| [账单](#账单-bill) | `/bill` | 解析 `category=bill` 日记 |
| [银行卡](#银行卡-bank-card) | `/bank-card` | 固定标题日记内容 |
| [二维码](#二维码-qr) | `/qr-front` `/qr-manager` | 车贴二维码 |
| [地图路线](#地图路线-map-route) | `/map-route` | 路线 |
| [地图标点](#地图标点-map-pointer) | `/map-pointer` | 标点 |
| [五笔](#五笔-wubi) | `/wubi/dict` `/wubi/word` `/wubi/category` | 词库 / 词条 / 分类 |
| [饥荒](#饥荒-starve-new) | `/starve-new` | `starve_advance` 资料库 |
| [统计](#统计-statistic) | `/statistic` | 概览与图表数据 |
| [文件](#文件-file-manager) | `/file-manager` | 本地上传 |
| [七牛图床](#七牛图床-image-qiniu) | `/image-qiniu` | Token 与图片元数据 |
| [点赞](#点赞-thumbs-up) | `/thumbs-up` | 站点点赞计数 |
| [安装](#安装-setup--init) | `/setup` `/init` | 首次安装引导 |

鉴权图例：**P** 公开 · **A** 需登录 · **Admin** 管理员 · **Cond** 公开内容可匿名，否则需主人 · **Opt** 有 token 则带身份

---

## 用户 `/user`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/register` | P | 注册。已有用户时需邀请码（或系统全局邀请码）。首个用户为管理员 |
| POST | `/login` | P | 登录；`data` 含用户信息，其中 `password` 即后续 `Diary-Token` |
| GET | `/avatar` | P | Query `email` → 头像字段 |
| GET | `/detail` | Cond | Query `hash`：按二维码 hash 查（历史兼容） |
| POST | `/list` | A | Body `pageNo,pageSize`；管理员看全部，否则仅自己 |
| POST | `/add` | P | 创建用户（含 `group_id`）；**当前无鉴权** |
| PUT | `/set-profile` | A | `nickname,phone,avatar,city,geolocation`（演示账号受限） |
| PUT | `/modify` | A | 改资料；本人或管理员 |
| DELETE | `/delete` | Admin | Body `uid` |
| PUT | `/change-password` | A | Body `password`；改密后 token 变化 |
| DELETE | `/destroy-account` | A | 注销并清理关联数据 |

**注册 / 登录常用字段：** `email`, `password`, `username`, `nickname`, `invitationCode`, …

---

## 用户配置 `/user-config`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/user-config` 或 `/user-config/` | A | 读取配置 |
| PUT | 同上 | A | 部分更新：`theme`, `default_diary_category`, `editor_mode`, `config_json` |

---

## 系统配置 `/system-config`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/system-config` | P | 公开配置（演示账号、七牛公有信息、和风 key、注册提示等；**不含** invitation_code / 七牛密钥） |
| GET | `/system-config/admin` | Admin | 完整配置 |
| PUT | `/system-config` | Admin | 保存 |

主要字段：`is_show_demo_account`, `demo_account`, `demo_account_password`, `invitation_code`, `qiniu_*`, `hefeng_weather_*`, `register_tip`

---

## 邀请码 `/invitation`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/list` | Opt | 未绑定邀请码；非管理员仅 `is_shared=0` |
| POST | `/generate` | Admin | 生成 |
| POST | `/mark-shared` | Admin | Body `id` |
| DELETE | `/delete` | Admin | Query `id` |

---

## 日记 `/diary`

标题/内容含 emoji 时服务端以 `\uXXXX` 转义存入（utf8mb3 兼容），读写自动编解码。

### 列表共用 Query

| 参数 | 说明 |
|------|------|
| `keywords` | JSON 字符串数组，如 `["关键词"]` |
| `categories` | JSON 数组，分类 `name_en` |
| `filterShared` | `1` 时筛公开 |
| `timeStart` / `timeEnd` | 日期范围 |
| `pageNo` / `pageSize` | 分页（仅 `/list`） |

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/list` | A | 分页列表（含账单日解析） |
| GET | `/list-all` | A | 全量匹配 |
| GET | `/list-title-only` | A | 仅 id/date/title/category |
| GET | `/list-category-only` | A | 仅 id/date/category |
| GET | `/export` | A | 导出 |
| GET | `/temperature` | A | life 类气温，最多 100 条 |
| GET | `/detail` | Cond | Query `diaryId`；公开或主人 |
| POST | `/add` | A | 见下表 |
| PUT | `/modify` | A | 需 `id` + 同字段 |
| DELETE | `/delete` | A | Body `diaryId` |
| GET | `/latest-recommend` | P | 标题含「首页推荐」的公开日记（uid=3） |
| GET | `/get-diary-content-with-keyword` | A | Query `keyword`，本人最新匹配标题 |
| GET | `/get-latest-public-diary-with-keyword` | P | Query `keyword`，最新公开 |
| POST | `/clear` | A | 清空本人日记（演示账号受限） |

**add / modify Body：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` / `content` | string | |
| `category` | string | 分类 `name_en` |
| `weather` | string | |
| `temperature` / `temperature_outside` | number | |
| `date` | string | 日记日期 |
| `is_public` / `is_markdown` | int | 0/1 |

---

## 日记分类 `/diary-category`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/list` | P | 全部分类 |
| POST | `/add` | Admin | `name,name_en,color,sort_id` |
| PUT | `/modify` | Admin | 按 `name_en` 更新 |
| DELETE | `/delete` | Admin | Body `name_en` |

---

## 账单 `/bill`

解析本人 `category = bill` 的日记正文（每行 `品名 金额`）。

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/bill/` | A | 全部账单日结构化数据 |
| GET | `/sorted` | A | Query `years`（逗号分隔）、`keyword` |
| GET | `/keys` | A | 近 5 年条目词频 |
| GET | `/day-sum` | A | 日收支合计 |
| GET | `/month-sum` | A | 月合计；可选 `keyword` |
| GET | `/borrow` | A | 标题为「借还记录」的日记内容 |

---

## 银行卡 `/bank-card`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/bank-card` | A | 标题为「我的银行卡列表」的日记 `content` |

---

## 二维码 QR

### 前台 `/qr-front`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/qr-front/` | P | Query `hash`；公开二维码详情，并增加 `visit_count` |

### 管理 `/qr-manager`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/list` | A | Query `keywords`(JSON), `pageNo`, `pageSize` |
| GET | `/detail` | Cond | Query `hash` |
| POST | `/add` | A | `hash`, `message`, `description`, `is_public`, 车辆/微信相关字段等 |
| PUT | `/modify` | A | 同上 + 权限校验 |
| DELETE | `/delete` | A | Body `hash` |
| POST | `/clear-visit-count` | Admin | Body `hash` |

---

## 地图路线 `/map-route`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/list` | Opt | Body：`isMine`, `keyword`, `dateRange`, `pageNo`, `pageSize`；无身份仅公开 |
| GET | `/detail` | Cond | Query `id` |
| POST | `/add` | A | `name,area,road_type,policy,seasons,video_link,paths,note,thumb_up` |
| PUT | `/modify` | A | + `id`；主人或管理员 |
| DELETE | `/delete` | A | Body `id` |

---

## 地图标点 `/map-pointer`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/list` | Opt | `keyword,dateRange,pageNo,pageSize` |
| GET | `/detail` | Cond | Query `id` |
| POST | `/add` | A | `name,pointers,note,area,thumb_up,is_public` |
| PUT | `/modify` | A | + `id` |
| DELETE | `/delete` | A | Body `id` |

---

## 五笔 `/wubi`

### 词库备份 `/wubi/dict`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/pull` | A | Query `title` |
| PUT | `/push` | A | `title,content,contentSize,wordCount` |
| GET/POST | `/check-backup-exist` | A | `fileName`（query 或 JSON） |

### 词条 `/wubi/word`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/list` | A | `keyword,dateRange,category_id,approved,pageNo,pageSize` |
| POST | `/export-extra` | A | 已审核且非默认分类词 |
| POST | `/check-exist` | A | `word,code` |
| POST | `/add` | A | `word,code,priority,up,down,comment,category_id` |
| POST | `/add-batch` | A | `words[],category_id` |
| PUT | `/modify` | A | `id` + 字段；非管理员仅本人 |
| DELETE | `/delete` | A | `ids[]` |
| PUT | `/modify-batch` | A | `ids[]`, 可选 `category_id`,`approved` |
| POST | `/upload-dict` | — | 已废弃，返回错误 |
| GET | `/statistic` `/thumbs-up` `/thumbs-down` | — | 空 stub |

### 分类 `/wubi/category`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/list` | A | 分类及词数 |
| POST | `/add` | Admin | `name,sort_id` |
| PUT | `/modify` | Admin | `id,name,sort_id` |
| DELETE | `/delete` | Admin | `id` |

---

## 饥荒 `/starve-new`

库：`starve_advance`。实体：`character` `mob` `log` `plant` `thing` `material` `craft` `cookingrecipe` `coder` `version` `craft-tab`。

| 方法 | 路径模式 | 鉴权 | 说明 |
|------|----------|------|------|
| GET | `/{entity}/list` | P | Query：`keyword`, `version`, `tab`(craft), `all=1`, 可选 `pageNo`+`pageSize` |
| GET | `/{entity}/info` | P | Query `id` |
| POST | `/{entity}/add` | A | 实体字段；部分需 `version` / `tab` id 列表 |
| PUT | `/{entity}/modify` | A | 需 `id` |
| DELETE | `/{entity}/delete` | A | Body 或 Query `id` |

示例：`GET /portal/starve-new/craft/list?keyword=斧&version=1`

> 经典 `/starve` 已下线，勿再调用。

---

## 统计 `/statistic`

除注明外需登录 **A**；标 Admin 的仅管理员。

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/statistic/` | A | 概览计数（管理员全局 / 用户自身） |
| GET | `/user-data-diary` | Admin | 各用户日记数（`count_diary > 3`） |
| GET | `/user-data-words` | Admin | 有词条计数的用户 |
| GET | `/category` | A | 本人分类分布 |
| GET | `/year` | A | 本人年月直方图 |
| GET | `/manager-users` | Admin | Manager 用户统计：日记/码表/路书分组 + `diary_chart` |
| GET | `/weather` | A | 本人 life 气温 |

> 原 `GET /statistic/users`（diary 端用户统计）已移除，请改用 `/manager-users`。

---

## 文件 `/file-manager`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/upload` | A | multipart：`file`；form `note` |
| POST | `/modify` | A | `fileId,description` |
| DELETE | `/delete` | A | `fileId` |
| GET | `/list` | A | `pageNo,pageSize,keywords`(JSON), `dateFilter`(YYYYMM) |

文件落盘在服务端 `upload/`，文件名经 `filepath.Base` 消毒。

---

## 七牛图床 `/image-qiniu`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/image-qiniu/` | A* | Query `bucket`；可选 `hahaha` 可绕过鉴权（兼容旧客户端） |
| GET | `/list` | A | `keywords,bucket,pageNo,pageSize` |
| POST | `/add` | Admin | 上传后登记元数据 |
| DELETE | `/delete` | Admin | `id` |
| DELETE | `/batch-delete` | Admin | `ids[],bucket` |
| PUT | `/update` | Admin | `id,description` |

---

## 点赞 `/thumbs-up`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/thumbs-up/` | P | Query `key`（name）→ 计数 |
| GET | `/all` | P | 全部 |
| GET | `/list` | P | 列表 |
| POST | `/add` | Admin | `name,count,description,link_address` |
| PUT | `/modify` | Admin | 按 `name` 更新 |
| DELETE | `/delete` | Admin | `name` |

实时点赞还可走 WebSocket：`ws://host:9999/`（生产经 Nginx `/ws`）。

---

## 安装 Setup / Init

防重装：以 **`diary.users` 是否存在** 为准；`DATABASE_LOCK` 仅兼容标记。生产建议 `ALLOW_SETUP=0`。

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/setup/status` | P | `isInitialized`、`initializedByTables`、`initializedByLock`、`allowSetup`、`allowSetupEnv`、`hasRegisteredUsers` |
| POST | `/setup/config` | P | Body `databaseConfig:{host,user,password,port,…}`；仅 `allowSetup` 时可写 |
| POST | `/setup/init` | P | 执行 `init.sql`；已初始化或 `ALLOW_SETUP=0` 时拒绝 |
| GET | `/init` 或 `/init/` | P | 同上，返回 HTML 文案 |

---

## 健康检查

```http
GET /portal/
```

```json
{
  "status": "success",
  "message": "Portal API is running",
  "title": "Portal for Diary"
}
```

---

## 调用示例

### 登录

```http
POST /portal/user/login
Content-Type: application/json

{ "email": "you@example.com", "password": "secret" }
```

### 带鉴权拉日记列表

```http
GET /portal/diary/list?pageNo=1&pageSize=20
Diary-Token: $2a$10$....
Diary-Uid: 3
```

### 保存用户配置

```http
PUT /portal/user-config
Diary-Token: …
Diary-Uid: 3
Content-Type: application/json

{
  "theme": "dark",
  "default_diary_category": "life",
  "editor_mode": "markdown"
}
```

---

## 相关命令（非 HTTP）

| 命令 | 说明 |
|------|------|
| `./bin/portal` | HTTP API，默认 `:3000` |
| `./bin/ws` | WebSocket 点赞 `:9999` |
| `./bin/cron` | 刷新用户计数等 |
| `./bin/portal migrate` | 应用 `migrations/NNN_*.sql` |
