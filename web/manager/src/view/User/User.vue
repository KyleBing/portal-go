<template>
    <Container>
        <Toolbar>
            <template #left>

            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton type="success" icon="Plus" @click="addNewUser()"> 添加</ElButton>
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
                        <ElTableColumn width="60" prop="uid" label="UID"/>
                        <ElTableColumn width="200" prop="email" label="Email"/>
                        <ElTableColumn width="150" prop="nickname" label="昵称"/>
                        <ElTableColumn width="100" align="center" prop="username" label="用户名"/>
                        <ElTableColumn width="150" align="center" prop="phone" label="手机号"/>
                        <ElTableColumn width="100" align="center" prop="wx" label="微信"/>
<!--                        <ElTableColumn align="center" width="100" prop="homepage" label="主页"/>-->
<!--                        <ElTableColumn align="right" prop="gaode" label="高德组队码"/>-->
                        <ElTableColumn sortable align="right" width="60" prop="count_diary" label="日记"/>
                        <ElTableColumn sortable align="right" width="60" prop="count_dict" label="码表"/>
                        <ElTableColumn sortable align="right" width="60" prop="count_qr" label="二维码"/>
                        <ElTableColumn sortable align="right" width="60" prop="count_words" label="词条"/>
                        <ElTableColumn sortable align="right" width="100" prop="sync_count" label="同步次数"/>
                        <ElTableColumn align="center" width="160" prop="register_time" label="注册时间">
                            <template #default="{row}">
                                {{dayjs(row.register_time).format('YYYY-MM-DD HH:mm:ss')}}
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="center" width="160" prop="last_visit_time" label="最后访问时间">
                            <template #default="{row}">
                                {{dayjs(row.register_time).format('YYYY-MM-DD HH:mm:ss')}}
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="right" width="60" prop="group_id" label="组别">
                            <template #default="scope">
                                {{ scope.row.group_id === 1 ? '管理员' : '普通' }}
                            </template>
                        </ElTableColumn>
                        <ElTableColumn align="center" width="200" label="操作">
                            <template #default="scope">
                                <ElButton @click="goEdit(scope.row)" type="primary" icon="edit" plain size="small">编辑</ElButton>
                                <ElButton @click="goDelete(scope.row)" type="danger" icon="delete" plain size="small">删除</ElButton>
                            </template>
                        </ElTableColumn>
                    </ElTable>

                </ElCol>
            </ElRow>
        </Content>
        <!--  PAGINATION  -->
        <FooterPagination
            :pager-option="pager"
            @size-change="pageSizeChange"
            @pager-change="pageNoChange"
        />

        <!-- 编辑窗口-->
        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="30%"
            :before-close="closeModal">
            <ElForm
                :model="formUser"
                :rules="userRules"
                size="small"
                ref="register" label-width="100px">
                <ElFormItem label="邮箱" prop="email">
                    <ElInput :disabled="!isAdmin" autocomplete="off" v-model="formUser.email"/>
                </ElFormItem>
                <ElFormItem label="用户名" prop="username">
                    <ElInput :disabled="!isAdmin" autocomplete="off" v-model="formUser.username"/>
                </ElFormItem>
                <ElFormItem label="昵称" prop="nickname">
                    <ElInput autocomplete="off" v-model="formUser.nickname"/>
                </ElFormItem>
                <ElFormItem label="微信" prop="wx">
                    <ElInput autocomplete="off" v-model="formUser.wx"/>
                </ElFormItem>
                <ElFormItem label="手机" prop="phone">
                    <ElInput autocomplete="off" v-model="formUser.phone"/>
                </ElFormItem>
                <ElFormItem label="高德组队码" prop="gaode">
                    <ElInput autocomplete="off" v-model="formUser.gaode"/>
                </ElFormItem>
                <ElFormItem label="组别" prop="group_id">
                    <ElSelect :disabled="!isAdmin" v-model="formUser.group_id" placeholder="请选择">
                        <ElOption
                            v-for="item in groupOptions"
                            :key="item.id"
                            :label="item.name"
                            :value="item.id">
                        </ElOption>
                    </ElSelect>
                </ElFormItem>
                <ElFormItem label="主页" prop="homepage">
                    <ElInput autocomplete="off" v-model="formUser.homepage"/>
                </ElFormItem>
            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="clearForm" type="warning">清空</ElButton>
                <ElButton size="small" @click="closeModal">取 消</ElButton>
                <ElButton size="small" type="primary" @click="submit">{{ editingUid ? '修改' : '添加' }}</ElButton>
            </div>
        </ElDialog>
    </Container>>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import userApi from "@/api/userApi"
import { useProjectStore } from "@/pinia"
import Container from "@/layout/Container.vue";
import Toolbar from "@/layout/Toolbar.vue";
import FooterPagination from "@/layout/FooterPagination.vue";
import TableListDate from "@/components/TableListDate.vue";
import Content from "@/layout/Content.vue";
import dayjs from 'dayjs';

interface User {
    uid: string
    email: string
    username: string
    nickname: string
    wx: string
    phone: string
    homepage: string
    gaode: string
    group_id: number
    count_diary: number
    count_dict: number
    count_qr: number
    count_words: number
    sync_count: number
    register_time: string
    last_visit_time: string
}

interface GroupOption {
    id: number
    name: string
}

interface Pager {
    total: number,
    pageNo: number,
    pageSize: number
}

const projectStore = useProjectStore()
const isLoading = ref(false)
const tableData = ref<User[]>([])
const editingUid = ref<string | null>(null)
const modalEdit = ref(false)
const isAdmin = ref(false)
const register = ref<FormInstance>()

const groupOptions: GroupOption[] = [
    { id: 1, name: '管理员' },
    { id: 2, name: '普通用户' }
]

const formUser = ref<User>({
    uid: '',
    email: '',
    username: '',
    nickname: '',
    wx: '',
    phone: '',
    homepage: '',
    gaode: '',
    group_id: 2,
    count_diary: 0,
    count_dict: 0,
    count_qr: 0,
    count_words: 0,
    sync_count: 0,
    register_time: '',
    last_visit_time: ''
})

const userRules = {
    email: [
        { required: true, message: '请填写邮箱', trigger: 'blur' }
    ],
    username: [
        { required: true, message: '请填写用户名', trigger: 'blur' },
        {
            validator: (rule: any, value: string, callback: Function) => {
                if (!/^[a-z_]+$/.test(value)) {
                    callback(new Error("用户名只能是小写字母"))
                } else {
                    callback()
                }
            }
        }
    ],
    nickname: { required: true, message: '请填写昵称', trigger: 'blur' },
    wx: { required: true, message: '请填写微信', trigger: 'blur' },
    phone: { required: true, message: '请填写手机号', trigger: 'blur' },
    homepage: '',
    gaode: '',
    group_id: { required: true, message: '请选择组别', trigger: 'blur' }
}

const pager = ref<Pager>({
    pageSize: 30,
    pageNo: 1,
    total: 0
})

const modalTitle = computed(() => editingUid.value ? '编辑用户' : '新增用户')

function addNewUser() {
    modalEdit.value = true
    editingUid.value = null
    clearForm()
}

function clearForm() {
    formUser.value = {
        uid: '',
        email: '',
        username: '',
        nickname: '',
        wx: '',
        phone: '',
        homepage: '',
        gaode: '',
        group_id: 2,
        count_diary: 0,
        count_dict: 0,
        count_qr: 0,
        count_words: 0,
        sync_count: 0,
        register_time: '',
        last_visit_time: ''
    }
}

function closeModal(done?: () => void) {
    ElMessageBox.confirm('确认关闭？')
        .then(() => {
            editingUid.value = null
            modalEdit.value = false
            if (done) done()
        })
        .catch(() => {})
}

function pageSizeChange(pageSize: number){
    pager.value.pageSize = pageSize
    getUserList()
}

function pageNoChange(pageNo: number){
    pager.value.pageNo = pageNo
    getUserList()
}
function getUserList() {
    isLoading.value = true
    const params = {
        pageNo: pager.value.pageNo,
        pageSize: pager.value.pageSize
    }
    userApi.list(params)
        .then(res => {
            isLoading.value = false
            tableData.value = res.data.list
            pager.value.total = res.data.pager.total
        })
        .catch(err => {
            isLoading.value = false
        })
}

function goEdit(user: User) {
    editingUid.value = user.uid
    formUser.value = { ...user }
    modalEdit.value = true
}

function goDelete(user: User) {
    ElMessageBox.confirm(`删除用户 ${user.nickname}(${user.username})`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        const requestData = {
            uid: user.uid
        }
        userApi.delete(requestData)
            .then(res => {
                getUserList()
                ElMessage({
                    message: res.message,
                    type: 'success'
                })
            })
    })
}

function submit() {
    if (!register.value) return
    register.value.validate((valid) => {
        if (valid) {
            if (editingUid.value) {
                userModifySubmit()
            } else {
                userNewSubmit()
            }
        } else {
            console.log('error submit!!')
            return false
        }
    })
}

function userNewSubmit() {
    userApi.add(formUser.value)
        .then(res => {
            ElMessage({
                message: res.message,
                type: 'success'
            })
            getUserList()
            editingUid.value = null
            modalEdit.value = false
        })
}

function userModifySubmit() {
    userApi.modify(formUser.value)
        .then(res => {
            ElMessage({
                message: res.message,
                type: 'success'
            })
            getUserList()
            editingUid.value = null
            modalEdit.value = false
        })
}

onMounted(() => {
    getUserList()
    isAdmin.value = localStorage.getItem('email') === 'kylebing@163.com'
})
</script>

<style scoped lang="scss">
.user-list {
    height: 100%;
    overflow: auto;
}
</style>
