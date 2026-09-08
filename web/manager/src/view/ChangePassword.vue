<template>
    <Container>
        <Toolbar>
            <template #left>
            </template>
            <template #center>
            </template>
            <template #right>
            </template>
        </Toolbar>
        <ElRow>
            <ElCol :lg="10" :span="24">
                <ElForm
                    :model="formPassword"
                    :rules="passwordRules"
                    ref="formChangePassword" label-width="100px">
                    <ElFormItem label="新密码" prop="password">
                        <ElInput autocomplete="off" v-model="formPassword.password"/>
                    </ElFormItem>
                    <ElFormItem label="确认密码" prop="passwordRepeat">
                        <ElInput autocomplete="off" v-model="formPassword.passwordRepeat"/>
                    </ElFormItem>
                    <ElFormItem>
                        <ElButton type="primary" @click="submit" icon="check"> 确认修改</ElButton>
                    </ElFormItem>
                </ElForm>
            </ElCol>
        </ElRow>
    </Container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElNotification } from 'element-plus'
import type { FormInstance } from 'element-plus'
import userApi from "@/api/userApi"
import { deleteAuthorization } from "@/utility"
import Container from "@/layout/Container.vue"
import Toolbar from "@/layout/Toolbar.vue";

const router = useRouter()
const formChangePassword = ref<FormInstance>()

interface PasswordForm {
    password: string
    passwordRepeat: string
}

const formPassword = ref<PasswordForm>({
    password: '',
    passwordRepeat: '',
})

function validatePasswordRepeat(rule: any, value: string, callback: Function) {
    if (value !== formPassword.value.password) {
        callback(new Error("两次密码输入不一致"))
    } else {
        callback()
    }
}

const passwordRules = {
    password: [{required: true, message: '请填写新密码', trigger: 'blur'}],
    passwordRepeat: [
        {required: true, message: '请再填写一次新密码', trigger: 'blur'},
        {validator: validatePasswordRepeat, trigger: 'blur'},
    ],
}

function submit() {
    if (!formChangePassword.value) return
    formChangePassword.value.validate(function(valid) {
        if (valid) {
            changePassword()
        } else {
            console.log('error submit!!')
            return false
        }
    })
}

function changePassword(){
    const requestData = {
        password: formPassword.value.password
    }
    userApi
        .changePassword(requestData)
        .then(function(res) {
            ElNotification({
                title: res.message,
                message: '稍候会自动跳转至登录页面',
                position: 'top-right',
                type: 'success',
                onClose: function() {
                    deleteAuthorization()
                    router.push('/login')
                }
            })
        })
}
</script>

<style scoped>
</style>
