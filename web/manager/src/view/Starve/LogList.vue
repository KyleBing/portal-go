<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElForm size="small" inline>
                    <ElFormItem label="搜索">
                        <ElInput 
                            clearable 
                            placeholder="搜索详情" 
                            v-model="keyword"
                            style="width: 200px"
                        />
                    </ElFormItem>
                    <ElFormItem>
                        <ElButton type="primary" @click="handleSearch" icon="Search">搜索</ElButton>
                    </ElFormItem>
                </ElForm>
            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton type="success" icon="Plus" @click="addNewLog()"> 添加</ElButton>
            </template>
        </Toolbar>

        <Content padding="0" :is-show-pagination="true">
            <ElRow :gutter="10">
                <ElCol :span="24">
                    <ElTable
                        class="table-narrow"
                        size="small"
                        :height="projectStore.contentInsets.heightContent - 130"
                        stripe
                        :data="filteredData"
                        v-loading="isLoading"
                    >
                        <ElTableColumn width="60" prop="id" label="ID"/>
                        <ElTableColumn width="180" prop="date" label="日期">
                            <template #default="{row}">
                                <span v-if="row.date">{{dayjs(row.date).format('YYYY-MM-DD HH:mm:ss')}}</span>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn prop="detail" label="详情" show-overflow-tooltip/>
                        <ElTableColumn align="center" width="200" label="操作">
                            <template #default="scope">
                                <ElButton @click="goEdit(scope.row)" type="primary" icon="edit" plain size="small">编辑</ElButton>
                                <ElButton @click="goDelete(scope.row)" type="danger" icon="delete" plain size="small">删除</ElButton>
                            </template>
                        </ElTableColumn>
                    </ElTable>
                </ElCol>
            </ElRow>
            <FooterPagination
                :pager-option="pager"
                @size-change="sizeChange"
                @pager-change="pagerChange"
            />
        </Content>

        <!-- 编辑窗口-->
        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="40%"
            :before-close="closeModal">
            <ElForm
                :model="formLog"
                :rules="logRules"
                size="small"
                ref="formRef" label-width="100px">
                <ElFormItem label="日期" prop="date">
                    <ElDatePicker
                        v-model="formLog.date"
                        type="date"
                        placeholder="选择日期"
                        format="YYYY-MM-DD"
                        value-format="YYYY-MM-DD"
                        style="width: 100%"
                    />
                </ElFormItem>
                <ElFormItem label="详情" prop="detail">
                    <ElInput type="textarea" :rows="4" v-model="formLog.detail" placeholder="请输入详情"/>
                </ElFormItem>
            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="clearForm" type="warning" icon="RefreshLeft">清空</ElButton>
                <ElButton size="small" @click="closeModal" icon="Close">取 消</ElButton>
                <ElButton size="small" type="primary" @click="submit" icon="Check">确定</ElButton>
            </div>
        </ElDialog>
    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import starveApi from "@/api/starveApi"
import { useProjectStore } from "@/pinia"
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import Content from "@/layout/Content.vue";
import FooterPagination from "@/layout/FooterPagination.vue";
import { Log } from "@/model/starve";
import dayjs from 'dayjs';

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<Log[]>([])
const allData = ref<Log[]>([])
const filteredData = ref<Log[]>([])
const keyword = ref('')
let searchTimer: number | null = null
const editingId = ref<number | null>(null)
const modalEdit = ref(false)
const formRef = ref<FormInstance>()

// 分页
const pager = ref({
    total: 0,
    pageNo: 1,
    pageSize: 20
})

const formLog = ref<Log>({
    id: undefined,
    date: '',
    detail: ''
})

const logRules = {
    date: [
        { required: true, message: '请选择日期', trigger: 'change' }
    ],
    detail: [
        { required: true, message: '请输入详情', trigger: 'blur' }
    ]
}

const modalTitle = computed(() => editingId.value ? '编辑日志' : '新增日志')

function addNewLog() {
    modalEdit.value = true
    editingId.value = null
    clearForm()
}

function clearForm() {
    formLog.value = {
        id: undefined,
        date: '',
        detail: ''
    }
}

function closeModal(done?: () => void) {
    ElMessageBox.confirm('确认关闭？')
        .then(() => {
            editingId.value = null
            modalEdit.value = false
            if (done) done()
        })
        .catch(() => {})
}

function getLogList() {
    isLoading.value = true
    starveApi.logList()
        .then((res: any) => {
            allData.value = Array.isArray(res) ? res : (res.data || [])
            tableData.value = allData.value
            handleSearch()
            isLoading.value = false
        })
        .catch(err => {
            isLoading.value = false
            console.error(err)
        })
}

function handleSearch() {
    let result = allData.value
    
    // 关键词搜索
    if (keyword.value && keyword.value.trim() !== '') {
        const searchKey = keyword.value.trim().toLowerCase()
        result = result.filter(item => {
            const detail = (item.detail || '').toLowerCase()
            return detail.includes(searchKey)
        })
    }
    
    // 更新总数
    pager.value.total = result.length
    
    // 分页
    const start = (pager.value.pageNo - 1) * pager.value.pageSize
    const end = start + pager.value.pageSize
    filteredData.value = result.slice(start, end)
}

function sizeChange(size: number) {
    pager.value.pageSize = size
    pager.value.pageNo = 1
    handleSearch()
}

function pagerChange(pageNo: number) {
    pager.value.pageNo = pageNo
    handleSearch()
}

function goEdit(log: Log) {
    editingId.value = log.id || null
    formLog.value = { ...log }
    modalEdit.value = true
}

function goDelete(log: Log) {
    ElMessageBox.confirm(`删除日志 ${log.date}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        if (!log.id) return
        starveApi.logDelete({ id: log.id })
            .then(res => {
                getLogList()
                ElMessage({
                    message: res.message || '删除成功',
                    type: 'success'
                })
            })
    })
}

function submit() {
    if (!formRef.value) return
    formRef.value.validate((valid) => {
        if (valid) {
            if (editingId.value) {
                logModifySubmit()
            } else {
                logAddSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

function logAddSubmit() {
    starveApi.logAdd(formLog.value)
        .then(() => {
            ElMessage({
                message: '添加成功',
                type: 'success'
            })
            getLogList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

function logModifySubmit() {
    starveApi.logModify(formLog.value)
        .then(() => {
            ElMessage({
                message: '修改成功',
                type: 'success'
            })
            getLogList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

onMounted(() => {
    getLogList()
})

watch(keyword, () => {
    if (searchTimer) {
        clearTimeout(searchTimer)
    }
    pager.value.pageNo = 1
    searchTimer = setTimeout(() => {
        handleSearch()
    }, 500)
})
</script>

<style scoped lang="scss">
</style>

