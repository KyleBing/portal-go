<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElForm inline>
                    <ElFormItem label="选择文件">
                        <input type="file" ref="inputUpload" @change="inputFileChange"/>
                    </ElFormItem>
                    <ElFormItem label="文件描述">
                        <ElInput type="text" v-model="fileDescription"></ElInput>
                    </ElFormItem>
                    <ElFormItem>
                        <ElButton icon="Upload" type="success" @click="uploadFile">确认添加</ElButton>
                    </ElFormItem>
                </ElForm>
            </template>
            <template #center>
            </template>
            <template #right>
            </template>
        </Toolbar>

        <Content padding="0">
            <ElRow :gutter="10">
                <ElCol :span="24">
                    <ElTable
                        class="table-narrow"
                        size="small"
                        :height="projectStore.contentInsets.heightContent - 130"
                        stripe
                        :data="tableData"
                        v-loading="isLoading"
                    >
                        <ElTableColumn prop="id" label="ID" width="80"/>
                        <ElTableColumn prop="name_original" label="原文件名" width="400"/>
                        <ElTableColumn prop="description" label="描述"/>
                        <ElTableColumn prop="type" width="180" label="文件类型"/>
                        <ElTableColumn prop="path" label="路径">
                            <template #default="scope">
                                <a target="_blank" :href="`${BASE_URL}${scope.row.name}`">{{scope.row.name}}</a>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="right" width="120" prop="size" label="文件大小">
                            <template #default="scope">
                                <span>{{(scope.row.size / 1024).toFixed(0)}}kb</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn sortable align="center" width="200" prop="date_create" label="添加时间">
                            <template #default="scope">
                                <TableListDate :dates="[scope.row.date_create]" :names="['创建']"/>
                            </template>
                        </ElTableColumn>

                        <ElTableColumn align="center" width="300" label="操作">
                            <template #default="scope">
                                <ElButton size="small" plain class="clipboard" :data-clipboard="`${BASE_URL}${scope.row.name}`" icon="CopyDocument" type="primary">复制文件地址</ElButton>
                                <ElButton size="small" plain @click="goDelete(scope.row)" type="danger" icon="delete">删除</ElButton>
                            </template>
                        </ElTableColumn>
                    </ElTable>

                </ElCol>
            </ElRow>
        </Content>
        <!--  PAGINATION  -->
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
import fileManagerApi from "@/api/fileManagerApi"
import ClipboardJS from 'clipboard'
import {useProjectStore} from "@/pinia";
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import { getAuthorization } from "@/utility";
import FooterPagination from "@/layout/FooterPagination.vue";
import TableListDate from "@/components/TableListDate.vue";
import Content from "@/layout/Content.vue";

interface FileInfo {
    id: number
    name_original: string
    description: string
    type: string
    name: string
    size: number
    date_create: string
}

interface Pager {
    total: number,
    pageNo: number,
    pageSize: number
}

const projectStore = useProjectStore()

const BASE_URL = 'http://kylebing.cn/'
const isLoading = ref(false)
const editingUid = ref<number | null>(null)
const tableData = ref<FileInfo[]>([])
const modalEdit = ref(false)
const isAdmin = ref(false)
const currentFile = ref<File | null>(null)
const fileDescription = ref('')
const inputUpload = ref<HTMLInputElement | null>(null)
const clipboard = ref<ClipboardJS | null>(null)

const pager = ref<Pager>({
    pageSize: 30,
    pageNo: 1,
    total: 0
})

const inputFileChange = (event: Event) => {
    const target = event.target as HTMLInputElement
    if (target.files && target.files.length > 0) {
        currentFile.value = target.files[0]
    }
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

    const requestData = new FormData()
    requestData.append('file', currentFile.value)
    requestData.append('note', fileDescription.value)

    fileManagerApi.upload(requestData)
        .then(res => {
            ElMessage.success(res.message)
            currentFile.value = null
            fileDescription.value = ''
            getFileList()
        })
        .catch(error => {
            console.error('Upload failed:', error)
        })
}

function getFileList() {
    isLoading.value = true
    const params = {
        pageNo: pager.value.pageNo,
        pageSize: pager.value.pageSize
    }
    fileManagerApi
        .list(params)
        .then(res => {
            tableData.value = res.data
            isLoading.value = false
        })
        .catch(error => {
            isLoading.value = false
            console.error('Failed to get file list:', error)
        })
}

const goDelete = (fileInfo: FileInfo) => {
    ElMessageBox.confirm(
        `删除记录 ${fileInfo.name_original}`,
        '删除',
        {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
        }
    ).then(() => {
        const requestData = {
            fileId: fileInfo.id
        }
        fileManagerApi.delete(requestData)
            .then(() => {
                ElMessage.success('删除成功')
                getFileList()
            })
            .catch(error => {
                console.error('Delete failed:', error)
            })
    }).catch(() => {
        // User cancelled the deletion
    })
}

onMounted(() => {
    getFileList()
    isAdmin.value = getAuthorization().email === 'kylebing@163.com'
    
    // Initialize clipboard
    clipboard.value = new ClipboardJS('.clipboard', {
        text: trigger => trigger.getAttribute('data-clipboard') || ''
    })
    
    clipboard.value.on('success', () => {
        ElMessage.success('复制成功')
    })
})
</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;

.tool-bar{}

.thumbnail{
    width: 50px;
    padding: 2px;
    @include border-radius(2px);
    border: 1px solid $color-border;
    img{
        display: block;
        width: 100%;
    }
}

:deep(.el-button) {
    margin-right: 8px;
    
    &:last-child {
        margin-right: 0;
    }
    
    &.clipboard {
        margin-left: 8px;
    }
}
</style>
