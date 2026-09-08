<template>
    <Container>
        <Toolbar>
            <template #left>

            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton
                    type="success" icon="Plus" @click="addNewDiaryCategory()"> 添加</ElButton>
            </template>
        </Toolbar>
        <Content padding="0">
            <ElRow :gutter="10">
                <ElCol :span="24">
                    <ElTable size="small" stripe :data="categoryList" v-loading="isLoading">
                        <ElTableColumn width="120" align="left" prop="name_en" label="英文标识"/>
                        <ElTableColumn width="100" align="center" prop="name" label="类别"/>
                        <ElTableColumn align="center" prop="color" width="50px" label="颜色">
                            <template #default="{ row }">
                                <div class="color-square" :style="`width:20px; height:20px; background-color: ${row.color}`"></div>
                            </template>
                        </ElTableColumn>
                        <ElTableColumn width="100" align="right" prop="count" label="数量"/>
                        <ElTableColumn width="100" align="right" prop="sort_id" label="排列序号"/>
                        <ElTableColumn sortable align="left" width="200" prop="date_init" label="添加时间">
                            <template #default="scope">
                                <TableListDate :dates="[scope.row.date_init]" :names="['创建']"/>
                            </template>
                        </ElTableColumn>

                        <ElTableColumn align="left" fixed="right" label="操作" width="">
                            <template #default="{ row }">
                                <ElButton @click="goEdit(row)" type="primary" plain icon="Edit" size="small"> 编辑</ElButton>
                                <ElButton @click="goDelete(row)" type="danger" plain icon="Delete" size="small"> 删除</ElButton>
                            </template>
                        </ElTableColumn>
                    </ElTable>

                </ElCol>
            </ElRow>
        </Content>

        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="600px"
            :before-close="closeModal">
            <ElForm
                label-position="right"
                label-width="120px"
                size="small"
                :model="formCategory"
                :rules="categoryRules"
                ref="categoryModify">
                <ElRow>
                    <ElCol>
                        <ElFormItem label="类别名" prop="name">
                            <ElInput autocomplete="off" :disabled="!isAdmin" v-model="formCategory.name"></ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol>
                        <ElFormItem label="英文标识" prop="name_en">
                            <ElInput autocomplete="off" :disabled="!isAdmin" v-model="formCategory.name_en"></ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol>
                        <ElFormItem label="类别颜色" prop="color">
                            <ElColorPicker :disabled="!isAdmin" v-model="formCategory.color"></ElColorPicker>
                        </ElFormItem>
                    </ElCol>
                    <ElCol>
                        <ElFormItem label="排序序号" prop="sort_id">
                            <ElInput type="number" autocomplete="off" :disabled="!isAdmin" v-model="formCategory.sort_id"></ElInput>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

            </ElForm>
            <template #footer>
                <div class="dialog-footer">
                    <ElButton size="small" @click="clearForm" type="warning">
                        <ElIcon><Delete /></ElIcon>
                        清空
                    </ElButton>
                    <ElButton size="small" @click="closeModal">
                        <ElIcon><Close /></ElIcon>
                        取 消
                    </ElButton>
                    <ElButton size="small" type="primary" @click="submit">
                        <ElIcon><Check /></ElIcon>
                        {{ editingCategory ? ' 修改' : ' 添加' }}
                    </ElButton>
                </div>
            </template>
        </ElDialog>

    </Container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessageBox, ElNotification } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import diaryCategoryApi from "@/api/diaryCategoryApi"
import { getAuthorization } from "@/utility"
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import TableListDate from "@/components/TableListDate.vue";
import Content from "@/layout/Content.vue";

interface CategoryItem {
    name: string
    name_en: string
    color: string
    sort_id: number
    date_init: string
}

interface ApiResponse {
    message: string
    data: CategoryItem[]
}

interface BaseResponse {
    message: string
    data?: any
}

const isLoading = ref(false)
const categoryList = ref<CategoryItem[]>([])
const editingCategory = ref<string | null>(null)
const modalEdit = ref(false)
const categoryModify = ref<FormInstance>()
const isAdmin = ref(false)

interface CategoryForm {
    name: string
    name_en: string
    color: string
    sort_id: number
}

const formCategory = ref<CategoryForm>({
    name: '',
    name_en: '',
    color: '',
    sort_id: 0
})

const categoryRules: FormRules = {
    name: {required: true, message: '请填写名字', trigger: 'blur'},
    name_en: {required: true, message: '请填写英文名标识', trigger: 'blur'},
    color: {required: true, message: '请填写类别颜色', trigger: 'blur'},
    sort_id: {required: true, message: '请填写排序序号', trigger: 'blur'},
}

const modalTitle = computed(() => editingCategory.value ? '编辑类别' : '新增类别')

const addNewDiaryCategory = () => {
    modalEdit.value = true
    editingCategory.value = null
    clearForm()
}

const clearForm = () => {
    formCategory.value = {
        name: '',
        name_en: '',
        color: '#ccc',
        sort_id: 0
    }
}

const closeModal = () => {
    editingCategory.value = null
    modalEdit.value = false
}

const getCategoryList = () => {
    isLoading.value = true
    const params = {
        pageNo: 1,
        pageSize: 50
    }
    diaryCategoryApi.list(params)
        .then((res: unknown) => {
            const response = res as ApiResponse
            isLoading.value = false
            categoryList.value = response.data
        })
        .catch(() => {
            isLoading.value = false
        })
}

const goEdit = (category: CategoryItem) => {
    editingCategory.value = category.name_en
    formCategory.value = category
    modalEdit.value = true
}

const goDelete = (category: CategoryItem) => {
    ElMessageBox.confirm(`删除类别 ${category.name}(${category.name_en})`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        const requestData = {
            id: category.sort_id // Using sort_id as id since it seems to be the identifier
        }
        diaryCategoryApi.delete(requestData)
            .then((res: unknown) => {
                const response = res as BaseResponse
                getCategoryList()
                ElNotification({
                    title: response.message,
                    position: 'top-right',
                    type: 'success'
                })
            })
    })
}

const submit = () => {
    if (!categoryModify.value) return
    categoryModify.value.validate((valid) => {
        if (valid) {
            if (editingCategory.value) {
                categoryModifySubmit()
            } else {
                categoryNewSubmit()
            }
        } else {
            console.log('error submit!!')
        }
    })
}

const categoryNewSubmit = () => {
    diaryCategoryApi
        .add(formCategory.value)
        .then((res: unknown) => {
            const response = res as BaseResponse
            ElNotification({
                title: response.message,
                position: 'top-right',
                type: 'success'
            })
            getCategoryList()
            editingCategory.value = null
            modalEdit.value = false
        })
}

const categoryModifySubmit = () => {
    diaryCategoryApi
        .modify(formCategory.value)
        .then((res: unknown) => {
            const response = res as BaseResponse
            ElNotification({
                title: response.message,
                position: 'top-right',
                type: 'success'
            })
            getCategoryList()
            editingCategory.value = null
            modalEdit.value = false
        })
}

onMounted(() => {
    getCategoryList()
    const auth = getAuthorization()
    isAdmin.value = auth?.email === 'kylebing@163.com'
})
</script>

<style scoped lang="scss">
.table-description{
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: pointer;
}

.dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
}

:deep(.el-button) {
    display: inline-flex;
    align-items: center;
    gap: 4px;
}
</style>
