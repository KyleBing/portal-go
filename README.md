# Portal Go

Go 重写的 Portal API，并统一托管原 `manager-src` 管理后台。

## 功能

- 鉴权：`Authorization: Bearer <jwt>`（登录返回 `data.token`）；响应体 `{success,message,data}`
- 接口说明：[docs/api.md](docs/api.md)
- 模块：用户/用户配置/邀请码、日记/账单/银行卡、地图、二维码、五笔、文件、七牛、点赞、统计、饥荒 API、安装引导/系统配置
- 管理后台：`/manager/`（Vue3 + Vite + Element Plus）
- WebSocket 点赞：`cmd/ws` 端口 `9999`
- 用户计数刷新：`cmd/cron`
- 增量库表迁移：`portal migrate`（嵌入 `migrations/NNN_*.sql`）

## 目录

```
cmd/portal|ws|cron
internal/...
web/manager/          # 管理后台源码
config/configDatabase.json
migrations/init.sql   # 首次安装
migrations/001_*.sql  # 增量迁移（嵌入二进制）
deploy.sh             # 一键构建 + 部署 + migrate
```

## 本地运行

```bash
go mod tidy
export JWT_SECRET='change-me-to-a-long-random-string'
vim config/configDatabase.json
make backend
./bin/portal migrate   # 应用增量迁移
./bin/portal           # :3000
./bin/ws               # :9999
```

管理后台开发：

```bash
cd web/manager && yarn && yarn dev   # :4000，代理 /portal -> :3000
```

## 生产部署

本地一条命令完成交叉编译、上传、migrate、重启（原 `build.sh` 已改名为 `deploy.sh`）：

```bash
cp deploy.env.example deploy.env   # 按需改主机/路径
./deploy.sh                        # linux/amd64 + 部署
./deploy.sh --no-deploy            # 只编译
./deploy.sh --no-frontend          # 不构建 manager
```

`deploy.sh` 会：

1. 可选构建 `web/manager`
2. 交叉编译 `portal` / `ws` / `cron`
3. rsync 到 `DEPLOY_PATH`
4. 远程执行 `./bin/portal migrate`
5. `systemctl restart portal-go portal-ws`

新增表结构时：在 `migrations/` 增加 `NNN_description.sql`（勿改已发布文件），再跑 `./deploy.sh`。

## Nginx 示例

```nginx
location /portal/  { proxy_pass http://127.0.0.1:3000/portal/; }
location /manager/ { proxy_pass http://127.0.0.1:3000/manager/; }
location /ws {
  proxy_pass http://127.0.0.1:9999/;
  proxy_http_version 1.1;
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection $connection_upgrade;
}
```

## 初始化 / 防重装

是否已初始化以 **`diary.users` 表是否存在** 为准（`DATABASE_LOCK` 仅作兼容标记）。

- `GET /setup/status`：返回 `isInitialized` / `initializedByTables` / `allowSetup` 等
- `POST /setup/config`、`POST /setup/init`、`GET /init`：未初始化且允许安装时才可执行
- 生产环境请设置 **`ALLOW_SETUP=0`**（systemd `Environment=`），即使误删锁文件也无法走安装引导
- 确需重装：清空/重建库表，临时设 `ALLOW_SETUP=1`，完成后再改回 `0`

## Cron

```cron
17 * * * * cd /var/www/html/portal-go && ./bin/cron
```
