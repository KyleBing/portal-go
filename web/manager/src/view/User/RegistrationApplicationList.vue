<template>
    <Container>
        <Toolbar :is-show-search-bar="false">
            <template #left>
                <ElRadioGroup v-model="status" @change="onStatusChange">
                    <ElRadioButton value="all">全部 {{ summary.total }}</ElRadioButton>
                    <ElRadioButton value="pending">待审核 {{ summary.pending }}</ElRadioButton>
                    <ElRadioButton value="approved">已通过 {{ summary.approved }}</ElRadioButton>
                    <ElRadioButton value="rejected">已拒绝 {{ summary.rejected }}</ElRadioButton>
                </ElRadioGroup>
            </template>
        </Toolbar>
        <Content padding="0" :is-show-pagination="true">
            <ElTable
                class="table-narrow"
                size="small"
                stripe
                :height="projectStore.contentInsets.heightContent - projectStore.contentInsets.heightToolbar - projectStore.contentInsets.heightPagination"
                :data="list"
                v-loading="loading"
                empty-text="暂无申请"
            >
                <ElTableColumn type="index" width="50" label="#" />
                <ElTableColumn label="申请人" min-width="220">
                    <template #default="{ row }">
                        <div class="applicant-cell">
                            <span class="applicant-cell__email">{{ row.email }}</span>
                            <span class="applicant-cell__account">付费昵称：{{ row.paid_account }}</span>
                        </div>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="金额" width="72" align="center">
                    <template #default="{ row }">
                        {{ formatPayAmount(row.pay_amount_yuan) }}
                    </template>
                </ElTableColumn>
                <ElTableColumn label="付费平台" width="120" align="center">
                    <template #default="{ row }">
                        <span class="platform-chip">
                            <img
                                v-if="payIcon(row.pay_platform)"
                                class="platform-chip__icon"
                                :src="payIcon(row.pay_platform)"
                                :alt="payLabel(row.pay_platform)"
                            >
                            <span>{{ payLabel(row.pay_platform) }}</span>
                        </span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="状态" width="100" align="center">
                    <template #default="{ row }">
                        <ElTag v-if="row.status === 'pending'" type="warning" size="small" effect="light">待审核</ElTag>
                        <ElTag v-else-if="row.status === 'approved'" type="success" size="small" effect="light">已通过</ElTag>
                        <ElTag v-else type="info" size="small" effect="light">已拒绝</ElTag>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="邀请码" min-width="200">
                    <template #default="{ row }">
                        <template v-if="row.invitation_code">
                            <div v-if="isInvitationMissing(row)" class="invite-missing">
                                <code class="code code--struck">{{ row.invitation_code }}</code>
                                <ElTag type="danger" size="small" effect="light">邀请码已删除</ElTag>
                            </div>
                            <code v-else class="code">{{ row.invitation_code }}</code>
                        </template>
                        <span v-else class="muted">—</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="申请时间" width="128" align="center">
                    <template #default="{ row }">
                        <div class="time-cell">
                            <span class="time-cell__date">{{ formatDate(row.date_create) }}</span>
                            <span class="time-cell__time">{{ formatClock(row.date_create) }}</span>
                        </div>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="操作" width="168" fixed="right" align="center">
                    <template #default="{ row }">
                        <ElButton
                            size="small"
                            :type="row.status === 'pending' ? 'primary' : 'default'"
                            :icon="row.status === 'pending' ? 'DocumentChecked' : 'View'"
                            plain
                            @click="openDetail(row.id)"
                        >
                            {{ row.status === 'pending' ? '审核' : '查看' }}
                        </ElButton>
                        <ElPopconfirm
                            title="确定删除该注册申请记录？"
                            confirm-button-text="删除"
                            cancel-button-text="取消"
                            confirm-button-type="danger"
                            width="240"
                            @confirm="remove(row.id)"
                        >
                            <template #reference>
                                <ElButton size="small" type="danger" plain icon="Delete" />
                            </template>
                        </ElPopconfirm>
                    </template>
                </ElTableColumn>
            </ElTable>
        </Content>
        <FooterPagination
            :pager-option="pager"
            @size-change="pageSizeChange"
            @pager-change="pageNoChange"
        />

        <ElDialog
            v-model="detailVisible"
            class="review-dialog"
            :title="detail?.status === 'pending' ? '审核注册申请' : '注册申请详情'"
            width="560px"
            append-to-body
            destroy-on-close
        >
            <div v-if="detail" class="detail-body" v-loading="detailLoading">
                <div class="detail-status" :class="`detail-status--${detail.status}`">
                    <ElTag :type="statusTagType(detail.status)" size="small" effect="dark">
                        {{ statusLabel(detail.status) }}
                    </ElTag>
                    <span v-if="detail.status === 'pending'" class="detail-status__hint">请核对邮箱与付费昵称后决定是否通过</span>
                    <span v-else-if="detail.date_review" class="detail-status__hint">
                        审核于 {{ formatTime(detail.date_review) }}
                    </span>
                </div>

                <ElDescriptions :column="1" border size="small" class="detail-desc">
                    <ElDescriptionsItem label="联系邮箱">
                        <span class="detail-email">{{ detail.email }}</span>
                    </ElDescriptionsItem>
                    <ElDescriptionsItem label="付费昵称">
                        {{ detail.paid_account }}
                    </ElDescriptionsItem>
                    <ElDescriptionsItem label="应付金额">
                        {{ formatPayAmount(detail.pay_amount_yuan) }}
                    </ElDescriptionsItem>
                    <ElDescriptionsItem label="付费平台">
                        <span class="platform-chip platform-chip--lg">
                            <img
                                v-if="payIcon(detail.pay_platform)"
                                class="platform-chip__icon"
                                :src="payIcon(detail.pay_platform)"
                                :alt="payLabel(detail.pay_platform)"
                            >
                            <span>{{ payLabel(detail.pay_platform) }}</span>
                        </span>
                    </ElDescriptionsItem>
                    <ElDescriptionsItem label="申请时间">
                        <span class="mono">{{ formatTime(detail.date_create) }}</span>
                    </ElDescriptionsItem>
                    <ElDescriptionsItem v-if="detail.invitation_code" label="邀请码">
                        <div v-if="detail && isInvitationMissing(detail)" class="invite-missing-detail">
                            <ElAlert type="warning" :closable="false" show-icon>
                                该邀请码已在邀请码管理中删除，用户将无法使用此码注册。如需补发，请在邀请码管理中新建并手动通知用户。
                            </ElAlert>
                            <code class="code code--block code--struck">{{ detail.invitation_code }}</code>
                        </div>
                        <div v-else class="invite-row">
                            <code class="code code--block">{{ detail.invitation_code }}</code>
                            <ElButton size="small" plain icon="CopyDocument" @click="copyCode(detail.invitation_code!)">
                                复制
                            </ElButton>
                        </div>
                    </ElDescriptionsItem>
                    <ElDescriptionsItem v-if="detail.admin_note" label="审核备注">
                        <p class="admin-note">{{ detail.admin_note }}</p>
                    </ElDescriptionsItem>
                </ElDescriptions>

                <div v-if="detail.status === 'pending'" class="review-form">
                    <div class="review-form__title">审核操作</div>
                    <ElForm label-position="top">
                        <ElFormItem label="备注（可选）">
                            <ElInput
                                v-model="adminNote"
                                type="textarea"
                                :rows="3"
                                placeholder="内部备注；拒绝时可填写原因，便于日后查阅"
                            />
                        </ElFormItem>
                    </ElForm>
                </div>
            </div>
            <template #footer>
                <div class="dialog-footer">
                    <ElButton icon="Close" @click="detailVisible = false">关闭</ElButton>
                    <template v-if="detail?.status === 'pending'">
                        <ElButton type="danger" plain icon="CircleClose" :loading="acting" @click="reject">拒绝</ElButton>
                        <ElButton type="success" icon="Promotion" :loading="acting" @click="approve">通过并发送邀请码邮件</ElButton>
                    </template>
                </div>
            </template>
        </ElDialog>
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
import registrationApplicationApi, {
    type ApplicationStatus,
    type RegistrationApplicationRow,
} from '@/api/registrationApplicationApi'
import alipayIcon from '@/assets/alipay.png'
import wepayIcon from '@/assets/wepay.png'
import { formatRegistrationPayAmount } from '@/utils/payQrAssets'

const formatPayAmount = formatRegistrationPayAmount

const projectStore = useProjectStore()

const status = ref<ApplicationStatus>('pending')
const loading = ref(false)
const list = ref<RegistrationApplicationRow[]>([])
const summary = reactive({ total: 0, pending: 0, approved: 0, rejected: 0 })
const pager = reactive({ pageNo: 1, pageSize: 20, total: 0 })

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<RegistrationApplicationRow | null>(null)
const adminNote = ref('')
const acting = ref(false)
let currentId = 0

function payLabel(p: string) {
    if (p === 'alipay') return '支付宝'
    if (p === 'wechat') return '微信'
    return p || '—'
}

function payIcon(p: string): string | undefined {
    if (p === 'alipay') return alipayIcon
    if (p === 'wechat') return wepayIcon
    return undefined
}

function statusLabel(s: string) {
    if (s === 'pending') return '待审核'
    if (s === 'approved') return '已通过'
    if (s === 'rejected') return '已拒绝'
    return s
}

function statusTagType(s: string): 'warning' | 'success' | 'info' {
    if (s === 'pending') return 'warning'
    if (s === 'approved') return 'success'
    return 'info'
}

function formatTime(v: string | null) {
    if (!v) return '—'
    return dayjs(v).format('YYYY-MM-DD HH:mm')
}

function formatDate(v: string | null) {
    if (!v) return '—'
    return dayjs(v).format('YYYY-MM-DD')
}

function formatClock(v: string | null) {
    if (!v) return ''
    return dayjs(v).format('HH:mm')
}

function isInvitationMissing(row: RegistrationApplicationRow) {
    return Boolean(row.invitation_code) && Number(row.invitation_missing) === 1
}

function copyCode(code: string) {
    navigator.clipboard.writeText(code).then(
        () => ElMessage.success('已复制邀请码'),
        () => ElMessage.error('复制失败'),
    )
}

function remove(id: number) {
    registrationApplicationApi
        .delete({ id })
        .then((res: any) => {
            ElMessage.success(res.message || '已删除')
            if (detailVisible.value && currentId === id) {
                detailVisible.value = false
            }
            load()
        })
        .catch(() => {})
}

function load() {
    loading.value = true
    registrationApplicationApi
        .manage({ status: status.value, pageNo: pager.pageNo, pageSize: pager.pageSize })
        .then((res: any) => {
            list.value = res.data?.list ?? []
            pager.total = res.data?.pager?.total ?? 0
            const s = res.data?.summary ?? {}
            summary.total = s.total ?? 0
            summary.pending = s.pending ?? 0
            summary.approved = s.approved ?? 0
            summary.rejected = s.rejected ?? 0
        })
        .finally(() => {
            loading.value = false
        })
}

function onStatusChange() {
    pager.pageNo = 1
    load()
}

function pageNoChange(page: number) {
    pager.pageNo = page
    load()
}

function pageSizeChange(size: number) {
    pager.pageSize = size
    pager.pageNo = 1
    load()
}

function openDetail(id: number) {
    currentId = id
    detailVisible.value = true
    detail.value = null
    adminNote.value = ''
    detailLoading.value = true
    registrationApplicationApi
        .detail({ id })
        .then((res: any) => {
            detail.value = res.data
        })
        .finally(() => {
            detailLoading.value = false
        })
}

function approve() {
    if (!currentId) return
    ElMessageBox.confirm('确认通过该申请？系统将生成邀请码并发送到用户邮箱。', '通过审核', {
        type: 'warning',
    })
        .then(() => {
            acting.value = true
            return registrationApplicationApi.approve({ id: currentId, adminNote: adminNote.value.trim() || undefined })
        })
        .then((res: any) => {
            ElMessage.success(res.message || '已通过')
            detailVisible.value = false
            load()
        })
        .catch(() => {})
        .finally(() => {
            acting.value = false
        })
}

function reject() {
    if (!currentId) return
    ElMessageBox.confirm('确认拒绝该申请？', '拒绝申请', { type: 'warning' })
        .then(() => {
            acting.value = true
            return registrationApplicationApi.reject({ id: currentId, adminNote: adminNote.value.trim() || undefined })
        })
        .then((res: any) => {
            ElMessage.success(res.message || '已拒绝')
            detailVisible.value = false
            load()
        })
        .catch(() => {})
        .finally(() => {
            acting.value = false
        })
}

onMounted(load)
</script>

<style scoped lang="scss">
.applicant-cell {
    display: flex;
    flex-direction: column;
    gap: 4px;
    line-height: 1.35;

    &__email {
        color: #1f2329;
        font-weight: 600;
        word-break: break-all;
    }

    &__account {
        color: #86909c;
        font-size: 12px;
    }
}

.platform-chip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    font-size: 12px;
    color: #4e5969;

    &--lg {
        font-size: 13px;
    }

    &__icon {
        width: 18px;
        height: 18px;
        object-fit: contain;
    }
}

.time-cell {
    display: flex;
    flex-direction: column;
    gap: 2px;
    line-height: 1.3;

    &__date {
        font-size: 12px;
        color: #1f2329;
        font-variant-numeric: tabular-nums;
    }

    &__time {
        font-size: 11px;
        color: #86909c;
        font-variant-numeric: tabular-nums;
    }
}

.invite-missing {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
}

.invite-missing-detail {
    display: flex;
    flex-direction: column;
    gap: 10px;
    width: 100%;
}

.code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 12px;
    word-break: break-all;

    &--struck {
        color: #86909c;
        text-decoration: line-through;
    }

    &--block {
        flex: 1;
        min-width: 0;
        padding: 6px 10px;
        border-radius: 6px;
        background: #f4f5f7;
        border: 1px solid #e5e6eb;
    }
}

.muted {
    color: #c0c4cc;
    font-size: 12px;
}

.mono {
    font-variant-numeric: tabular-nums;
    font-size: 13px;
}

.detail-body {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.detail-status {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
    padding: 12px 14px;
    border-radius: 8px;
    border: 1px solid #e5e6eb;

    &--pending {
        background: #fffbeb;
        border-color: #fde68a;
    }

    &--approved {
        background: #ecfdf5;
        border-color: #a7f3d0;
    }

    &--rejected {
        background: #f4f5f7;
    }

    &__hint {
        font-size: 13px;
        color: #4e5969;
    }
}

.detail-desc {
    :deep(.el-descriptions__label) {
        width: 108px;
        font-weight: 500;
        color: #4e5969;
    }

    :deep(.el-descriptions__content) {
        color: #1f2329;
    }
}

.detail-email {
    font-weight: 600;
    word-break: break-all;
}

.invite-row {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
}

.admin-note {
    margin: 0;
    padding: 8px 10px;
    border-radius: 6px;
    background: #f9fafb;
    border-left: 3px solid #d1d5db;
    font-size: 13px;
    line-height: 1.5;
    color: #374151;
    white-space: pre-wrap;
    word-break: break-word;
}

.review-form {
    padding-top: 4px;
    border-top: 1px dashed #e5e6eb;

    &__title {
        margin-bottom: 12px;
        font-size: 13px;
        font-weight: 600;
        color: #1f2329;
    }
}

.dialog-footer {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 8px;
}
</style>
