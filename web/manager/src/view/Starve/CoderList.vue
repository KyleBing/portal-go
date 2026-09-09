<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElForm size="small" inline>
                    <ElFormItem label="搜索">
                        <ElInput 
                            clearable 
                            placeholder="搜索用途说明" 
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
                <ElButton type="success" icon="Plus" @click="addNewCoder()"> 添加</ElButton>
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
                        <ElTableColumn width="60" prop="id" label="ID" fixed="left"/>
                        <ElTableColumn width="100" prop="is_active" label="状态">
                            <template #default="{row}">
                                <ElTag :type="row.is_active === 1 ? 'success' : 'info'">
                                    {{ row.is_active === 1 ? '启用' : '禁用' }}
                                </ElTag>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn width="200" prop="usage" label="用途说明"/>
                        <ElTableColumn prop="code" label="代码内容" show-overflow-tooltip/>
                        <ElTableColumn prop="note" label="备注" show-overflow-tooltip/>
                        <ElTableColumn align="center" width="200" label="操作" fixed="right">
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
            width="60%"
            :before-close="closeModal">
            <ElForm
                :model="formCoder"
                :rules="coderRules"
                size="small"
                ref="formRef" label-width="120px">
                <ElFormItem label="用途说明" prop="usage">
                    <ElInput v-model="formCoder.usage" placeholder="请输入用途说明"/>
                </ElFormItem>
                <ElFormItem label="代码内容" prop="code">
                    <ElInput type="textarea" :rows="10" v-model="formCoder.code" placeholder="请输入代码内容"/>
                </ElFormItem>
                <ElFormItem label="备注" prop="note">
                    <ElInput type="textarea" :rows="4" v-model="formCoder.note" placeholder="请输入备注"/>
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
import { Coder } from "@/model/starve";

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<Coder[]>([])
const allData = ref<Coder[]>([])
const filteredData = ref<Coder[]>([])
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

const formCoder = ref<Coder>({
    id: undefined,
    usage: '',
    code: '',
    note: null
})

const coderRules = {
    usage: [
        { required: true, message: '请输入用途说明', trigger: 'blur' }
    ],
    code: [
        { required: true, message: '请输入代码内容', trigger: 'blur' }
    ]
}

const modalTitle = computed(() => editingId.value ? '编辑代码' : '新增代码')

function addNewCoder() {
    modalEdit.value = true
    editingId.value = null
    clearForm()
}

function clearForm() {
    formCoder.value = {
        id: undefined,
        usage: '',
        code: '',
        note: null
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

function getCoderList() {
    isLoading.value = true
    starveApi.coderList()
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
            const usage = (item.usage || '').toLowerCase()
            return usage.includes(searchKey)
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

function goEdit(coder: Coder) {
    editingId.value = coder.id || null
    formCoder.value = { ...coder }
    modalEdit.value = true
}

function goDelete(coder: Coder) {
    ElMessageBox.confirm(`删除代码 ${coder.usage}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        if (!coder.id) return
        starveApi.coderDelete({ id: coder.id })
            .then(res => {
                getCoderList()
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
                coderModifySubmit()
            } else {
                coderAddSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

function coderAddSubmit() {
    starveApi.coderAdd(formCoder.value)
        .then(() => {
            ElMessage({
                message: '添加成功',
                type: 'success'
            })
            getCoderList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

function coderModifySubmit() {
    starveApi.coderModify(formCoder.value)
        .then(() => {
            ElMessage({
                message: '修改成功',
                type: 'success'
            })
            getCoderList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

onMounted(() => {
    getCoderList()
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

