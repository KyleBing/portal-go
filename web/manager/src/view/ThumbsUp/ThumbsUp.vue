<template>
    <Container>
        <Toolbar>
            <template #left>
            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton type="success" @click="addNewDiaryThumbsUp">添加</ElButton>
            </template>
        </Toolbar>

        <ElRow :gutter="10">
            <ElCol :span="24">
                <ElTable size="small" stripe :data="thumbsUpList" v-loading="isLoading">
                    <ElTableColumn align="left" prop="name" width="120" label="标识"/>
                    <ElTableColumn align="right" prop="count" width="100" label="数量"/>
                    <ElTableColumn align="left" prop="description" label="说明"/>
                    <ElTableColumn align="left" prop="link_address" label="网址"/>
                    <ElTableColumn align="right" prop="date_init" label="添加时间">
                        <template #default="{row}">
                            {{Moment(row.date_init).format('YYYY-MM-DD HH:mm:ss')}}
                        </template>
                    </ElTableColumn>

                    <ElTableColumn align="center" fixed="right" label="操作" width="200px">
                        <template #default="scope">
                            <ElButton @click="goEdit(scope.row)" type="primary" plain size="small">编辑</ElButton>
                            <ElButton @click="goDelete(scope.row)" type="danger" plain size="small">删除</ElButton>
                        </template>
                    </ElTableColumn>
                </ElTable>
            </ElCol>
        </ElRow>

        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="600px"
            :before-close="closeModal">
            <ElForm
                label-position="right"
                label-width="120px"
                size="small"
                :model="formThumbsUp"
                :rules="thumbsUpRules"
                ref="thumbsUpModify">
                <ElRow>
                    <ElCol>
                        <ElFormItem label="标识名" prop="name">
                            <ElInput autocomplete="off" :disabled="!isAdmin" v-model="formThumbsUp.name"></ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol>
                        <ElFormItem label="点赞数量" prop="count">
                            <ElInput autocomplete="off" :disabled="!isAdmin" v-model="formThumbsUp.count"></ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol>
                        <ElFormItem label="说明" prop="description">
                            <ElInput autocomplete="off" :disabled="!isAdmin" v-model="formThumbsUp.description"></ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol>
                        <ElFormItem label="网址" prop="link_address">
                            <ElInput autocomplete="off" :disabled="!isAdmin" v-model="formThumbsUp.link_address"></ElInput>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="clearForm" type="warning">清空</ElButton>
                <ElButton size="small" @click="closeModal" icon="close">取 消</ElButton>
                <ElButton size="small" type="primary" @click="submit" icon="check">{{ editingThumbsUp ? '修改' : '添加' }}</ElButton>
            </div>
        </ElDialog>

    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import thumbsUpApi from "@/api/thumbsUpApi"
import Container from "@/layout/Container.vue"
import Toolbar from "@/layout/Toolbar.vue"
import Moment from "moment/moment";

interface ThumbsUp {
    name: string
    count: number
    description: string
    link_address: string
    date_init: string
}

const isLoading = ref(false)
const thumbsUpList = ref<ThumbsUp[]>([])
const editingThumbsUp = ref<string | null>(null)
const modalEdit = ref(false)
const thumbsUpModify = ref<FormInstance>()
const isAdmin = ref(false)

const formThumbsUp = ref<ThumbsUp>({
    name: '',
    count: 0,
    description: '',
    link_address: '',
    date_init: ''
})

const thumbsUpRules = {
    name: { required: true, message: '请填写标识名', trigger: 'blur' },
    count: { required: true, message: '请填写初始点赞数量', trigger: 'blur' },
    description: { required: false, message: '请填写类别颜色', trigger: 'blur' },
    link_address: { required: true, message: '请填写站点地址', trigger: 'blur' }
}

const modalTitle = computed(() => editingThumbsUp.value ? '编辑类别' : '新增类别')

function addNewDiaryThumbsUp() {
    modalEdit.value = true
    editingThumbsUp.value = null
    clearForm()
}

function clearForm() {
    formThumbsUp.value = {
        name: '',
        count: 0,
        description: '',
        link_address: '',
        date_init: ''
    }
}

function closeModal(done?: () => void) {
    ElMessageBox.confirm('确认关闭？')
        .then(() => {
            editingThumbsUp.value = null
            modalEdit.value = false
            if (done) done()
        })
        .catch(() => {})
}

function getThumbsUpList() {
    isLoading.value = true
    const params = {
        pageNo: 1,
        pageSize: 50
    }
    thumbsUpApi
        .list(params)
        .then(res => {
            isLoading.value = false
            thumbsUpList.value = res.data
        })
        .catch(err => {
            isLoading.value = false
        })
}

function goEdit(thumbsUp: ThumbsUp) {
    editingThumbsUp.value = thumbsUp.name
    formThumbsUp.value = thumbsUp
    modalEdit.value = true
}

function goDelete(thumbsUp: ThumbsUp) {
    ElMessageBox.confirm(`删除类别 ${thumbsUp.name}(${thumbsUp.name})`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        const requestData = {
            name: thumbsUp.name
        }
        thumbsUpApi.delete(requestData)
            .then(res => {
                getThumbsUpList()
                ElMessage({
                    message: res.message,
                    type: 'success'
                })
            })
    })
}

function submit() {
    if (!thumbsUpModify.value) return
    thumbsUpModify.value.validate((valid) => {
        if (valid) {
            if (editingThumbsUp.value) {
                thumbsUpModifySubmit()
            } else {
                thumbsUpNewSubmit()
            }
        } else {
            console.log('error submit!!')
            return false
        }
    })
}

function thumbsUpNewSubmit() {
    thumbsUpApi
        .add(formThumbsUp.value)
        .then(res => {
            ElMessage({
                message: res.message,
                type: 'success'
            })
            getThumbsUpList()
            editingThumbsUp.value = null
            modalEdit.value = false
        })
}

function thumbsUpModifySubmit() {
    thumbsUpApi
        .modify(formThumbsUp.value)
        .then(res => {
            ElMessage({
                message: res.message,
                type: 'success'
            })
            getThumbsUpList()
            editingThumbsUp.value = null
            modalEdit.value = false
        })
}

onMounted(() => {
    getThumbsUpList()
    isAdmin.value = localStorage.getItem('email') === 'kylebing@163.com'
})
</script>

<style scoped lang="scss">
.table-description {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: pointer;
}
</style>
