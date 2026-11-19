# 更新记录

## v1.5.1 `2026-11-19`

- 修复和风响应 gzip 未解压导致 JSON 解析失败（`\x1f`）
- Manager：**系统配置 → 天气测试**（status / geo / now / daily）

## v1.5.0 `2026-11-19`

- **天气代理** `/portal/weather/*`：服务端用和风 JWT（`QWEATHER_*` 环境变量）或回退系统配置 API Key，代理 `now` / `daily` / `geo`；NOTE4C 等设备只带 Portal 登录 Token
- 文档与 Manager API 目录同步

## v1.4.0 `2026-11-03`

- **注册申请**：开放 `registration-application` 提交与管理员审核台账；用户获取邀请码需先支付当前配置金额（默认 ¥28），审核通过后邮件下发邀请码
- **系统配置**：新增注册收款档位 `registration_pay_amount_yuan`（¥6/8/18/28/38/48 或自定义无固定码）；管理后台增加系统配置页与收款码预览
- **邀请码**：记录申请邮箱与支付昵称；用户列表展示注册方式（付费/赠送/通用码等）

## v1.3.0 `2026-10-22`

- 五笔模块已拆分到 `portal-wubi-go`，其对外前缀改为 `/wubi/portal/*`，管理后台在 `/wubi/manager/`
- README 的模块说明与 nginx 示例同步到新路径；`/wubi/` 根路径由五笔说明站占用
- 版本号 1.3.0
