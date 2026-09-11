<template>
    <Container>
        <Toolbar :is-show-search-bar="false"/>
        <Content padding="0" :is-transparent="true">
            <div class="api-doc">
                <aside class="api-nav">
                    <div class="nav-head">
                        <h2>API 文档</h2>
                        <p>前缀 <code>/portal</code> · 响应 <code>{success,message,data}</code></p>
                    </div>
                    <button
                        v-for="m in filteredModules"
                        :key="m.id"
                        type="button"
                        class="nav-item"
                        :class="{ active: activeModule === m.id }"
                        @click="scrollToModule(m.id)"
                    >
                        <span class="nav-name">{{ m.name }}</span>
                        <span class="nav-count">{{ m.endpoints.length }}</span>
                    </button>
                </aside>

                <main class="api-main">
                    <section class="toolbar-card">
                        <ElInput
                            v-model="keyword"
                            clearable
                            placeholder="搜索路径、说明、模块…"
                            class="search"
                        >
                            <template #prefix>
                                <ElIcon><Search/></ElIcon>
                            </template>
                        </ElInput>
                        <div class="filters">
                            <button
                                v-for="m in methodOptions"
                                :key="m || 'all'"
                                type="button"
                                class="chip"
                                :class="{ on: methodFilter === m }"
                                @click="methodFilter = m"
                            >{{ m || '全部方法' }}</button>
                            <button
                                v-for="a in authOptions"
                                :key="a || 'all-auth'"
                                type="button"
                                class="chip auth"
                                :class="{ on: authFilter === a }"
                                @click="authFilter = a"
                            >{{ a ? AUTH_LABEL[a] : '全部鉴权' }}</button>
                        </div>
                        <div class="legend">
                            <span v-for="(label, key) in AUTH_LABEL" :key="key" class="legend-item">
                                <i :class="['auth-dot', `auth-${key}`]"/>{{ label }}
                                <small>{{ AUTH_TIP[key] }}</small>
                            </span>
                        </div>
                        <div class="auth-box">
                            <strong>鉴权 Header</strong>
                            <code>Authorization: Bearer &lt;jwt&gt;</code>（登录返回的 token）
                            <ElButton size="small" text type="primary" @click="copyText('Authorization: Bearer <jwt>')">复制</ElButton>
                            <div class="auth-renew-tip">剩余不足 7 天时响应头会带 <code>X-Access-Token</code> 自动续签</div>
                        </div>
                    </section>

                    <p class="result-meta">共 {{ totalEndpoints }} 个接口<span v-if="keyword || methodFilter || authFilter">（已过滤）</span></p>

                    <section
                        v-for="mod in filteredModules"
                        :id="`mod-${mod.id}`"
                        :key="mod.id"
                        class="module-block"
                    >
                        <header class="module-head">
                            <div>
                                <h3>{{ mod.name }}</h3>
                                <p>{{ mod.description }}</p>
                            </div>
                            <code class="prefix">{{ mod.prefix }}</code>
                        </header>

                        <article
                            v-for="(ep, idx) in mod.endpoints"
                            :key="mod.id + ep.method + ep.path + idx"
                            class="ep-card"
                            :class="{ open: isOpen(mod.id, idx) }"
                        >
                            <button type="button" class="ep-row" @click="toggle(mod.id, idx)">
                                <span :class="['method', ep.method.toLowerCase()]">{{ ep.method }}</span>
                                <span class="path">{{ ep.path }}</span>
                                <span :class="['auth-tag', `auth-${ep.auth}`]">{{ AUTH_LABEL[ep.auth] }}</span>
                                <span class="summary">{{ ep.summary }}</span>
                                <ElIcon class="chevron"><ArrowRight/></ElIcon>
                            </button>
                            <div v-if="isOpen(mod.id, idx)" class="ep-detail">
                                <div class="detail-actions">
                                    <ElButton size="small" @click.stop="copyText(fullPath(ep.path))">
                                        <ElIcon><CopyDocument/></ElIcon>复制路径
                                    </ElButton>
                                    <ElButton size="small" @click.stop="copyText(curlOf(ep))">
                                        <ElIcon><CopyDocument/></ElIcon>复制 curl
                                    </ElButton>
                                </div>
                                <dl>
                                    <div v-if="ep.params">
                                        <dt>Query</dt>
                                        <dd><code>{{ ep.params }}</code></dd>
                                    </div>
                                    <div v-if="ep.body">
                                        <dt>Body</dt>
                                        <dd><code>{{ ep.body }}</code></dd>
                                    </div>
                                    <div>
                                        <dt>鉴权</dt>
                                        <dd>{{ AUTH_TIP[ep.auth] }}</dd>
                                    </div>
                                    <div v-if="ep.notes">
                                        <dt>备注</dt>
                                        <dd>{{ ep.notes }}</dd>
                                    </div>
                                </dl>
                                <pre class="curl">{{ curlOf(ep) }}</pre>
                            </div>
                        </article>
                    </section>

                    <p v-if="!filteredModules.length" class="empty">没有匹配的接口</p>
                </main>
            </div>
        </Content>
    </Container>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'
import {
    API_BASE,
    API_MODULES,
    AUTH_LABEL,
    AUTH_TIP,
    type ApiEndpoint,
    type AuthLevel,
    type HttpMethod,
} from '@/view/System/apiCatalog'

const keyword = ref('')
const methodFilter = ref<HttpMethod | ''>('')
const authFilter = ref<AuthLevel | ''>('')
const activeModule = ref(API_MODULES[0]?.id || '')
const openKey = ref('')

const methodOptions = ['', 'GET', 'POST', 'PUT', 'DELETE'] as const
const authOptions = ['', 'P', 'A', 'Admin', 'Cond', 'Opt'] as const

const filteredModules = computed(() => {
    const q = keyword.value.trim().toLowerCase()
    return API_MODULES
        .map(mod => {
            const endpoints = mod.endpoints.filter(ep => {
                if (methodFilter.value && ep.method !== methodFilter.value) return false
                if (authFilter.value && ep.auth !== authFilter.value) return false
                if (!q) return true
                const hay = `${mod.name} ${mod.prefix} ${ep.path} ${ep.summary} ${ep.params || ''} ${ep.body || ''} ${ep.notes || ''}`.toLowerCase()
                return hay.includes(q)
            })
            return { ...mod, endpoints }
        })
        .filter(mod => mod.endpoints.length > 0)
})

const totalEndpoints = computed(() =>
    filteredModules.value.reduce((n, m) => n + m.endpoints.length, 0),
)

function fullPath(path: string) {
    return `${API_BASE}${path}`
}

function keyOf(modId: string, idx: number) {
    return `${modId}:${idx}`
}

function isOpen(modId: string, idx: number) {
    return openKey.value === keyOf(modId, idx)
}

function toggle(modId: string, idx: number) {
    const k = keyOf(modId, idx)
    openKey.value = openKey.value === k ? '' : k
}

function scrollToModule(id: string) {
    activeModule.value = id
    document.getElementById(`mod-${id}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function curlOf(ep: ApiEndpoint) {
    const url = `https://kylebing.cn${fullPath(ep.path)}`
    const lines = [`curl -X ${ep.method} '${url}'`]
    if (ep.auth === 'A' || ep.auth === 'Admin' || ep.auth === 'Cond' || ep.auth === 'Opt') {
        lines.push(`  -H 'Authorization: Bearer <jwt>'`)
    }
    if (ep.method !== 'GET') {
        lines.push(`  -H 'Content-Type: application/json'`)
        lines.push(`  -d '{}'`)
    }
    return lines.join(' \\\n')
}

async function copyText(text: string) {
    try {
        await navigator.clipboard.writeText(text)
        ElMessage.success('已复制')
    } catch {
        ElMessage.error('复制失败')
    }
}
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;

.api-doc {
    display: grid;
    grid-template-columns: 220px minmax(0, 1fr);
    min-height: calc(100% - 0px);
    background: #f4f5f7;
}

.api-nav {
    position: sticky;
    top: 0;
    align-self: start;
    max-height: calc(100vh - 100px);
    overflow: auto;
    padding: 16px 12px;
    background: #fff;
    border-right: 1px solid #e8eaed;
}

.nav-head {
    padding: 4px 8px 14px;
    h2 {
        margin: 0 0 6px;
        font-size: 16px;
        color: #1f2329;
    }
    p {
        margin: 0;
        font-size: 12px;
        color: #8a919f;
        line-height: 1.5;
    }
    code {
        font-size: 11px;
        background: #f0f2f5;
        padding: 1px 4px;
        border-radius: 3px;
    }
}

.nav-item {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    border: 0;
    background: transparent;
    text-align: left;
    padding: 8px 10px;
    border-radius: 8px;
    cursor: pointer;
    color: #4e5969;
    margin-bottom: 2px;
    &:hover { background: #f5f6f8; }
    &.active {
        background: #eef5ff;
        color: #1677ff;
        font-weight: 600;
    }
}

.nav-count {
    font-size: 11px;
    color: #86909c;
    background: #f2f3f5;
    border-radius: 999px;
    padding: 0 7px;
    line-height: 18px;
}

.api-main {
    padding: 16px 20px 40px;
    overflow: auto;
}

.toolbar-card {
    background: #fff;
    border-radius: 12px;
    padding: 16px;
    border: 1px solid #e8eaed;
    margin-bottom: 14px;
}

.search {
    max-width: 420px;
    margin-bottom: 12px;
}

.filters {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
}

.chip {
    border: 1px solid #e5e6eb;
    background: #fff;
    color: #4e5969;
    border-radius: 999px;
    padding: 4px 12px;
    font-size: 12px;
    cursor: pointer;
    &.on {
        border-color: #1677ff;
        color: #1677ff;
        background: #eef5ff;
    }
}

.legend {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 16px;
    margin-bottom: 10px;
}

.legend-item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: #4e5969;
    small {
        color: #86909c;
    }
}

.auth-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    display: inline-block;
    flex-shrink: 0;
    &.auth-P { background: #389e0d; }
    &.auth-A { background: #1677ff; }
    &.auth-Admin { background: #cf1322; }
    &.auth-Cond { background: #d46b08; }
    &.auth-Opt { background: #722ed1; }
    &.auth-- { background: #8c8c8c; }
}

.auth-box {
    font-size: 13px;
    color: #4e5969;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    code {
        background: #f0f2f5;
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 12px;
    }
    .auth-renew-tip {
        flex: 1 0 100%;
        font-size: 12px;
        color: #86909c;
    }
}

.result-meta {
    margin: 0 0 12px;
    font-size: 13px;
    color: #86909c;
}

.module-block {
    background: #fff;
    border: 1px solid #e8eaed;
    border-radius: 12px;
    margin-bottom: 14px;
    overflow: hidden;
}

.module-head {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: flex-start;
    padding: 14px 16px;
    border-bottom: 1px solid #f0f2f5;
    h3 {
        margin: 0 0 4px;
        font-size: 15px;
        color: #1f2329;
    }
    p {
        margin: 0;
        font-size: 12px;
        color: #86909c;
    }
    .prefix {
        font-size: 12px;
        color: #1677ff;
        background: #eef5ff;
        padding: 4px 8px;
        border-radius: 6px;
        white-space: nowrap;
    }
}

.ep-card {
    border-bottom: 1px solid #f5f6f8;
    &:last-child { border-bottom: 0; }
    &.open { background: #fafbfc; }
}

.ep-row {
    width: 100%;
    border: 0;
    background: transparent;
    display: grid;
    grid-template-columns: 64px minmax(160px, 1.2fr) 72px minmax(0, 1.6fr) 20px;
    gap: 10px;
    align-items: center;
    padding: 10px 16px;
    cursor: pointer;
    text-align: left;
    &:hover { background: #f7f8fa; }
}

.method {
    font-size: 11px;
    font-weight: 700;
    text-align: center;
    border-radius: 4px;
    padding: 3px 0;
    letter-spacing: 0.02em;
    &.get { color: #08979c; background: #e6fffb; }
    &.post { color: #389e0d; background: #f6ffed; }
    &.put { color: #d46b08; background: #fff7e6; }
    &.delete { color: #cf1322; background: #fff1f0; }
}

.path {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    color: #1f2329;
    word-break: break-all;
}

.auth-tag {
    font-size: 11px;
    text-align: center;
    border-radius: 999px;
    padding: 2px 0;
}

.auth-tag.auth-P { background: #f6ffed; color: #389e0d; }
.auth-tag.auth-A { background: #e6f4ff; color: #1677ff; }
.auth-tag.auth-Admin { background: #fff1f0; color: #cf1322; }
.auth-tag.auth-Cond { background: #fff7e6; color: #d46b08; }
.auth-tag.auth-Opt { background: #f9f0ff; color: #722ed1; }
.auth-tag.auth-- { background: #f5f5f5; color: #8c8c8c; }

.summary {
    font-size: 13px;
    color: #4e5969;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.chevron {
    color: #c0c4cc;
    transition: transform .15s ease;
}
.ep-card.open .chevron {
    transform: rotate(90deg);
}

.ep-detail {
    padding: 0 16px 14px 16px;
}

.detail-actions {
    display: flex;
    gap: 8px;
    margin-bottom: 10px;
}

dl {
    margin: 0 0 10px;
    display: grid;
    gap: 8px;
    div {
        display: grid;
        grid-template-columns: 56px 1fr;
        gap: 8px;
        font-size: 13px;
    }
    dt {
        color: #86909c;
    }
    dd {
        margin: 0;
        color: #1f2329;
        code {
            font-size: 12px;
            background: #f0f2f5;
            padding: 2px 6px;
            border-radius: 4px;
            word-break: break-all;
        }
    }
}

.curl {
    margin: 0;
    background: #1f2329;
    color: #e8eaed;
    border-radius: 8px;
    padding: 12px 14px;
    font-size: 12px;
    line-height: 1.55;
    overflow: auto;
}

.empty {
    text-align: center;
    color: #86909c;
    padding: 40px 0;
}

@media (max-width: 960px) {
    .api-doc {
        grid-template-columns: 1fr;
    }
    .api-nav {
        position: static;
        max-height: none;
        display: flex;
        flex-wrap: wrap;
        gap: 4px;
        border-right: 0;
        border-bottom: 1px solid #e8eaed;
    }
    .nav-head { width: 100%; }
    .nav-item { width: auto; }
    .ep-row {
        grid-template-columns: 56px 1fr 20px;
        grid-template-areas:
            "method path chev"
            "auth summary chev";
        .method { grid-area: method; }
        .path { grid-area: path; }
        .auth-tag { grid-area: auth; }
        .summary { grid-area: summary; white-space: normal; }
        .chevron { grid-area: chev; }
    }
}
</style>
