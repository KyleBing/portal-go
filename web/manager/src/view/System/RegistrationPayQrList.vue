<template>
    <Container>
        <Toolbar :is-show-search-bar="false">
            <template #left>
                <ElButton type="success" icon="Plus" @click="openCreate">新增金额档</ElButton>
            </template>
        </Toolbar>
        <Content padding="0">
            <ElTable
                class="table-narrow"
                size="small"
                stripe
                :height="tableHeight"
                :data="list"
                v-loading="loading"
                empty-text="还没有收款码档位，请先新增"
            >
                <ElTableColumn label="金额" width="120" align="center">
                    <template #default="{ row }">
                        <div class="amount-cell">
                            <strong>{{ formatRegistrationPayAmount(row.amount_yuan) }}</strong>
                            <ElTag v-if="row.amount_yuan === currentAmount" type="success" size="small" effect="plain">当前</ElTag>
                        </div>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="支付宝" min-width="240">
                    <template #default="{ row }">
                        <div class="qr-cell">
                            <img v-if="resolvePayQrUrl(row.alipay_url)" :src="resolvePayQrUrl(row.alipay_url)" alt="支付宝收款码">
                            <span v-else class="muted">未上传</span>
                            <p class="qr-cell__note">{{ row.alipay_note || '—' }}</p>
                        </div>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="微信" min-width="240">
                    <template #default="{ row }">
                        <div class="qr-cell">
                            <img v-if="resolvePayQrUrl(row.wechat_url)" :src="resolvePayQrUrl(row.wechat_url)" alt="微信收款码">
                            <span v-else class="muted">未上传</span>
                            <p class="qr-cell__note">{{ row.wechat_note || '—' }}</p>
                        </div>
                    </template>
                </ElTableColumn>
                <ElTableColumn label="操作" width="220" fixed="right" align="center">
                    <template #default="{ row }">
                        <ElButton
                            size="small"
                            type="primary"
                            plain
                            :disabled="row.amount_yuan === currentAmount"
                            :loading="settingId === row.id"
                            @click="setCurrent(row)"
                        >设为当前</ElButton>
                        <ElButton size="small" @click="openEdit(row)">编辑</ElButton>
                        <ElButton size="small" type="danger" plain @click="remove(row)">删除</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
        </Content>

        <ElDialog v-model="dialogVisible" :title="form.id ? '编辑收款码' : '新增金额档'" width="620px">
            <ElForm label-width="110px" size="small">
                <ElFormItem label="金额（元）">
                    <ElInputNumber v-model="form.amount_yuan" :min="0" :max="100000" :step="1" :precision="0" />
                    <p class="hint">填 0 表示自定义金额（无固定应付额）。</p>
                </ElFormItem>
                <ElFormItem label="支付宝收款码">
                    <div class="qr-edit">
                        <ElUpload
                            action="#"
                            accept="image/*"
                            list-type="picture-card"
                            :limit="1"
                            :file-list="alipayFileList"
                            :before-upload="beforeUploadFile"
                            :http-request="(opt) => uploadToQiniu('alipay', opt)"
                            :on-remove="() => clearQr('alipay')"
                        >
                            <ElIcon><Plus /></ElIcon>
                        </ElUpload>
                        <div class="qr-edit__fields">
                            <ElInput v-model="form.alipay_url" placeholder="也可直接填写图片地址" clearable @change="syncFileList('alipay')" />
                            <ElInput v-model="form.alipay_note" type="textarea" :rows="2" maxlength="80" show-word-limit placeholder="支付宝收款码备注说明" />
                        </div>
                    </div>
                </ElFormItem>
                <ElFormItem label="微信收款码">
                    <div class="qr-edit">
                        <ElUpload
                            action="#"
                            accept="image/*"
                            list-type="picture-card"
                            :limit="1"
                            :file-list="wechatFileList"
                            :before-upload="beforeUploadFile"
                            :http-request="(opt) => uploadToQiniu('wechat', opt)"
                            :on-remove="() => clearQr('wechat')"
                        >
                            <ElIcon><Plus /></ElIcon>
                        </ElUpload>
                        <div class="qr-edit__fields">
                            <ElInput v-model="form.wechat_url" placeholder="也可直接填写图片地址" clearable @change="syncFileList('wechat')" />
                            <ElInput v-model="form.wechat_note" type="textarea" :rows="2" maxlength="80" show-word-limit placeholder="微信收款码备注说明" />
                        </div>
                    </div>
                </ElFormItem>
            </ElForm>
            <template #footer>
                <ElButton @click="dialogVisible = false">取消</ElButton>
                <ElButton type="primary" :loading="saving" @click="saveItem">保存</ElButton>
            </template>
        </ElDialog>
    </Container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { UploadRawFile, UploadRequestOptions, UploadUserFile } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import * as qiniu from 'qiniu-js'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'
import { useProjectStore } from '@/pinia'
import { BucketQiniuMap } from '@/configProject'
import systemApi from '@/api/systemApi'
import imageQiniuApi from '@/api/imageQiniuApi'
import registrationPayQrApi, { type PayQrPlatform, type RegistrationPayQrItem } from '@/api/registrationPayQrApi'
import { formatRegistrationPayAmount, resolvePayQrUrl } from '@/utils/payQrAssets'

const PAY_QR_BUCKET = 'diary-container'

const projectStore = useProjectStore()
const tableHeight = computed(() =>
    projectStore.contentInsets.heightContent - projectStore.contentInsets.heightToolbar,
)

const loading = ref(false)
const saving = ref(false)
const settingId = ref(0)
const list = ref<RegistrationPayQrItem[]>([])
const currentAmount = ref(28)
const dialogVisible = ref(false)
const form = ref<RegistrationPayQrItem>(emptyForm())
const alipayFileList = ref<UploadUserFile[]>([])
const wechatFileList = ref<UploadUserFile[]>([])

function emptyForm(): RegistrationPayQrItem {
    return {
        id: 0,
        amount_yuan: undefined as unknown as number,
        alipay_url: '',
        alipay_note: '',
        wechat_url: '',
        wechat_note: '',
    }
}

function fileListFromUrl(url: string, name: string): UploadUserFile[] {
    const resolved = resolvePayQrUrl(url)
    return resolved ? [{ name, url: resolved }] : []
}

function syncFileList(platform: PayQrPlatform) {
    if (platform === 'alipay') {
        alipayFileList.value = fileListFromUrl(form.value.alipay_url, 'alipay')
    } else {
        wechatFileList.value = fileListFromUrl(form.value.wechat_url, 'wechat')
    }
}

function clearQr(platform: PayQrPlatform) {
    if (platform === 'alipay') {
        form.value.alipay_url = ''
        alipayFileList.value = []
    } else {
        form.value.wechat_url = ''
        wechatFileList.value = []
    }
}

function load() {
    loading.value = true
    Promise.all([registrationPayQrApi.list(), systemApi.getAdmin()])
        .then(([listRes, cfgRes]: any[]) => {
            list.value = listRes.data?.list || []
            const n = Number(cfgRes.data?.registration_pay_amount_yuan)
            currentAmount.value = Number.isFinite(n) ? n : 28
        })
        .finally(() => {
            loading.value = false
        })
}

function openCreate() {
    form.value = emptyForm()
    alipayFileList.value = []
    wechatFileList.value = []
    dialogVisible.value = true
}

function openEdit(row: RegistrationPayQrItem) {
    form.value = { ...row }
    syncFileList('alipay')
    syncFileList('wechat')
    dialogVisible.value = true
}

function saveItem() {
    if (form.value.amount_yuan === undefined || form.value.amount_yuan === null || Number.isNaN(Number(form.value.amount_yuan))) {
        ElMessage.warning('请填写金额')
        return
    }
    saving.value = true
    const payload = { ...form.value }
    const req = payload.id
        ? registrationPayQrApi.update(payload)
        : registrationPayQrApi.create(payload)
    req.then((res: any) => {
        ElMessage.success(res.message || '已保存')
        dialogVisible.value = false
        load()
    }).finally(() => {
        saving.value = false
    })
}

function setCurrent(row: RegistrationPayQrItem) {
    settingId.value = row.id
    systemApi.save({ registration_pay_amount_yuan: row.amount_yuan })
        .then((res: any) => {
            ElMessage.success(res.message || '已设为当前应付金额')
            currentAmount.value = row.amount_yuan
        })
        .finally(() => {
            settingId.value = 0
        })
}

function remove(row: RegistrationPayQrItem) {
    ElMessageBox.confirm(`删除金额档 ${formatRegistrationPayAmount(row.amount_yuan)}？`, '删除确认', {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消',
    }).then(() => {
        registrationPayQrApi.remove(row.id).then((res: any) => {
            ElMessage.success(res.message || '已删除')
            if (form.value.id === row.id) dialogVisible.value = false
            load()
        })
    }).catch(() => {})
}

function beforeUploadFile(file: UploadRawFile) {
    if (!/image\/.*/.test(file.type)) {
        ElMessage.warning('请选择图片文件')
        return false
    }
    if (file.size > 10 * 1024 * 1024) {
        ElMessage.warning('图片大小请控制在 10MB 以内')
        return false
    }
    return true
}

function qiniuPublicUrl(key: string) {
    const info = BucketQiniuMap.get(PAY_QR_BUCKET)
    const base = String(info?.baseUrl || '').replace(/\/+$/, '').replace(/^http:/, 'https:')
    return `${base}/${key}`
}

function uploadToQiniu(platform: PayQrPlatform, uploadInfo: UploadRequestOptions) {
    imageQiniuApi
        .getUploadToken({ bucket: PAY_QR_BUCKET })
        .then((res: any) => {
            const observer = {
                error: () => {
                    ElMessage.error('上传失败')
                    uploadInfo.onError?.(new Error('上传失败') as any)
                },
                complete: (result: { key: string }) => {
                    const url = qiniuPublicUrl(result.key)
                    if (platform === 'alipay') {
                        form.value.alipay_url = url
                        alipayFileList.value = [{ name: result.key, url }]
                    } else {
                        form.value.wechat_url = url
                        wechatFileList.value = [{ name: result.key, url }]
                    }
                    imageQiniuApi.add({
                        id: result.key,
                        description: `注册收款码-${platform === 'alipay' ? '支付宝' : '微信'}`,
                        type: 'image',
                        bucket: PAY_QR_BUCKET,
                        base_url: BucketQiniuMap.get(PAY_QR_BUCKET)?.baseUrl || '',
                    } as any).catch(() => {})
                    uploadInfo.onSuccess?.(result as any)
                    ElMessage.success('收款码已上传')
                },
            }
            qiniu.upload(uploadInfo.file as File, null, res.data, {}, {}).subscribe(observer)
        })
        .catch(() => {
            ElMessage.error('获取上传凭证失败')
            uploadInfo.onError?.(new Error('获取上传凭证失败') as any)
        })
}

onMounted(load)
</script>

<style scoped lang="scss">
.amount-cell {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
}
.qr-cell {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    img {
        width: 72px;
        border-radius: 6px;
        border: 1px solid #ebeef5;
        background: #fff;
    }
}
.qr-cell__note {
    margin: 0;
    font-size: 12px;
    line-height: 1.5;
    color: #606266;
    white-space: pre-wrap;
}
.qr-edit {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    width: 100%;
}
.qr-edit__fields {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
}
.hint {
    margin: 4px 0 0;
    color: #909399;
    font-size: 12px;
    line-height: 1.4;
}
.muted {
    color: #c0c4cc;
    font-size: 12px;
}
:deep(.el-upload--picture-card) {
    width: 96px;
    height: 96px;
}
:deep(.el-upload-list--picture-card .el-upload-list__item) {
    width: 96px;
    height: 96px;
}
</style>
