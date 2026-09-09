<template>
    <Container>
        <Toolbar>
            <template #left>
                <ElForm size="small" inline>
                    <ElFormItem label="搜索">
                        <ElInput 
                            clearable 
                            placeholder="搜索代码或名称" 
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
                <ElButton type="success" icon="Plus" @click="addNewCraftTab()"> 添加</ElButton>
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
                        <ElTableColumn width="120" prop="code" label="标签代码"/>
                        <ElTableColumn width="150" prop="name" label="标签名称"/>
                        <ElTableColumn width="150" prop="name_en" label="英文名称"/>
                        <ElTableColumn width="100" prop="sort_order" label="排序"/>
                        <ElTableColumn width="180" prop="created_at" label="创建时间">
                            <template #default="{row}">
                                <span v-if="row.created_at">{{formatTime(row.created_at)}}</span>
                                <span v-else>-</span>
                            </template>
                        </ElTableColumn>
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
            width="50%"
            :before-close="closeModal">
            <ElForm
                :model="formCraftTab"
                :rules="craftTabRules"
                size="small"
                ref="formRef" label-width="120px">
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="标签代码" prop="code">
                            <ElInput v-model="formCraftTab.code" placeholder="请输入标签代码"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="排序" prop="sort_order">
                            <ElInputNumber v-model="formCraftTab.sort_order" placeholder="请输入排序" style="width: 100%"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow :gutter="20">
                    <ElCol :span="12">
                        <ElFormItem label="标签名称" prop="name">
                            <ElInput v-model="formCraftTab.name" placeholder="请输入标签名称"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :span="12">
                        <ElFormItem label="英文名称" prop="name_en">
                            <ElInput v-model="formCraftTab.name_en" placeholder="请输入英文名称"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElFormItem label="状态" prop="is_active">
                    <ElRadioGroup v-model="formCraftTab.is_active">
                        <ElRadio :label="1">启用</ElRadio>
                        <ElRadio :label="0">禁用</ElRadio>
                    </ElRadioGroup>
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
import { CraftTab } from "@/model/starve";
import dayjs from 'dayjs';

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<CraftTab[]>([])
const allData = ref<CraftTab[]>([])
const filteredData = ref<CraftTab[]>([])
const keyword = ref('')
let searchTimer: number | null = null
const editingId = ref<number | null>(null)

// 分页
const pager = ref({
    total: 0,
    pageNo: 1,
    pageSize: 20
})
const modalEdit = ref(false)
const formRef = ref<FormInstance>()

const formCraftTab = ref<CraftTab>({
    id: undefined,
    code: '',
    name: '',
    name_en: null,
    sort_order: 0,
    is_active: 1
})

const craftTabRules = {
    code: [
        { required: true, message: '请输入标签代码', trigger: 'blur' }
    ],
    name: [
        { required: true, message: '请输入标签名称', trigger: 'blur' }
    ]
}

const modalTitle = computed(() => editingId.value ? '编辑制作标签' : '新增制作标签')

function addNewCraftTab() {
    modalEdit.value = true
    editingId.value = null
    clearForm()
}

function clearForm() {
    formCraftTab.value = {
        id: undefined,
        code: '',
        name: '',
        name_en: null,
        sort_order: 0,
        is_active: 1
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

function getCraftTabList() {
    isLoading.value = true
    starveApi.craftTabList()
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
            const code = (item.code || '').toLowerCase()
            const name = (item.name || '').toLowerCase()
            const nameEn = (item.name_en || '').toLowerCase()
            return code.includes(searchKey) || name.includes(searchKey) || nameEn.includes(searchKey)
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

function goEdit(craftTab: CraftTab) {
    editingId.value = craftTab.id || null
    formCraftTab.value = { ...craftTab }
    modalEdit.value = true
}

function goDelete(craftTab: CraftTab) {
    ElMessageBox.confirm(`删除制作标签 ${craftTab.name}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        if (!craftTab.id) return
        starveApi.craftTabDelete({ id: craftTab.id })
            .then(res => {
                getCraftTabList()
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
                craftTabModifySubmit()
            } else {
                craftTabAddSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

function craftTabAddSubmit() {
    starveApi.craftTabAdd(formCraftTab.value)
        .then(() => {
            ElMessage({
                message: '添加成功',
                type: 'success'
            })
            getCraftTabList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

function craftTabModifySubmit() {
    starveApi.craftTabModify(formCraftTab.value)
        .then(() => {
            ElMessage({
                message: '修改成功',
                type: 'success'
            })
            getCraftTabList()
            editingId.value = null
            modalEdit.value = false
        })
        .catch(err => {
            console.error(err)
        })
}

onMounted(() => {
    getCraftTabList()
})

function formatTime(time: string | null | undefined): string {
    if (!time) return '-'
    return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

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

