<template>
    <Container>
        <Toolbar :is-show-search-bar="false">
            <template #left>
                <ElRadioGroup v-model="status" size="small" @change="onStatusChange">
                    <ElRadioButton value="all">全部 {{ summary.total }}</ElRadioButton>
                    <ElRadioButton value="unused">未使用 {{ summary.unused }}</ElRadioButton>
                    <ElRadioButton value="used">已使用 {{ summary.used }}</ElRadioButton>
                </ElRadioGroup>
                <span class="shared-hint">已分享未用 {{ summary.shared_unused }}</span>
            </template>
            <template #right>
                <ElButton type="success" icon="Plus" :loading="generating" @click="generate">生成邀请码</ElButton>
            </template>
        </Toolbar>
        <Content padding="0" :is-show-pagination="true">
            <ElTable
                class="table-narrow"
                size="small"
                stripe
                :height="projectStore.contentInsets.heightContent - 130"
                :data="list"
                v-loading="loading"
                empty-text="暂无邀请码"
            >
                <ElTableColumn type="index" width="50" label="#"/>
                <ElTableColumn prop="id" label="邀请码" min-width="180">
                    <template #default="{ row }">
                        <code class="code">{{ row.id }}</code>
                        <ElButton link type="primary" size="small" @click="copyCode(row.id)">复制</ElButton>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="状态" width="100" align="center">
                    <template #default="{ row }">
                        <ElTag v-if="row.binding_uid" type="success" size="small">已使用</ElTag>
                        <ElTag v-else-if="Number(row.is_shared) === 1" type="warning" size="small">已分享</ElTag>
                        <ElTag v-else type="info" size="small">未使用</ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="绑定用户" min-width="200">
                    <template #default="{ row }">
                        <template v-if="row.binding_uid">
                            <div class="user-cell">
                                <span class="name">{{ row.binding_nickname || row.binding_username || '—' }}</span>
                                <span class="meta">uid {{ row.binding_uid }} · {{ row.binding_email || '无邮箱' }}</span>
                            </div>
                        </template>
                        <span v-else class="muted">—</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="创建时间" width="160" align="center">
                    <template #default="{ row }">
                        <span class="mono">{{ formatTime(row.date_create) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="注册时间" width="160" align="center">
                    <template #default="{ row }">
                        <span class="mono">{{ formatTime(row.date_register) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="操作" width="200" fixed="right" align="left">
                    <template #default="{ row }">
                        <template v-if="!row.binding_uid">
                            <ElButton
                                v-if="Number(row.is_shared) !== 1"
                                size="small"
                                type="warning"
                                plain
                                @click="markShared(row.id)"
                            >标记已分享</ElButton>
                            <ElButton size="small" type="danger" plain icon="Delete" @click="remove(row.id)">删除</ElButton>
                        </template>
                        <span v-else class="muted">已绑定不可删</span>
                    </template>
                </ElTableColumn>
            </ElTable>
        </Content>
        <FooterPagination
            :pager-option="pager"
            @size-change="pageSizeChange"
            @pager-change="pageNoChange"
        />
    </Container>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import dayjs from 'dayjs'
import { ElMessage, ElMessageBox } from 'element-plus'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'
import FooterPagination from '@/layout/FooterPagination.vue'
import { useProjectStore } from '@/pinia'
import invitationApi, { type InvitationStatus } from '@/api/invitationApi'

interface InvitationRow {
    id: string
    date_create: string
    date_register: string | null
    binding_uid: number | null
    is_shared: number
    binding_nickname?: string
    binding_email?: string
    binding_username?: string
}

interface Pager {
    total: number
    pageNo: number
    pageSize: number
}

const projectStore = useProjectStore()
const status = ref<InvitationStatus>('all')
const loading = ref(false)
const generating = ref(false)
const list = ref<InvitationRow[]>([])
const pager = ref<Pager>({ total: 0, pageNo: 1, pageSize: 20 })
const summary = reactive({
    total: 0,
    unused: 0,
    used: 0,
    shared_unused: 0,
})

function formatTime(v: string | null | undefined) {
    if (!v) return '—'
    return dayjs(v).format('YYYY-MM-DD HH:mm')
}

function load() {
    loading.value = true
    invitationApi.manage({
        status: status.value,
        pageNo: pager.value.pageNo,
        pageSize: pager.value.pageSize,
    })
        .then((res: any) => {
            const data = res?.data || {}
            list.value = data.list || []
            pager.value.total = Number(data.pager?.total) || 0
            Object.assign(summary, {
                total: Number(data.summary?.total) || 0,
                unused: Number(data.summary?.unused) || 0,
                used: Number(data.summary?.used) || 0,
                shared_unused: Number(data.summary?.shared_unused) || 0,
            })
        })
        .catch(() => {
            list.value = []
            pager.value.total = 0
        })
        .finally(() => {
            loading.value = false
        })
}

function onStatusChange() {
    pager.value.pageNo = 1
    load()
}

function pageSizeChange(pageSize: number) {
    pager.value.pageSize = pageSize
    pager.value.pageNo = 1
    load()
}

function pageNoChange(pageNo: number) {
    pager.value.pageNo = pageNo
    load()
}

function generate() {
    generating.value = true
    invitationApi.generate()
        .then((res: any) => {
            ElMessage.success(res?.message || '已生成')
            status.value = 'unused'
            pager.value.pageNo = 1
            load()
        })
        .finally(() => {
            generating.value = false
        })
}

async function copyCode(id: string) {
    const text = `邀请码：\n\n${id}`
    try {
        await navigator.clipboard.writeText(text)
        ElMessage.success('已复制邀请码')
    } catch {
        ElMessage.error('复制失败')
    }
}

function markShared(id: string) {
    invitationApi.markShared({ id }).then((res: any) => {
        ElMessage.success(res?.message || '已标记')
        load()
    })
}

function remove(id: string) {
    ElMessageBox.confirm(`确定删除邀请码 ${id}？`, '删除确认', { type: 'warning' })
        .then(() => invitationApi.delete({ id }))
        .then((res: any) => {
            ElMessage.success(res?.message || '已删除')
            load()
        })
        .catch(() => {})
}

onMounted(load)
</script>

<style scoped lang="scss">
.shared-hint {
    margin-left: 12px;
    font-size: 12px;
    color: #86909c;
}

.code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    margin-right: 6px;
    word-break: break-all;
}

.user-cell {
    display: flex;
    flex-direction: column;
    gap: 2px;
    .name { color: #1f2329; font-weight: 600; }
    .meta { color: #86909c; font-size: 12px; }
}

.mono {
    font-variant-numeric: tabular-nums;
    font-size: 12px;
}

.muted { color: #c0c4cc; font-size: 12px; }
</style>
