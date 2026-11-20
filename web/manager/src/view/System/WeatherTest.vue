<template>
    <Container>
        <Toolbar :is-show-search-bar="false">
            <template #left>
                <ElButton type="primary" :loading="statusLoading" @click="loadStatus">刷新状态</ElButton>
            </template>
        </Toolbar>
        <Content padding="10px" v-loading="statusLoading && !status">
            <div class="panel">
                <div class="panel-head">
                    <div class="panel-title">代理状态</div>
                    <p class="panel-desc">服务端和风凭据（JWT / API Key），不含密钥明文</p>
                </div>
                <div v-if="status" class="status-row">
                    <span class="pill" :class="status.configured ? 'on' : 'off'">
                        {{ status.configured ? '已配置' : '未配置' }}
                    </span>
                    <span class="meta">认证 {{ status.auth || '—' }}</span>
                    <span class="meta">Host {{ status.host || '—' }}</span>
                </div>
                <p v-if="status?.message" class="err">{{ status.message }}</p>
            </div>

            <div class="panel">
                <div class="panel-head">
                    <div class="panel-title">城市搜索 Geo</div>
                    <p class="panel-desc">GET /portal/geo/v2/city/lookup（和风官方路径）</p>
                </div>
                <div class="row">
                    <ElInput v-model="geoQuery" clearable placeholder="城市名，例如 杭州" @keyup.enter="runGeo" />
                    <ElButton type="primary" :loading="geoLoading" @click="runGeo">搜索</ElButton>
                </div>
                <ul v-if="geoList.length" class="geo-list">
                    <li v-for="item in geoList" :key="item.id">
                        <button type="button" class="geo-pick" @click="useLocation(item)">
                            <strong>{{ item.name }}</strong>
                            <code>{{ item.id }}</code>
                            <span class="sub">{{ [item.adm1, item.adm2, item.country].filter(Boolean).join(' · ') }}</span>
                        </button>
                    </li>
                </ul>
            </div>

            <div class="panel">
                <div class="panel-head">
                    <div class="panel-title">实况 / 预报</div>
                    <p class="panel-desc">官方路径 /v7/weather/now · 24h · 7d；LocationID 或 lon,lat</p>
                </div>
                <div class="row">
                    <ElInput v-model="locationId" clearable placeholder="LocationID，例如 101210101" />
                    <ElSelect v-model="days" style="width: 110px">
                        <ElOption label="7 天" value="7d" />
                        <ElOption label="15 天" value="15d" />
                        <ElOption label="3 天" value="3d" />
                    </ElSelect>
                    <ElButton type="primary" :loading="nowLoading" @click="runNow">实况</ElButton>
                    <ElButton :loading="hourlyLoading" @click="runHourly">24 小时</ElButton>
                    <ElButton :loading="dailyLoading" @click="runDaily">预报</ElButton>
                </div>
            </div>

            <div v-if="resultTitle" class="panel">
                <div class="panel-head panel-head--row">
                    <div class="panel-title">{{ resultTitle }}</div>
                    <ElButton text type="primary" @click="copyResult">复制 JSON</ElButton>
                </div>
                <pre class="json">{{ resultText }}</pre>
            </div>
        </Content>
    </Container>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'
import weatherApi from '@/api/weatherApi'

type WeatherStatus = {
    configured?: boolean
    auth?: string
    host?: string
    message?: string
}

type GeoItem = {
    name: string
    id: string
    adm1?: string
    adm2?: string
    country?: string
}

const statusLoading = ref(false)
const status = ref<WeatherStatus | null>(null)

const geoQuery = ref('杭州')
const geoLoading = ref(false)
const geoList = ref<GeoItem[]>([])

const locationId = ref('')
const days = ref('7d')
const nowLoading = ref(false)
const hourlyLoading = ref(false)
const dailyLoading = ref(false)

const resultTitle = ref('')
const resultText = ref('')

async function loadStatus() {
    statusLoading.value = true
    try {
        const res: any = await weatherApi.status()
        status.value = (res.data || {}) as WeatherStatus
    } catch {
        status.value = null
    } finally {
        statusLoading.value = false
    }
}

async function runGeo() {
    const q = geoQuery.value.trim()
    if (!q) {
        ElMessage.warning('请输入城市名')
        return
    }
    geoLoading.value = true
    geoList.value = []
    try {
        const res: any = await weatherApi.geo(q)
        const data = res.data || {}
        const arr = Array.isArray(data.location) ? data.location : []
        geoList.value = arr.map((x: any) => ({
            name: String(x.name || ''),
            id: String(x.id || ''),
            adm1: x.adm1,
            adm2: x.adm2,
            country: x.country,
        }))
        resultTitle.value = `Geo · ${q}`
        resultText.value = JSON.stringify(data, null, 2)
        if (!geoList.value.length) ElMessage.info('无匹配城市')
    } catch {
        /* request 已弹错 */
    } finally {
        geoLoading.value = false
    }
}

function useLocation(item: GeoItem) {
    locationId.value = item.id
    ElMessage.success(`已填入 ${item.name}（${item.id}）`)
}

async function runNow() {
    const loc = locationId.value.trim()
    if (!loc) {
        ElMessage.warning('请填写 LocationID')
        return
    }
    nowLoading.value = true
    try {
        const res: any = await weatherApi.now(loc)
        resultTitle.value = `实况 · ${loc}`
        resultText.value = JSON.stringify(res.data || {}, null, 2)
    } catch {
        /* */
    } finally {
        nowLoading.value = false
    }
}

async function runHourly() {
    const loc = locationId.value.trim()
    if (!loc) {
        ElMessage.warning('请填写 LocationID')
        return
    }
    hourlyLoading.value = true
    try {
        const res: any = await weatherApi.hourly(loc, '24h')
        resultTitle.value = `逐小时 24h · ${loc}`
        resultText.value = JSON.stringify(res.data || {}, null, 2)
    } catch {
        /* */
    } finally {
        hourlyLoading.value = false
    }
}

async function runDaily() {
    const loc = locationId.value.trim()
    if (!loc) {
        ElMessage.warning('请填写 LocationID')
        return
    }
    dailyLoading.value = true
    try {
        const res: any = await weatherApi.daily(loc, days.value)
        resultTitle.value = `预报 ${days.value} · ${loc}`
        resultText.value = JSON.stringify(res.data || {}, null, 2)
    } catch {
        /* */
    } finally {
        dailyLoading.value = false
    }
}

async function copyResult() {
    if (!resultText.value) return
    try {
        await navigator.clipboard.writeText(resultText.value)
        ElMessage.success('已复制')
    } catch {
        ElMessage.error('复制失败')
    }
}

onMounted(loadStatus)
</script>

<style scoped lang="scss">
.panel {
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 12px;
}
.panel-head {
    margin-bottom: 12px;
}
.panel-head--row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}
.panel-title {
    font-size: 15px;
    font-weight: 650;
}
.panel-desc {
    margin: 4px 0 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.status-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 16px;
    align-items: center;
}
.pill {
    display: inline-flex;
    padding: 2px 10px;
    border-radius: 999px;
    font-size: 12px;
    font-weight: 650;
}
.pill.on {
    background: #e8f7ef;
    color: #0a7a4b;
}
.pill.off {
    background: #f3f3f3;
    color: #888;
}
.meta {
    font-size: 13px;
    color: var(--el-text-color-regular);
}
.err {
    margin: 10px 0 0;
    color: var(--el-color-danger);
    font-size: 13px;
}
.row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
}
.geo-list {
    list-style: none;
    margin: 12px 0 0;
    padding: 0;
}
.geo-list li {
    border-bottom: 1px solid var(--el-border-color-extra-light);
}
.geo-list li:last-child {
    border-bottom: 0;
}
.geo-pick {
    width: 100%;
    text-align: left;
    border: 0;
    background: transparent;
    padding: 10px 0;
    cursor: pointer;
    display: flex;
    flex-wrap: wrap;
    gap: 6px 10px;
    align-items: baseline;
}
.geo-pick:hover {
    color: var(--el-color-primary);
}
.geo-pick code {
    font-size: 12px;
    background: var(--el-fill-color);
    padding: 1px 6px;
    border-radius: 4px;
}
.geo-pick .sub {
    width: 100%;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.json {
    margin: 0;
    padding: 12px;
    border-radius: 8px;
    background: var(--el-fill-color-light);
    font-size: 12px;
    line-height: 1.45;
    overflow: auto;
    max-height: 480px;
    white-space: pre-wrap;
    word-break: break-all;
}
</style>
