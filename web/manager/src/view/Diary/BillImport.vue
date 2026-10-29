<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElRadioGroup v-model="activeTab" size="small">
                    <ElRadioButton value="preview">上传与预览</ElRadioButton>
                    <ElRadioButton value="aliases">别名管理</ElRadioButton>
                    <ElRadioButton value="batches">导入记录</ElRadioButton>
                </ElRadioGroup>
                <ElForm v-if="activeTab === 'preview'" inline class="toolbar-form">
                    <ElFormItem>
                        <input type="file" accept=".csv" ref="fileInput" @change="onFileChange"/>
                    </ElFormItem>
                </ElForm>
            </template>
            <template #right>
                <template v-if="activeTab === 'preview'">
                    <ElButton type="success" icon="Upload" :loading="uploading" @click="uploadFile">上传并预览</ElButton>
                    <ElButton v-if="batchId" :loading="previewing" @click="refreshPreview">刷新预览</ElButton>
                    <ElButton v-if="batchId" type="warning" :loading="seeding" @click="seedFromUnmapped">从未识别生成草稿</ElButton>
                    <ElButton v-if="batchId" type="primary" :loading="applying" @click="doApply">写入日记账单</ElButton>
                </template>
                <template v-else-if="activeTab === 'aliases'">
                    <ElButton type="success" icon="Plus" @click="addRule">新增规则</ElButton>
                    <ElButton @click="loadAliases">重新加载</ElButton>
                    <ElButton type="primary" :loading="aliasSaving" @click="saveAliases">保存别名</ElButton>
                </template>
                <template v-else-if="activeTab === 'batches'">
                    <ElButton icon="Refresh" :loading="batchLoading" @click="loadBatches">刷新</ElButton>
                </template>
            </template>
        </Toolbar>

        <!-- 上传与预览 -->
        <Content v-if="activeTab === 'preview'" padding="12px">
            <div v-if="batchId" class="batch-meta">
                批次 #{{ batchId }}
                <span v-if="batchMeta?.filename"> · {{ batchMeta.filename }}</span>
                <span v-if="counts"> · 共 {{ counts.row }} 笔（可写 {{ counts.ready }} / 已导入 {{ counts.imported || 0 }} / 待确认名 {{ counts.guess }} / 大额 {{ counts.confirm }} / 未识别 {{ counts.unmapped }} / 跳过 {{ counts.skip }}）</span>
            </div>

            <ElForm v-if="batchId" inline class="apply-bar">
                <ElFormItem>
                    <ElCheckbox v-model="hideZeroAmount">隐藏金额为 0</ElCheckbox>
                </ElFormItem>
                <ElFormItem>
                    <ElCheckbox v-model="includeGuess">含名称待确认</ElCheckbox>
                </ElFormItem>
                <ElFormItem>
                    <ElCheckbox v-model="includeConfirm">含大额待确认</ElCheckbox>
                </ElFormItem>
                <ElFormItem>
                    <ElCheckbox v-model="includeZero">写入含金额 0</ElCheckbox>
                </ElFormItem>
            </ElForm>

            <ElRow :gutter="12" v-if="preview">
                <ElCol :span="6">
                    <h4>可写入</h4>
                    <pre class="preview-block">{{ filteredReadyDays }}</pre>
                </ElCol>
                <ElCol :span="6">
                    <h4>已导入</h4>
                    <pre class="preview-block imported-block">{{ filteredImportedDays }}</pre>
                </ElCol>
                <ElCol :span="6">
                    <h4>名称待确认</h4>
                    <pre class="preview-block">{{ filteredGuessDays }}</pre>
                </ElCol>
                <ElCol :span="6">
                    <h4>大额待确认</h4>
                    <pre class="preview-block">{{ filteredConfirmDays }}</pre>
                </ElCol>
            </ElRow>

            <ElRow :gutter="12" v-if="preview" style="margin-top: 12px">
                <ElCol :span="12">
                    <h4>未识别</h4>
                    <ElTable
                        class="table-narrow"
                        size="small"
                        stripe
                        :data="filteredUnmapped"
                        :height="previewTableHeight"
                        empty-text="(无)"
                    >
                        <ElTableColumn prop="date" label="日期" width="100"/>
                        <ElTableColumn label="金额" width="90" align="right">
                            <template #default="{ row }">{{ formatAmt(row.amount) }}</template>
                        </ElTableColumn>
                        <ElTableColumn prop="who" label="对方" min-width="120" show-overflow-tooltip/>
                        <ElTableColumn prop="product" label="商品" min-width="140" show-overflow-tooltip/>
                    </ElTable>
                </ElCol>
                <ElCol :span="12">
                    <h4>已跳过</h4>
                    <ElTable
                        class="table-narrow"
                        size="small"
                        stripe
                        :data="filteredSkip"
                        :height="previewTableHeight"
                        empty-text="(无)"
                    >
                        <ElTableColumn prop="date" label="日期" width="100"/>
                        <ElTableColumn prop="reason" label="原因" width="120"/>
                        <ElTableColumn label="金额" width="90" align="right">
                            <template #default="{ row }">{{ formatAmt(row.amount) }}</template>
                        </ElTableColumn>
                        <ElTableColumn prop="who" label="对方" min-width="120" show-overflow-tooltip/>
                        <ElTableColumn prop="product" label="商品" min-width="140" show-overflow-tooltip/>
                    </ElTable>
                </ElCol>
            </ElRow>
        </Content>

        <!-- 别名管理：操作在 Toolbar，表格顶满高度 -->
        <Content v-else-if="activeTab === 'aliases'" padding="0">
            <ElTable
                class="table-narrow"
                size="small"
                stripe
                :data="aliasRules"
                :height="tableHeight"
                v-loading="aliasLoading"
            >
                <ElTableColumn type="index" label="#" width="50" align="center"/>
                <ElTableColumn label="参考时间" width="140" align="center">
                    <template #default="{ row }">
                        <span class="ref-muted">{{ displayRef(row).when || '—' }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="参考金额" width="90" align="right">
                    <template #default="{ row }">
                        <span class="ref-muted">{{ displayRef(row).when ? formatAmt(displayRef(row).amount) : '—' }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="参考对方 / 商品" min-width="160" show-overflow-tooltip>
                    <template #default="{ row }">
                        <span class="ref-muted" :title="ruleRefTip(row)">{{ ruleRefLabel(row) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="匹配词（逗号分隔）" min-width="200">
                    <template #default="{ row }">
                        <ElInput v-model="row.matchText" size="small" placeholder="济南和仟,金舆店"/>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="短名称" min-width="140">
                    <template #default="{ row }">
                        <ElInput v-model="row.name" size="small"/>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="餐次门店" min-width="120">
                    <template #default="{ row }">
                        <ElInput v-model="row.meal_place" size="small" placeholder="meal_place"/>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="商品限定" min-width="120">
                    <template #default="{ row }">
                        <ElInput v-model="row.whenProductText" size="small" placeholder="可选，逗号分隔"/>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="待确认" width="70" align="center">
                    <template #default="{ row }">
                        <ElCheckbox v-model="row.guess"/>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="大额" width="60" align="center">
                    <template #default="{ row }">
                        <ElCheckbox v-model="row.confirm"/>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="操作" width="90" align="center" fixed="right">
                    <template #default="{ $index }">
                        <ElButton type="danger" plain icon="Delete" size="small" @click="aliasRules.splice($index, 1)">删除</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
        </Content>

        <!-- 导入记录 -->
        <Content v-else padding="0">
            <ElTable
                class="table-narrow"
                size="small"
                stripe
                :data="batchList"
                :height="tableHeight"
                v-loading="batchLoading"
            >
                <ElTableColumn prop="id" label="ID" width="70"/>
                <ElTableColumn prop="filename" label="文件名" min-width="200" show-overflow-tooltip/>
                <ElTableColumn label="区间" min-width="180">
                    <template #default="{ row }">
                        <span v-if="row.period_start">{{ String(row.period_start).slice(0, 10) }} ~ {{ String(row.period_end || '').slice(0, 10) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn prop="row_count" label="笔数" width="70" align="right"/>
                <ElTableColumn prop="ready_count" label="可写" width="60" align="right"/>
                <ElTableColumn prop="unmapped_count" label="未识别" width="70" align="right"/>
                <ElTableColumn prop="applied_count" label="已写" width="60" align="right"/>
                <ElTableColumn prop="status" label="状态" width="90"/>
                <ElTableColumn prop="date_init" label="上传时间" width="160"/>
                <ElTableColumn label="操作" width="180" align="center" fixed="right">
                    <template #default="{ row }">
                        <ElButton size="small" plain type="primary" @click="openBatch(row.id)">预览</ElButton>
                        <ElButton size="small" plain @click="showOrders(row.id)">订单</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
        </Content>

        <ElDialog v-model="ordersVisible" title="已写入订单" width="720px">
            <ElTable size="small" stripe :data="orderList" max-height="420">
                <ElTableColumn prop="order_id" label="订单号" min-width="200" show-overflow-tooltip/>
                <ElTableColumn prop="bill_date" label="日期" width="110"/>
                <ElTableColumn prop="name" label="名称" min-width="120"/>
                <ElTableColumn prop="amount" label="金额" width="90" align="right"/>
            </ElTable>
        </ElDialog>
    </Container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import billImportApi, { type AliasConfig, type AliasRule } from '@/api/billImportApi'
import { useProjectStore } from '@/pinia'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'

interface EditableRule extends AliasRule {
    matchText: string
    whenProductText: string
}

interface RuleRef {
    match_key: string
    when: string
    amount: number
    who: string
    product: string
    count: number
}

const projectStore = useProjectStore()
// Content 已扣除 toolbar；无分页时表格高度应等于内容区，避免底部留白
const tableHeight = computed(() =>
    projectStore.contentInsets.heightContent - projectStore.contentInsets.heightToolbar
)
const previewTableHeight = computed(() => Math.max(220, tableHeight.value - 280))

const activeTab = ref('preview')
const fileInput = ref<HTMLInputElement | null>(null)
const currentFile = ref<File | null>(null)
const uploading = ref(false)
const previewing = ref(false)
const seeding = ref(false)
const applying = ref(false)
const includeGuess = ref(false)
const includeConfirm = ref(false)
const includeZero = ref(false)
const hideZeroAmount = ref(false)

const batchId = ref<number | null>(null)
const batchMeta = ref<any>(null)
const preview = ref<any>(null)
const counts = ref<any>(null)
const ruleRefs = ref<RuleRef[]>([])

function isZeroAmount(n: number) {
    return Math.abs(Number(n) || 0) < 0.0001
}

function filterItems<T extends { amount: number }>(list: T[] | undefined): T[] {
    const rows = list || []
    if (!hideZeroAmount.value) return rows
    return rows.filter((r) => !isZeroAmount(r.amount))
}

function renderFilteredDays(lines: any[] | undefined) {
    const rows = filterItems(lines)
    if (!rows.length) return '(无)'
    const byDate: Record<string, any[]> = {}
    for (const line of rows) {
        ;(byDate[line.date] ||= []).push(line)
    }
    return Object.keys(byDate).sort().map((date) => {
        const body = byDate[date].map((l) => `${l.name} ${formatAmt(l.amount)}`).join('\n')
        return `${date}\n${body}`
    }).join('\n\n')
}

const filteredReadyDays = computed(() => renderFilteredDays(preview.value?.ready))
const filteredImportedDays = computed(() => renderFilteredDays(preview.value?.imported))
const filteredGuessDays = computed(() => renderFilteredDays(preview.value?.guess))
const filteredConfirmDays = computed(() => renderFilteredDays(preview.value?.confirm))
const filteredUnmapped = computed(() => filterItems(preview.value?.unmapped))
const filteredSkip = computed(() => filterItems(preview.value?.skip))

// 优先本批次命中，否则用已持久化的参考值
function displayRef(row: EditableRule): RuleRef {
    const key = (row.matchText || '').split(/[,，]/).map((s) => s.trim()).filter(Boolean).join(',')
    const hit = ruleRefs.value.find((r) => r.match_key === key && r.when)
    if (hit) return hit
    return {
        match_key: key,
        when: row.ref_when || '',
        amount: row.ref_amount == null ? 0 : Number(row.ref_amount),
        who: row.ref_who || '',
        product: row.ref_product || '',
        count: 0,
    }
}

function ruleRefLabel(row: EditableRule) {
    const ref = displayRef(row)
    if (!ref.when) return batchId.value ? '本批次无命中' : '—'
    const who = ref.who || '—'
    const n = ref.count > 1 ? ` ×${ref.count}` : ''
    return `${who}${n}`
}

function ruleRefTip(row: EditableRule) {
    const ref = displayRef(row)
    if (!ref.when) return ''
    return `${ref.when}  ${formatAmt(ref.amount)}  ${ref.who} / ${ref.product}`
}

// 把批次样例写回当前编辑行（保存别名时一并持久化）
function mergeRefsIntoAliasRules() {
    for (const row of aliasRules.value) {
        const key = (row.matchText || '').split(/[,，]/).map((s) => s.trim()).filter(Boolean).join(',')
        const hit = ruleRefs.value.find((r) => r.match_key === key && r.when)
        if (!hit) continue
        row.ref_when = hit.when
        row.ref_amount = hit.amount
        row.ref_who = hit.who
        row.ref_product = hit.product
    }
}

const aliasSlots = ref<AliasConfig['slots']>([])
const aliasRules = ref<EditableRule[]>([])
const aliasSaving = ref(false)
const aliasLoading = ref(false)

const batchList = ref<any[]>([])
const batchLoading = ref(false)
const ordersVisible = ref(false)
const orderList = ref<any[]>([])

function formatAmt(n: number) {
    if (n == null) return ''
    const t = Math.abs(n).toFixed(2).replace(/\.?0+$/, '')
    return n > 0 ? `+${t}` : `-${t}`
}

function onFileChange(e: Event) {
    const input = e.target as HTMLInputElement
    currentFile.value = input.files?.[0] || null
}

function applyStoredRefs(cfg: AliasConfig) {
    const byKey = new Map<string, AliasRule>()
    for (const r of cfg.rules || []) {
        byKey.set((r.match || []).join(','), r)
    }
    for (const row of aliasRules.value) {
        const key = (row.matchText || '').split(/[,，]/).map((s) => s.trim()).filter(Boolean).join(',')
        const src = byKey.get(key)
        if (!src?.ref_when) continue
        row.ref_when = src.ref_when
        row.ref_amount = src.ref_amount == null ? null : Number(src.ref_amount)
        row.ref_who = src.ref_who || ''
        row.ref_product = src.ref_product || ''
    }
}

function applyPreviewPayload(data: any, opts?: { replaceAliases?: boolean }) {
    batchId.value = data.batch_id
    preview.value = data.preview
    counts.value = data.preview?.counts
    batchMeta.value = data.batch || batchMeta.value
    if (data.preview?.rule_refs) {
        ruleRefs.value = data.preview.rule_refs
    }
    if (data.aliases) {
        if (opts?.replaceAliases || !aliasRules.value.length) {
            setAliasEditor(data.aliases)
        } else {
            applyStoredRefs(data.aliases)
            mergeRefsIntoAliasRules()
        }
    } else if (data.preview?.rule_refs) {
        mergeRefsIntoAliasRules()
    }
}

async function uploadFile() {
    if (!currentFile.value) {
        ElMessage.warning('请先选择 CSV 文件')
        return
    }
    uploading.value = true
    try {
        const fd = new FormData()
        fd.append('file', currentFile.value)
        const res: any = await billImportApi.upload(fd)
        applyPreviewPayload(res.data)
        ElMessage.success(res.message || '上传成功')
    } catch (e: any) {
        ElMessage.error(e?.message || '上传失败')
    } finally {
        uploading.value = false
    }
}

async function refreshPreview() {
    if (!batchId.value) return
    previewing.value = true
    try {
        const res: any = await billImportApi.preview(batchId.value)
        applyPreviewPayload(res.data)
    } catch (e: any) {
        ElMessage.error(e?.message || '预览失败')
    } finally {
        previewing.value = false
    }
}

async function seedFromUnmapped() {
    if (!batchId.value) return
    seeding.value = true
    try {
        const res: any = await billImportApi.seedAliases(batchId.value)
        applyPreviewPayload({
            batch_id: batchId.value,
            preview: res.data.preview,
            batch: batchMeta.value,
            aliases: res.data.aliases,
        }, { replaceAliases: true })
        ElMessage.success(res.message || `已追加 ${res.data.added} 条`)
        activeTab.value = 'aliases'
    } catch (e: any) {
        ElMessage.error(e?.message || '生成草稿失败')
    } finally {
        seeding.value = false
    }
}

async function doApply() {
    if (!batchId.value) return
    try {
        await ElMessageBox.confirm(
            `将写入「可写入」${includeGuess.value ? ' + 名称待确认' : ''}${includeConfirm.value ? ' + 大额待确认' : ''}，确认？`,
            '写入日记',
            { type: 'warning' },
        )
    } catch {
        return
    }
    applying.value = true
    try {
        const res: any = await billImportApi.apply({
            batchId: batchId.value,
            includeGuess: includeGuess.value,
            includeConfirm: includeConfirm.value,
            includeZero: includeZero.value,
        })
        const diaryCount = res.data?.diary_count ?? 0
        const lineCount = res.data?.line_count ?? res.data?.applied ?? 0
        ElNotification({
            title: '写入完成',
            message: diaryCount > 0
                ? `已写入 ${diaryCount} 条日记账单（${lineCount} 笔明细）`
                : (res.message || '没有新的日记需要写入'),
            type: diaryCount > 0 ? 'success' : 'info',
            duration: 4500,
        })
        loadBatches()
    } catch (e: any) {
        ElNotification({
            title: '写入失败',
            message: e?.message || '写入失败',
            type: 'error',
        })
    } finally {
        applying.value = false
    }
}

function toEditable(rules: AliasRule[]): EditableRule[] {
    return (rules || []).map((r) => ({
        ...r,
        matchText: (r.match || []).join(','),
        whenProductText: (r.when_product || []).join(','),
    }))
}

function fromEditable(rules: EditableRule[]): AliasRule[] {
    return rules.map((r) => {
        const out: AliasRule = {
            match: r.matchText.split(/[,，]/).map((s) => s.trim()).filter(Boolean),
            guess: !!r.guess,
            confirm: !!r.confirm,
        }
        if (r.name) out.name = r.name
        if (r.meal_place) out.meal_place = r.meal_place
        const wp = r.whenProductText.split(/[,，]/).map((s) => s.trim()).filter(Boolean)
        if (wp.length) out.when_product = wp
        if (r.meal_order) out.meal_order = r.meal_order
        if (r.slot_names) out.slot_names = r.slot_names
        // 保留已记录的参考流水
        if (r.ref_when) {
            out.ref_when = r.ref_when
            out.ref_who = r.ref_who || ''
            out.ref_product = r.ref_product || ''
            if (r.ref_amount != null && r.ref_amount !== undefined) {
                out.ref_amount = Number(r.ref_amount)
            }
        }
        return out
    })
}

function setAliasEditor(cfg: AliasConfig) {
    aliasSlots.value = cfg.slots || []
    aliasRules.value = toEditable(cfg.rules || [])
}

async function loadAliases() {
    aliasLoading.value = true
    try {
        const res: any = await billImportApi.getAliases()
        setAliasEditor(res.data)
    } catch (e: any) {
        ElMessage.error(e?.message || '读取别名失败')
    } finally {
        aliasLoading.value = false
    }
}

async function saveAliases() {
    aliasSaving.value = true
    try {
        mergeRefsIntoAliasRules()
        const payload: AliasConfig = {
            slots: aliasSlots.value,
            rules: fromEditable(aliasRules.value),
        }
        const res: any = await billImportApi.putAliases(payload)
        setAliasEditor(res.data)
        ElMessage.success('别名已保存')
        if (batchId.value) refreshPreview()
    } catch (e: any) {
        ElMessage.error(e?.message || '保存失败')
    } finally {
        aliasSaving.value = false
    }
}

function addRule() {
    aliasRules.value.unshift({
        match: [],
        matchText: '',
        whenProductText: '',
        name: '待命名-',
        guess: true,
    })
}

async function loadBatches() {
    batchLoading.value = true
    try {
        const res: any = await billImportApi.batches({ pageNo: 1, pageSize: 50 })
        batchList.value = res.data?.list || []
    } catch (e: any) {
        ElMessage.error(e?.message || '读取导入记录失败')
    } finally {
        batchLoading.value = false
    }
}

function openBatch(id: number) {
    batchId.value = id
    activeTab.value = 'preview'
    refreshPreview()
}

async function showOrders(id: number) {
    try {
        const res: any = await billImportApi.orders(id)
        orderList.value = res.data || []
        ordersVisible.value = true
    } catch (e: any) {
        ElMessage.error(e?.message || '读取订单失败')
    }
}

watch(activeTab, (tab) => {
    if (tab === 'aliases') {
        if (aliasRules.value.length === 0) loadAliases()
        // 有批次时刷新样例金额/时间
        if (batchId.value && !ruleRefs.value.length) refreshPreview()
    }
    if (tab === 'batches') loadBatches()
})

onMounted(() => {
    loadAliases()
})
</script>

<style scoped>
.toolbar-form {
    margin-left: 12px;
    display: inline-flex;
    align-items: center;
}
.toolbar-form :deep(.el-form-item) {
    margin-bottom: 0;
}
.batch-meta {
    margin: 0 0 10px;
    color: #666;
    font-size: 13px;
}
.apply-bar {
    margin-bottom: 8px;
}
.preview-block {
    background: #f7f7f7;
    border: 1px solid #eee;
    border-radius: 4px;
    padding: 10px;
    min-height: 100px;
    max-height: 240px;
    overflow: auto;
    font-size: 12px;
    line-height: 1.5;
    white-space: pre-wrap;
}
.imported-block {
    background: #f0f9eb;
    border-color: #e1f3d8;
    color: #67c23a;
}
h4 {
    margin: 0 0 8px;
    font-size: 14px;
}
.ref-muted {
    color: #888;
    font-size: 12px;
}
</style>
