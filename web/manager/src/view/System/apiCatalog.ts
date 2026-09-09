export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'
export type AuthLevel = 'P' | 'A' | 'Admin' | 'Cond' | 'Opt' | '-'

export interface ApiEndpoint {
    method: HttpMethod
    path: string
    auth: AuthLevel
    summary: string
    params?: string
    body?: string
    notes?: string
}

export interface ApiModule {
    id: string
    name: string
    prefix: string
    description: string
    endpoints: ApiEndpoint[]
}

export const AUTH_LABEL: Record<AuthLevel, string> = {
    P: '公开',
    A: '登录',
    Admin: '管理员',
    Cond: '条件',
    Opt: '可选登录',
    '-': '—',
}

export const AUTH_TIP: Record<AuthLevel, string> = {
    P: '无需鉴权',
    A: '需要 Diary-Token + Diary-Uid',
    Admin: '需要管理员（group_id = 1）',
    Cond: '公开内容可匿名，否则需主人',
    Opt: '有 token 则带身份，否则仅公开数据',
    '-': '特殊/废弃',
}

export const API_BASE = '/portal'

export const API_MODULES: ApiModule[] = [
    {
        id: 'user',
        name: '用户',
        prefix: '/user',
        description: '注册登录、资料与账号管理',
        endpoints: [
            { method: 'POST', path: '/user/register', auth: 'P', summary: '注册（已有用户时需邀请码；首个用户为管理员）', body: 'email, password, username, nickname, invitationCode, …' },
            { method: 'POST', path: '/user/login', auth: 'P', summary: '登录；返回用户信息，password 字段即 Diary-Token', body: 'email, password' },
            { method: 'GET', path: '/user/avatar', auth: 'P', summary: '头像字段', params: 'email' },
            { method: 'GET', path: '/user/detail', auth: 'Cond', summary: '按二维码 hash 查（历史兼容）', params: 'hash' },
            { method: 'POST', path: '/user/list', auth: 'A', summary: '用户列表；管理员全部，否则仅自己', body: 'pageNo, pageSize' },
            { method: 'POST', path: '/user/add', auth: 'P', summary: '创建用户（当前无鉴权）', body: '同注册 + group_id', notes: '注意：接口公开' },
            { method: 'PUT', path: '/user/set-profile', auth: 'A', summary: '更新个人资料', body: 'nickname, phone, avatar, city, geolocation' },
            { method: 'PUT', path: '/user/modify', auth: 'A', summary: '改资料；本人或管理员', body: 'uid, email, nickname, …' },
            { method: 'DELETE', path: '/user/delete', auth: 'Admin', summary: '删除用户', body: 'uid' },
            { method: 'PUT', path: '/user/change-password', auth: 'A', summary: '改密后 token 会变化', body: 'password' },
            { method: 'DELETE', path: '/user/destroy-account', auth: 'A', summary: '注销并清理关联数据' },
        ],
    },
    {
        id: 'user-config',
        name: '用户配置',
        prefix: '/user-config',
        description: '主题、默认分类、编辑器偏好',
        endpoints: [
            { method: 'GET', path: '/user-config', auth: 'A', summary: '读取配置（有/无尾斜杠均可）' },
            { method: 'PUT', path: '/user-config', auth: 'A', summary: '部分更新', body: 'theme, default_diary_category, editor_mode, config_json' },
        ],
    },
    {
        id: 'system-config',
        name: '系统配置',
        prefix: '/system-config',
        description: '演示账号、七牛、和风天气等',
        endpoints: [
            { method: 'GET', path: '/system-config', auth: 'P', summary: '公开配置（不含 invitation_code / 七牛密钥）' },
            { method: 'GET', path: '/system-config/admin', auth: 'Admin', summary: '完整配置含密钥' },
            { method: 'PUT', path: '/system-config', auth: 'Admin', summary: '保存系统配置' },
        ],
    },
    {
        id: 'invitation',
        name: '邀请码',
        prefix: '/invitation',
        description: '注册邀请码',
        endpoints: [
            { method: 'GET', path: '/invitation/list', auth: 'Opt', summary: '未绑定邀请码；非管理员仅 is_shared=0' },
            { method: 'GET', path: '/invitation/manage', auth: 'Admin', summary: 'Manager 管理列表（含已使用与绑定用户）', params: 'status=all|unused|used, pageNo, pageSize' },
            { method: 'POST', path: '/invitation/generate', auth: 'Admin', summary: '生成邀请码' },
            { method: 'POST', path: '/invitation/mark-shared', auth: 'Admin', summary: '标记已分享', body: 'id' },
            { method: 'DELETE', path: '/invitation/delete', auth: 'Admin', summary: '删除', params: 'id' },
        ],
    },
    {
        id: 'diary',
        name: '日记',
        prefix: '/diary',
        description: '核心日记 CRUD；emoji 以 \\uXXXX 存 utf8mb3',
        endpoints: [
            { method: 'GET', path: '/diary/list', auth: 'A', summary: '分页列表', params: 'pageNo, pageSize, keywords, categories, filterShared, timeStart, timeEnd' },
            { method: 'GET', path: '/diary/list-all', auth: 'A', summary: '全量匹配', params: '同上过滤项' },
            { method: 'GET', path: '/diary/list-title-only', auth: 'A', summary: '仅 id/date/title/category' },
            { method: 'GET', path: '/diary/list-category-only', auth: 'A', summary: '仅 id/date/category' },
            { method: 'GET', path: '/diary/export', auth: 'A', summary: '导出' },
            { method: 'GET', path: '/diary/temperature', auth: 'A', summary: 'life 类气温（最多 100）', params: 'timeStart, timeEnd' },
            { method: 'GET', path: '/diary/detail', auth: 'Cond', summary: '详情；公开或主人', params: 'diaryId' },
            { method: 'POST', path: '/diary/add', auth: 'A', summary: '新增', body: 'title, content, category, weather, temperature, temperature_outside, date, is_public, is_markdown' },
            { method: 'PUT', path: '/diary/modify', auth: 'A', summary: '修改本人日记', body: 'id + 同上字段' },
            { method: 'DELETE', path: '/diary/delete', auth: 'A', summary: '删除', body: 'diaryId' },
            { method: 'GET', path: '/diary/latest-recommend', auth: 'P', summary: '首页推荐（公开，uid=3）' },
            { method: 'GET', path: '/diary/get-diary-content-with-keyword', auth: 'A', summary: '本人最新匹配标题', params: 'keyword' },
            { method: 'GET', path: '/diary/get-latest-public-diary-with-keyword', auth: 'P', summary: '最新公开匹配', params: 'keyword' },
            { method: 'POST', path: '/diary/clear', auth: 'A', summary: '清空本人日记（演示账号受限）' },
        ],
    },
    {
        id: 'diary-category',
        name: '日记分类',
        prefix: '/diary-category',
        description: '分类字典',
        endpoints: [
            { method: 'GET', path: '/diary-category/list', auth: 'P', summary: '全部分类' },
            { method: 'POST', path: '/diary-category/add', auth: 'Admin', summary: '新增', body: 'name, name_en, color, sort_id' },
            { method: 'PUT', path: '/diary-category/modify', auth: 'Admin', summary: '按 name_en 更新' },
            { method: 'DELETE', path: '/diary-category/delete', auth: 'Admin', summary: '删除', body: 'name_en' },
        ],
    },
    {
        id: 'bill',
        name: '账单',
        prefix: '/bill',
        description: '解析 category=bill 日记正文',
        endpoints: [
            { method: 'GET', path: '/bill/', auth: 'A', summary: '全部账单日结构化数据' },
            { method: 'GET', path: '/bill/sorted', auth: 'A', summary: '按年整理', params: 'years, keyword' },
            { method: 'GET', path: '/bill/keys', auth: 'A', summary: '近 5 年条目词频' },
            { method: 'GET', path: '/bill/day-sum', auth: 'A', summary: '日收支合计' },
            { method: 'GET', path: '/bill/month-sum', auth: 'A', summary: '月合计', params: 'keyword?' },
            { method: 'GET', path: '/bill/borrow', auth: 'A', summary: '借还记录日记内容' },
        ],
    },
    {
        id: 'bank-card',
        name: '银行卡',
        prefix: '/bank-card',
        description: '固定标题日记内容',
        endpoints: [
            { method: 'GET', path: '/bank-card', auth: 'A', summary: '「我的银行卡列表」日记 content' },
        ],
    },
    {
        id: 'qr',
        name: '二维码',
        prefix: '/qr-front · /qr-manager',
        description: '车贴二维码前台与管理',
        endpoints: [
            { method: 'GET', path: '/qr-front/', auth: 'P', summary: '公开详情并增加 visit_count', params: 'hash' },
            { method: 'GET', path: '/qr-manager/list', auth: 'A', summary: '列表', params: 'keywords, pageNo, pageSize' },
            { method: 'GET', path: '/qr-manager/detail', auth: 'Cond', summary: '详情', params: 'hash' },
            { method: 'POST', path: '/qr-manager/add', auth: 'A', summary: '新增', body: 'hash, message, description, is_public, …' },
            { method: 'PUT', path: '/qr-manager/modify', auth: 'A', summary: '修改' },
            { method: 'DELETE', path: '/qr-manager/delete', auth: 'A', summary: '删除', body: 'hash' },
            { method: 'POST', path: '/qr-manager/clear-visit-count', auth: 'Admin', summary: '清访问计数', body: 'hash' },
        ],
    },
    {
        id: 'map-route',
        name: '地图路线',
        prefix: '/map-route',
        description: '路书路线',
        endpoints: [
            { method: 'POST', path: '/map-route/list', auth: 'Opt', summary: '分页列表；无身份仅公开', body: 'isMine, keyword, dateRange, pageNo, pageSize' },
            { method: 'GET', path: '/map-route/detail', auth: 'Cond', summary: '详情', params: 'id' },
            { method: 'POST', path: '/map-route/add', auth: 'A', summary: '新增', body: 'name, area, road_type, policy, seasons, video_link, paths, note, thumb_up' },
            { method: 'PUT', path: '/map-route/modify', auth: 'A', summary: '修改（主人/管理员）', body: 'id + 字段' },
            { method: 'DELETE', path: '/map-route/delete', auth: 'A', summary: '删除', body: 'id' },
        ],
    },
    {
        id: 'map-pointer',
        name: '地图标点',
        prefix: '/map-pointer',
        description: '地图兴趣点',
        endpoints: [
            { method: 'POST', path: '/map-pointer/list', auth: 'Opt', summary: '分页列表', body: 'keyword, dateRange, pageNo, pageSize' },
            { method: 'GET', path: '/map-pointer/detail', auth: 'Cond', summary: '详情', params: 'id' },
            { method: 'POST', path: '/map-pointer/add', auth: 'A', summary: '新增', body: 'name, pointers, note, area, thumb_up, is_public' },
            { method: 'PUT', path: '/map-pointer/modify', auth: 'A', summary: '修改', body: 'id + 字段' },
            { method: 'DELETE', path: '/map-pointer/delete', auth: 'A', summary: '删除', body: 'id' },
        ],
    },
    {
        id: 'wubi',
        name: '五笔',
        prefix: '/wubi',
        description: '词库备份、词条、分类',
        endpoints: [
            { method: 'GET', path: '/wubi/dict/pull', auth: 'A', summary: '拉取备份', params: 'title' },
            { method: 'PUT', path: '/wubi/dict/push', auth: 'A', summary: '推送备份', body: 'title, content, contentSize, wordCount' },
            { method: 'GET', path: '/wubi/dict/check-backup-exist', auth: 'A', summary: '检查备份是否存在（亦支持 POST）', params: 'fileName' },
            { method: 'POST', path: '/wubi/word/list', auth: 'A', summary: '词条列表', body: 'keyword, dateRange, category_id, approved, pageNo, pageSize' },
            { method: 'POST', path: '/wubi/word/export-extra', auth: 'A', summary: '导出已审核非默认分类词' },
            { method: 'POST', path: '/wubi/word/check-exist', auth: 'A', summary: '模糊查重', body: 'word, code' },
            { method: 'POST', path: '/wubi/word/add', auth: 'A', summary: '新增词条', body: 'word, code, priority, up, down, comment, category_id' },
            { method: 'POST', path: '/wubi/word/add-batch', auth: 'A', summary: '批量新增', body: 'words[], category_id' },
            { method: 'PUT', path: '/wubi/word/modify', auth: 'A', summary: '修改；非管理员仅本人', body: 'id + 字段' },
            { method: 'DELETE', path: '/wubi/word/delete', auth: 'A', summary: '批量删除', body: 'ids[]' },
            { method: 'PUT', path: '/wubi/word/modify-batch', auth: 'A', summary: '批量更新', body: 'ids[], category_id?, approved?' },
            { method: 'GET', path: '/wubi/category/list', auth: 'A', summary: '分类及词数' },
            { method: 'POST', path: '/wubi/category/add', auth: 'Admin', summary: '新增分类', body: 'name, sort_id' },
            { method: 'PUT', path: '/wubi/category/modify', auth: 'Admin', summary: '修改分类', body: 'id, name, sort_id' },
            { method: 'DELETE', path: '/wubi/category/delete', auth: 'Admin', summary: '删除分类', body: 'id' },
        ],
    },
    {
        id: 'starve-new',
        name: '饥荒',
        prefix: '/starve-new',
        description: 'starve_advance 资料库；经典 /starve 已下线',
        endpoints: [
            { method: 'GET', path: '/starve-new/{entity}/list', auth: 'P', summary: '列表', params: 'keyword, version, tab, all, pageNo, pageSize', notes: 'entity: character/mob/log/plant/thing/material/craft/cookingrecipe/coder/version/craft-tab' },
            { method: 'GET', path: '/starve-new/{entity}/info', auth: 'P', summary: '详情', params: 'id' },
            { method: 'POST', path: '/starve-new/{entity}/add', auth: 'A', summary: '新增', body: '实体字段；部分需 version/tab id 列表' },
            { method: 'PUT', path: '/starve-new/{entity}/modify', auth: 'A', summary: '修改', body: 'id + 字段' },
            { method: 'DELETE', path: '/starve-new/{entity}/delete', auth: 'A', summary: '删除', body: 'id（或 query）' },
        ],
    },
    {
        id: 'statistic',
        name: '统计',
        prefix: '/statistic',
        description: '概览与图表；用户明细仅管理员',
        endpoints: [
            { method: 'GET', path: '/statistic/', auth: 'A', summary: '概览计数' },
            { method: 'GET', path: '/statistic/user-data-diary', auth: 'Admin', summary: '各用户日记数（>3）' },
            { method: 'GET', path: '/statistic/user-data-words', auth: 'Admin', summary: '各用户词条数' },
            { method: 'GET', path: '/statistic/category', auth: 'A', summary: '本人分类分布' },
            { method: 'GET', path: '/statistic/year', auth: 'A', summary: '本人年月直方图' },
            { method: 'GET', path: '/statistic/manager-users', auth: 'Admin', summary: 'Manager 用户统计分组 + diary_chart', notes: '原 /statistic/users 已移除' },
            { method: 'GET', path: '/statistic/weather', auth: 'A', summary: '本人 life 气温' },
        ],
    },
    {
        id: 'file-manager',
        name: '文件',
        prefix: '/file-manager',
        description: '本地上传 upload/{uid}/。WebRTC 互传信令仍由 portal-ws（rtc-*）提供，前后端均无互传页',
        endpoints: [
            { method: 'POST', path: '/file-manager/upload', auth: 'A', summary: '上传到 upload/{uid}/', body: 'multipart file + note' },
            { method: 'POST', path: '/file-manager/modify', auth: 'A', summary: '改描述', body: 'fileId, description' },
            { method: 'DELETE', path: '/file-manager/delete', auth: 'A', summary: '删除', body: 'fileId' },
            { method: 'GET', path: '/file-manager/list', auth: 'A', summary: '列表（含 download_url）', params: 'pageNo, pageSize, keywords, dateFilter' },
            { method: 'GET', path: '/file-manager/download', auth: 'A', summary: '鉴权下载', params: 'fileId' },
        ],
    },
    {
        id: 'image-qiniu',
        name: '七牛图床',
        prefix: '/image-qiniu',
        description: '上传 Token 与图片元数据',
        endpoints: [
            { method: 'GET', path: '/image-qiniu/', auth: 'A', summary: '上传 Token', params: 'bucket；可选 hahaha 绕过鉴权', notes: '兼容旧客户端' },
            { method: 'GET', path: '/image-qiniu/list', auth: 'A', summary: '列表', params: 'keywords, bucket, pageNo, pageSize' },
            { method: 'POST', path: '/image-qiniu/add', auth: 'Admin', summary: '登记元数据' },
            { method: 'DELETE', path: '/image-qiniu/delete', auth: 'Admin', summary: '删除', body: 'id' },
            { method: 'DELETE', path: '/image-qiniu/batch-delete', auth: 'Admin', summary: '批量删除', body: 'ids[], bucket' },
            { method: 'PUT', path: '/image-qiniu/update', auth: 'Admin', summary: '更新描述', body: 'id, description' },
        ],
    },
    {
        id: 'thumbs-up',
        name: '点赞',
        prefix: '/thumbs-up',
        description: '站点点赞计数；实时还可走 WebSocket :9999',
        endpoints: [
            { method: 'GET', path: '/thumbs-up/', auth: 'P', summary: '按 name 计数', params: 'key' },
            { method: 'GET', path: '/thumbs-up/all', auth: 'P', summary: '全部' },
            { method: 'GET', path: '/thumbs-up/list', auth: 'P', summary: '列表' },
            { method: 'POST', path: '/thumbs-up/add', auth: 'Admin', summary: '新增', body: 'name, count, description, link_address' },
            { method: 'PUT', path: '/thumbs-up/modify', auth: 'Admin', summary: '按 name 更新' },
            { method: 'DELETE', path: '/thumbs-up/delete', auth: 'Admin', summary: '删除', body: 'name' },
        ],
    },
    {
        id: 'setup',
        name: '安装引导',
        prefix: '/setup · /init',
        description: '以防 diary.users 为准；生产 ALLOW_SETUP=0',
        endpoints: [
            { method: 'GET', path: '/setup/status', auth: 'P', summary: '初始化状态与 allowSetup' },
            { method: 'POST', path: '/setup/config', auth: 'P', summary: '写库配置（仅 allowSetup）', body: 'databaseConfig:{host,user,password,port,…}' },
            { method: 'POST', path: '/setup/init', auth: 'P', summary: '执行 init.sql（已初始化或 ALLOW_SETUP=0 拒绝）' },
            { method: 'GET', path: '/init', auth: 'P', summary: '同上，返回 HTML' },
        ],
    },
]
