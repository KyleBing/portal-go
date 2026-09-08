<template>
    <Container>
        <Toolbar>
            <template #left>
            </template>
            <template #center>
            </template>
            <template #right>
                <ElButton type="success" @click="addNewCode" icon="Plus"> 添加</ElButton>
            </template>
        </Toolbar>
        <ElRow :gutter="10">
            <ElCol :span="24">
                <ElTable size="small" stripe :data="codeList" v-loading="isLoading">
                    <ElTableColumn align="center" prop="hash" label="HASH"/>
                    <ElTableColumn align="center" prop="is_public" label="状态" width="50px">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.is_public"><SuccessFilled/></ElIcon>
                            <ElIcon v-else><Minus/></ElIcon>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="username" label="所属"/>
                    <ElTableColumn align="center" prop="phone" min-width="100px" label="手机号"/>

                    <ElTableColumn align="center" prop="is_show_car" label="车状态" width="60px">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.is_show_car"><Check/></ElIcon>
                            <ElIcon v-else><Minus/></ElIcon>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn width="150px" align="center" prop="car_name" label="车">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.car_name === ''"><Minus/></ElIcon>
                            <span v-else>{{ scope.row.car_name }}</span>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="car_plate" label="车牌号">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.car_plate === ''"><Minus/></ElIcon>
                            <span v-else>{{ scope.row.car_plate }}</span>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="car_desc" min-width="150px" label="车说明">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.car_desc === ''"><Minus/></ElIcon>
                            <ElPopover v-else
                                        placement="left"
                                        title=""
                                        width="200"
                                        trigger="hover"
                                        :content="scope.row.car_desc">
                                <template #reference>
                                    <span class="table-description" slot="reference" >
                                        <ElIcon><Tickets/></ElIcon> {{scope.row.car_desc}}
                                    </span>
                                </template>
                            </ElPopover>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="message" min-width="100px" label="开篇">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.message === ''"><Minus/></ElIcon>
                            <ElPopover v-else
                                        placement="left"
                                        title=""
                                        width="200"
                                        trigger="hover"
                                        :content="scope.row.message">
                                <template #reference>
                                    <span class="table-description" slot="reference" >
                                    <ElIcon><Check/></ElIcon> {{scope.row.message}}</span>
                                </template>
                            </ElPopover>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="is_show_phone" label="手机" width="50px">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.is_show_phone"><Check/></ElIcon>
                            <ElIcon v-else><Minus/></ElIcon>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="is_show_wx" label="微信" width="50px">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.is_show_wx"><Check/></ElIcon>
                            <ElIcon v-else><Minus/></ElIcon>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="is_show_homepage" label="主页" width="50px">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.is_show_homepage"><Check/></ElIcon>
                            <ElIcon v-else><Minus/></ElIcon>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="visit_count" label="访问数"/>
                    <ElTableColumn align="center" prop="visit_count" label="图片数量">
                        <template #default="scope">
                            <span v-if="scope.row.imgs">{{ scope.row.imgs.split(',').length  }}</span>
                            <span v-else> - </span>
                        </template>
                    </ElTableColumn>
                    <ElTableColumn align="center" prop="description" min-width="100px" label="简介">
                        <template #default="scope">
                            <ElIcon v-if="scope.row.description === ''"><Minus/></ElIcon>
                            <ElPopover v-else
                                        placement="left"
                                        title=""
                                        width="200"
                                        trigger="hover"
                                        :content="scope.row.description">
                                <template #reference>
                                    <span class="table-description" slot="reference" >
                                        <ElIcon><Tickets/></ElIcon> {{scope.row.description}}</span>
                                </template>
                            </ElPopover>
                        </template>
                    </ElTableColumn>

                    <ElTableColumn align="center" fixed="right" label="操作" width="200px">
                        <template #default="scope">
                            <ElButton @click="goEdit(scope.row)" type="primary" plain icon="edit" size="small">编辑</ElButton>
                            <ElButton @click="goDelete(scope.row)" type="danger" plain icon="delete" size="small">删除</ElButton>
                        </template>
                    </ElTableColumn>
                </ElTable>

            </ElCol>
        </ElRow>


        <!-- MODAL -->
        <ElDialog
            :title="modalTitle"
            v-model="modalEdit"
            width="50%"
            :before-close="closeModal">
            <ElForm
                size="small"
                :model="formCode"
                :rules="codeRules"
                ref="codeModify" label-width="100px">
                <ElRow>

                    <ElCol :xs="24" :md="12">
                        <ElFormItem label="创建时间" prop="date_init">
                            <ElInput autocomplete="off" :disabled="true" v-model="formCode.date_init"></ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :xs="24" :md="12">
                        <ElFormItem label="所属" prop="uid">
                            <ElSelect v-model="formCode.uid" placeholder="请选择">
                                <ElOption
                                    v-for="user in userList"
                                    :key="user.uid"
                                    :label="user.nickname + (user.username? `(${user.username})`: '') "
                                    :value="user.uid">
                                </ElOption>
                            </ElSelect>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :xs="24" :md="12">
                        <ElFormItem label="hash" prop="hash">
                            <ElInput autocomplete="off" :disabled="!!editingHash" v-model="formCode.hash"></ElInput>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :xs="24" :md="8">
                        <ElFormItem class="mb-0" label="被访次数" prop="visit_count">
                            <ElInput autocomplete="off" v-model="formCode.visit_count">
                                <ElButton slot="append" icon="Close" @click="clearVisitCount">清空</ElButton>
                            </ElInput>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

                <ElDivider content-position="left"><b>开关</b></ElDivider>
                <ElRow>
                    <ElCol :xs="24" :md="6">
                        <ElFormItem label="公开" prop="is_public">
                            <ElSwitch :active-value="1" :inactive-value="0" v-model="formCode.is_public"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>
                <ElRow>
                    <ElCol :xs="24" :md="8">
                        <ElFormItem label="展示手机" prop="is_show_phone">
                            <ElSwitch :active-value="1" :inactive-value="0" v-model="formCode.is_show_phone"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :xs="24" :md="8">
                        <ElFormItem class="mb-0" label="展示高德组队" prop="is_show_gaode">
                            <ElSwitch :active-value="1" :inactive-value="0" v-model="formCode.is_show_gaode"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :xs="24" :md="8">
                        <ElFormItem class="mb-0" label="展示主页" prop="is_show_homepage">
                            <ElSwitch :active-value="1" :inactive-value="0" v-model="formCode.is_show_homepage"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>


                <ElDivider content-position="left"><b>文字描述</b></ElDivider>
                <ElRow>
                    <ElCol :xs="24" :md="24">
                        <ElFormItem label="挪车提示" prop="message">
                            <ElInput type="textarea" :rows="4" placeholder="请输入内容" v-model="formCode.message"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :xs="24" :md="24">
                        <ElFormItem label="">
                            <p>如果不填写，内容默认为</p>
                            <div class="text-success">
                                <p>您好，<br>
                                   如果车辆碍到您了，请联系我挪车</p>
                            </div>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

                <ElDivider content-position="left"><b>二维码信息</b></ElDivider>
                <ElRow>
                    <ElCol :xs="24" :md="24">
                        <ElFormItem label="展示微信" prop="is_show_wx">
                            <ElSwitch :active-value="1" :inactive-value="0" v-model="formCode.is_show_wx"/>
                        </ElFormItem>
                    </ElCol>

                    <ElCol :xs="24" :md="24">
                        <ElFormItem label="微信二维码（正方形）">
                            <ElUpload
                                action="files.kylebing.cn"
                                list-type="picture-card"
                                :before-upload="beforeUploadFile"
                                :http-request="uploadFileWx"
                                :file-list="wxCodeImgList">
                                <ElIcon><Upload/></ElIcon>
                            </ElUpload>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

                <ElDivider content-position="left"><b>车辆信息</b></ElDivider>
                <ElRow>
                    <ElCol :xs="24" :md="10">
                        <ElFormItem label="车辆名称" prop="car_name">
                            <ElInput autocomplete="off" v-model="formCode.car_name"></ElInput>
                        </ElFormItem>
                    </ElCol>

                    <ElCol :xs="24" :md="7">
                        <ElFormItem label="车辆类别" prop="car_type">
                            <ElSelect v-model="formCode.car_type">
                                <ElOption label="汽车" :value="0"></ElOption>
                                <ElOption label="摩托车" :value="1"></ElOption>
                            </ElSelect>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :xs="24" :md="7">
                        <ElFormItem label="是否显示" prop="is_show_car">
                            <ElSwitch :active-value="1" :inactive-value="0" v-model="formCode.is_show_car"/>
                        </ElFormItem>
                    </ElCol>
                    <ElCol :xs="24" :md="24">
                        <ElFormItem label="车牌号" prop="car_plate">
                            <ElInput autocomplete="off" v-model="formCode.car_plate"></ElInput>
                        </ElFormItem>
                    </ElCol>

                    <ElCol :xs="24" :md="24">
                        <ElFormItem class="mb-0" label="车辆描述" prop="car_desc">
                            <ElInput type="textarea" :rows="4" placeholder="请输入内容" v-model="formCode.car_desc"/>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

                <ElRow>
                    <ElCol :xs="24" :md="24">
                        <ElFormItem label="首页图片" prop="is_show_car">
                            <ElUpload
                                action="files.kylebing.cn"
                                list-type="picture-card"
                                :before-upload="beforeUploadFile"
                                :http-request="uploadFileImgs"
                                :file-list="currentImgList"
                                :on-remove="handleRemove">
                                <ElIcon><Upload/></ElIcon>
                            </ElUpload>
                        </ElFormItem>
                    </ElCol>
                </ElRow>

            </ElForm>
            <div slot="footer" class="dialog-footer">
                <ElButton size="small" @click="clearForm" type="warning">清空</ElButton>
                <ElButton size="small" @click="closeModal" icon="close">取 消</ElButton>
                <ElButton size="small" type="primary" @click="submit" icon="check">{{ editingHash ? '修改' : '添加' }}</ElButton>
            </div>
        </ElDialog>

    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import * as qiniu from 'qiniu-js'
import codeApi from "@/api/codeApi"
import userApi from "@/api/userApi"
import fileApi from "@/api/imageQiniuApi"
import Const from "@/const"
import Container from "@/layout/Container.vue"
import Toolbar from "@/layout/Toolbar.vue"

interface Code {
    date_modify: string
    date_init: string
    hash: string
    uid: string
    visit_count: string
    is_public: number
    is_show_phone: number
    is_show_homepage: number
    is_show_gaode: number
    message: string
    description: string
    is_show_wx: number
    wx_code_img: string
    car_name: string
    car_type: number
    car_plate: string
    car_desc: string
    is_show_car: number
    imgs: string
}

interface User {
    uid: string
    nickname: string
    username?: string
}

const isLoading = ref(false)
const codeList = ref<Code[]>([])
const userList = ref<User[]>([])
const currentImgList = ref<any[]>([])
const wxCodeImgList = ref<any[]>([])
const editingHash = ref<string | null>(null)
const modalEdit = ref(false)
const codeModify = ref<FormInstance>()

const formCode = ref<Code>({
    date_modify: '',
    date_init: '',
    hash: '',
    uid: '',
    visit_count: '',
    is_public: 0,
    is_show_phone: 0,
    is_show_homepage: 0,
    is_show_gaode: 0,
    message: '',
    description: '',
    is_show_wx: 0,
    wx_code_img: '',
    car_name: '',
    car_type: 0,
    car_plate: '',
    car_desc: '',
    is_show_car: 0,
    imgs: ''
})

const codeRules = {
    hash: { required: true, message: 'hash 必填', trigger: 'blur' },
    uid: { required: true, message: '用户必选', trigger: 'blur' },
    car_name: { required: true, message: '请填写车辆名称', trigger: 'change' },
    car_type: { required: true, message: '车辆类别必选', trigger: 'change' },
    car_plate: { required: true, message: '请填写车牌号', trigger: 'blur' }
}

const modalTitle = computed(() => editingHash.value ? '编辑二维码' : '新增二维码')

function beforeUploadFile(file: File) {
    if (file.size > 1024 * 1024 * 10) {
        ElMessage.warning('图片大小请控制在 10Mb 以内')
        return false
    }
    return true
}

function uploadFileImgs(uploadInfo: any) {
    fileApi
        .getUploadToken({
            bucket: 'qrmanager'
        })
        .then(res => {
            console.log('get token success')
            const observer = {
                next: (res: any) => {
                    console.log('next: ', res)
                },
                error: (err: any) => {
                    console.log(err)
                },
                complete: (res: any) => {
                    console.log('complete: ', res)
                    currentImgList.value.push({
                        name: res.key,
                        url: Const.imgBaseURL + res.key
                    })
                }
            }
            const observable = qiniu.upload(uploadInfo.file, null, res.data, {}, {})
            const subscription = observable.subscribe(observer)
        })
        .catch(err => {
            ElMessage.error('获取上传 token 失败')
        })
}

function uploadFileWx(uploadInfo: any) {
    fileApi
        .getUploadToken({
            bucket: 'qrmanager'
        })
        .then(res => {
            console.log('get token success')
            const observer = {
                next: (res: any) => {
                    console.log('next: ', res)
                },
                error: (err: any) => {
                    console.log(err)
                },
                complete: (res: any) => {
                    formCode.value.wx_code_img = Const.imgBaseURL + res.key
                    wxCodeImgList.value = [{ name: 'wx', url: formCode.value.wx_code_img }]
                }
            }
            const observable = qiniu.upload(uploadInfo.file, null, res.data, {}, {})
            const subscription = observable.subscribe(observer)
        })
        .catch(err => {
            ElMessage.error('获取上传 token 失败')
        })
}

function handleRemove(file: any, fileList: any[]) {
    currentImgList.value = fileList
}

function addNewCode() {
    modalEdit.value = true
    editingHash.value = null
    clearForm()
}

function clearForm() {
    formCode.value = {
        date_modify: '',
        date_init: '',
        hash: '',
        uid: '',
        visit_count: '',
        is_public: 0,
        is_show_phone: 0,
        is_show_homepage: 0,
        is_show_gaode: 0,
        message: '',
        description: '',
        is_show_wx: 0,
        wx_code_img: '',
        car_name: '',
        car_type: 0,
        car_plate: '',
        car_desc: '',
        is_show_car: 0,
        imgs: ''
    }
    currentImgList.value = []
    wxCodeImgList.value = []
}

function closeModal(done?: () => void) {
    ElMessageBox.confirm('确认关闭？')
        .then(() => {
            editingHash.value = null
            modalEdit.value = false
            if (done) done()
        })
        .catch(() => {})
}

function clearVisitCount() {
    const requestData = {
        hash: editingHash.value
    }
    codeApi.clearCount(requestData)
        .then(res => {
            formCode.value.visit_count = '0'
            ElMessage.success(res.message)
            getCodeList()
        })
}

function getUserList() {
    userApi
        .list({
            pageNo: 1,
            pageSize: 500
        })
        .then(res => {
            userList.value = res.data.list
        })
}

function getCodeList() {
    isLoading.value = true
    const requestData = {
        pageNo: 1,
        pageSize: 50
    }
    codeApi.list(requestData)
        .then(res => {
            isLoading.value = false
            codeList.value = res.data
        })
        .catch(err => {
            isLoading.value = false
        })
}

function goEdit(code: Code) {
    editingHash.value = code.hash
    formCode.value = { ...code }
    if (code.imgs) {
        currentImgList.value = code.imgs.split(',').map(item => ({
            url: item,
            name: ''
        }))
    } else {
        currentImgList.value = []
    }
    if (code.wx_code_img) {
        wxCodeImgList.value = [{ name: 'wx', url: formCode.value.wx_code_img }]
    } else {
        wxCodeImgList.value = []
    }
    modalEdit.value = true
}

function goDelete(code: Code) {
    ElMessageBox.confirm(`删除二维码 ${code.hash}`, '删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        const requestData = {
            hash: code.hash
        }
        codeApi.delete(requestData)
            .then(res => {
                getCodeList()
                ElMessage({
                    message: res.message,
                    type: 'success'
                })
            })
    })
}

function submit() {
    if (!codeModify.value) return
    codeModify.value.validate((valid) => {
        if (valid) {
            if (editingHash.value) {
                codeModifySubmit()
            } else {
                codeNewSubmit()
            }
        } else {
            console.log('error submit!!')
            return false
        }
    })
}

function codeNewSubmit() {
    formCode.value.imgs = currentImgList.value.map(item => item.url).join(',')
    codeApi
        .add(formCode.value)
        .then(res => {
            ElMessage({
                message: res.message,
                type: 'success'
            })
            getCodeList()
            editingHash.value = null
            modalEdit.value = false
        })
}

function codeModifySubmit() {
    formCode.value.imgs = currentImgList.value.map(item => item.url).join(',')
    codeApi
        .modify(formCode.value)
        .then(res => {
            ElMessage({
                message: res.message,
                type: 'success'
            })
            getCodeList()
            editingHash.value = null
            modalEdit.value = false
        })
}

onMounted(() => {
    getCodeList()
    getUserList()
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
