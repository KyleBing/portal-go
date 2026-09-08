<template>
    <div class="container">
        <ElRow>
            <ElCol :span="10" :offset="(24 - 10)/2">
                <div class="content"
                     :style="`min-height: ${windowInsets.height}px`"
                    >
                    <div class="register-header">
                        <img src="../assets/logo.png" alt="LOGO">
                        <h2>注册</h2>
                    </div>
                    <ElForm
                        label-position="top"
                        :model="formRegister"
                        :rules="userRules"
                        ref="registerForm"
                        label-width="100px">
                        <ElFormItem label="邮箱" prop="email">
                            <ElInput autocomplete="off" v-model="formRegister.email"/>
                        </ElFormItem>
                        <ElFormItem label="用户名" prop="username">
                            <ElInput autocomplete="off" v-model="formRegister.username"/>
                        </ElFormItem>
                        <ElFormItem label="昵称" prop="nickname">
                            <ElInput autocomplete="off" v-model="formRegister.nickname"/>
                        </ElFormItem>
                        <ElFormItem label="微信" prop="wx">
                            <ElInput autocomplete="off" v-model="formRegister.wx"/>
                        </ElFormItem>
                        <ElFormItem label="手机" prop="phone">
                            <ElInput autocomplete="off" v-model="formRegister.phone"/>
                        </ElFormItem>
                        <ElFormItem label="高德组队码" prop="gaode">
                            <ElInput autocomplete="off" v-model="formRegister.gaode"/>
                        </ElFormItem>
                        <ElFormItem label="主页" prop="homepage">
                            <ElInput autocomplete="off" v-model="formRegister.homepage"/>
                        </ElFormItem>
                        <ElFormItem label="密码" prop="password">
                            <ElInput autocomplete="off" v-model="formRegister.password"/>
                        </ElFormItem>
                        <ElFormItem label="确认密码" prop="passwordrepeat">
                            <ElInput autocomplete="off" v-model="formRegister.passwordrepeat"/>
                        </ElFormItem>
                        <ElFormItem label="邀请码" prop="invitationCode">
                            <ElInput autocomplete="off" v-model="formRegister.invitationCode"/>
                        </ElFormItem>
                    </ElForm>
                    <div class="submit">
                        <ElButton style="width: 200px" type="primary" @click="submit">注册</ElButton>
                    </div>
                </div>
            </ElCol>
        </ElRow>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useProjectStore } from "@/pinia"
import { useRouter } from 'vue-router'
import { ElNotification, ElButton, ElForm, ElFormItem, ElInput, ElRow, ElCol } from 'element-plus'
import type { FormInstance } from 'element-plus'
import userApi from "@/api/userApi"

const projectStore = useProjectStore()
const router = useRouter()

interface RegisterForm {
    email: string
    username: string
    nickname: string
    wx: string
    phone: string
    homepage: string
    gaode: string
    password: string
    passwordrepeat: string
    invitationCode: string
}

const formRegister = ref<RegisterForm>({
    email: '',
    username: '',
    nickname: '',
    wx: '',
    phone: '',
    homepage: '',
    gaode: '',
    password: '',
    passwordrepeat: '',
    invitationCode: ''
})

const registerForm = ref<FormInstance>()

function validatePassworDrepeat(rule: any, value: string, callback: Function) {
    if (value !== formRegister.value.password) {
        callback(new Error("两次密码输入不一致"))
    } else {
        callback()
    }
}

const userRules = {
    email: [
        { required: true, message: '请填写邮箱', trigger: 'blur' },
        {
            validator: function(rule: any, value: string, callback: Function) {
                if (!/(\w|\d)+@(\w|\d)+\.\w+/i.test(value)) {
                    callback(new Error("用户名只能是小写字母"))
                } else {
                    callback()
                }
            }
        }
    ],
    username: [
        { required: true, message: '请填写用户名', trigger: 'blur' },
        {
            validator: function(rule: any, value: string, callback: Function) {
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
    password: { required: true, message: '请填写原密码', trigger: 'blur' },
    passwordrepeat: [
        { required: true, message: '请再填写一次新密码', trigger: 'blur' },
        { validator: validatePassworDrepeat, trigger: 'blur' }
    ],
    invitationCode: { required: true, message: '请输入邀请码', trigger: 'blur' },
}

const windowInsets = computed(() => projectStore.windowInsets)

function submit() {
    if (!registerForm.value) return
    registerForm.value.validate((valid) => {
        if (valid) {
            register()
        } else {
            console.log('error submit!!')
            return false
        }
    })
}

function register() {
    userApi.register(formRegister.value)
        .then(res => {
            ElNotification({
                title: res.message,
                message: '请登录',
                position: 'top-right',
                type: 'success',
                onClose() {
                }
            })
            router.push('/login')
        })
}
</script>

<style lang="scss" scoped>
@use "../assets/scss/variables" as *;
@use "../assets/scss/utility" as *;
@use "sass:color";

$height: 60px;

.container {
    background-color: $color-border;
}

.content {
    padding: 60px;
    background-color: white;
    @include box-shadow(1px 2px 5px color.adjust(black, $alpha: -0.9))
}

.register-header {
    margin-bottom: 30px;
    display: flex;
    justify-content: center;

    img {
        padding: 5px;
        margin-right: 20px;
        display: block;
        height: $height;
    }

    h2 {
        color: $text-main;
        line-height: $height;
    }
}

.submit {
    display: flex;
    justify-content: center;
}
</style>
