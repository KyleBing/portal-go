<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElForm inline>
                    <ElFormItem label="选择文件">
                        <input type="file" ref="inputUpload" @change="inputFileChange"/>
                    </ElFormItem>
                    <ElFormItem label="文件描述">
                        <ElInput type="text" v-model="fileDescription"/>
                    </ElFormItem>
                    <ElFormItem>
                        <ElButton icon="Upload" type="success" :loading="uploading" @click="uploadFile">上传</ElButton>
                    </ElFormItem>
                </ElForm>
            </template>
        </Toolbar>

        <Content padding="0" :is-show-pagination="true">
            <ElTable
                class="table-narrow"
                size="small"
                :height="projectStore.contentInsets.heightContent - 130"
                stripe
                :data="tableData"
                v-loading="isLoading"
            >
                <ElTableColumn prop="id" label="ID" width="70"/>
                <ElTableColumn prop="name_original" label="原文件名" min-width="180"/>
                <ElTableColumn prop="description" label="描述" min-width="140"/>
                <ElTableColumn prop="type" width="140" label="类型"/>
                <ElTableColumn prop="path" label="路径" min-width="160" show-overflow-tooltip/>
                <ElTableColumn align="right" width="100" prop="size" label="大小">
                    <template #default="{ row }">
                        <span>{{ formatSize(row.size) }}</span>
                    </template>
                </ElTableColumn>
                <ElTableColumn sortable align="center" width="170" prop="date_create" label="添加时间">
                    <template #default="{ row }">
                        <TableListDate :dates="[row.date_create]" :names="['创建']"/>
                    </template>
                </ElTableColumn>
                <ElTableColumn align="center" width="260" label="操作" fixed="right">
                    <template #default="{ row }">
                        <ElButton size="small" plain type="primary" @click="download(row)">下载</ElButton>
                        <ElButton size="small" plain type="danger" icon="Delete" @click="goDelete(row)">删除</ElButton>
                    </template>
                </ElTableColumn>
            </ElTable>
        </Content>
        <FooterPagination
            :pager-option="pager"
            @size-change="getFileList"
            @pager-change="getFileList"
        />
    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import fileManagerApi from '@/api/fileManagerApi'
import { useProjectStore } from '@/pinia'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import FooterPagination from '@/layout/FooterPagination.vue'
import TableListDate from '@/components/TableListDate.vue'
import Content from '@/layout/Content.vue'
import { getAuthorization } from '@/utility'
import { downloadWithAuth } from '@/utils/webrtcTransfer'

interface FileInfo {
    id: number
    name_original: string
    description: string
    type: string
    path: string
    size: number
    date_create: string
    download_url?: string
}

interface Pager {
    total: number
    pageNo: number
    pageSize: number
}

const projectStore = useProjectStore()
const isLoading = ref(false)
const uploading = ref(false)
const tableData = ref<FileInfo[]>([])
const currentFile = ref<File | null>(null)
const fileDescription = ref('')
const inputUpload = ref<HTMLInputElement | null>(null)
const pager = ref<Pager>({ pageSize: 30, pageNo: 1, total: 0 })

function formatSize(n: number) {
    if (!n) return '0 KB'
    return `${(n / 1024).toFixed(0)} KB`
}

function inputFileChange(event: Event) {
    const target = event.target as HTMLInputElement
    currentFile.value = target.files?.[0] || null
}

function uploadFile() {
    if (!currentFile.value) {
        ElMessage.warning('未选择文件')
        return
    }
    if (!fileDescription.value) {
        ElMessage.warning('未填写文件描述')
        return
    }
    const fd = new FormData()
    fd.append('file', currentFile.value)
    fd.append('note', fileDescription.value)
    uploading.value = true
    fileManagerApi.upload(fd)
        .then((res: any) => {
            ElMessage.success(res.message || '上传成功')
            currentFile.value = null
            fileDescription.value = ''
            if (inputUpload.value) inputUpload.value.value = ''
            getFileList()
        })
        .finally(() => { uploading.value = false })
}

function getFileList() {
    isLoading.value = true
    fileManagerApi.list({ pageNo: pager.value.pageNo, pageSize: pager.value.pageSize })
        .then((res: any) => {
            tableData.value = res.data || []
        })
        .finally(() => { isLoading.value = false })
}

async function download(row: FileInfo) {
    const auth = getAuthorization()
    if (!auth?.token || !auth?.uid) {
        ElMessage.error('未登录')
        return
    }
    const url = row.download_url || `/portal/file-manager/download?fileId=${row.id}`
    try {
        await downloadWithAuth(url, auth.token, auth.uid, row.name_original || 'file')
    } catch {
        ElMessage.error('下载失败')
    }
}

function goDelete(fileInfo: FileInfo) {
    ElMessageBox.confirm(`删除记录 ${fileInfo.name_original}`, '删除', { type: 'warning' })
        .then(() => fileManagerApi.delete({ fileId: fileInfo.id }))
        .then(() => {
            ElMessage.success('删除成功')
            getFileList()
        })
        .catch(() => {})
}

onMounted(getFileList)
</script>
