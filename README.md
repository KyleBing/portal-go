# Portal Go

Go 重写的 Portal API，并统一托管原 `manager-src` 管理后台。

## 功能

- 与原 Node Portal **API 兼容**（路径、`Diary-Token` / `Diary-Uid`、`{success,message,data}`）
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
build.sh              # 本地交叉编译（不含上传）
deploy.sh             # rsync 上传 + migrate + 重启（不含编译）
```

## 本地运行

```bash
go mod tidy
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

先本地构建，再上传（与 diary / RideTrack 一样，`deploy.sh` 不编译）：

```bash
cp deploy.env.example deploy.env   # 按需改主机/路径
./build.sh                         # 交叉编译 + 可选 yarn build manager
./build.sh --no-frontend           # 只编 Go，不同步前端时也可 DEPLOY_FRONTEND=0 ./deploy.sh
./deploy.sh                        # rsync + migrate + systemctl restart
```

`deploy.sh` 会：

1. rsync `bin/linux/{portal,ws,cron}` 与可选 `web/manager/dist`
2. 远程执行 `./bin/portal migrate`
3. `systemctl restart portal-go portal-ws`

新增表结构时：在 `migrations/` 增加 `NNN_description.sql`（勿改已发布文件），再 `./build.sh && ./deploy.sh`。

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
