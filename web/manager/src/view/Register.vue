<template>
    <div class="register-bg" :style="`height:${viewportHeight}px`">
        <div class="register-panel">
            <div class="register-title">
                <h2>注册账号</h2>
                <p>填写以下信息。注册后请先完成邮箱验证，再登录。</p>
            </div>
            <ElForm size="large" :model="formRegister" :rules="userRules" ref="registerForm" label-width="0" @submit.prevent="submit">
                <ElFormItem prop="invitationCode">
                    <ElInput clearable autocomplete="off" placeholder="邀请码" v-model="formRegister.invitationCode"/>
                    <p v-if="!invitationRequired" class="field-hint">还没有用户时可以不填，首位注册者会成为管理员。</p>
                </ElFormItem>
                <ElFormItem prop="nickname">
                    <ElInput clearable autocomplete="off" placeholder="昵称" v-model="formRegister.nickname"/>
                </ElFormItem>
                <ElFormItem prop="email">
                    <ElInput clearable autocomplete="off" placeholder="邮箱" v-model="formRegister.email"/>
                </ElFormItem>
                <ElFormItem prop="password">
                    <ElInput clearable show-password type="password" autocomplete="new-password" placeholder="密码" v-model="formRegister.password"/>
                </ElFormItem>
                <ElFormItem prop="passwordrepeat">
                    <ElInput clearable show-password type="password" autocomplete="new-password" placeholder="再次输入密码"
                             @keydown.enter="submit" v-model="formRegister.passwordrepeat"/>
                </ElFormItem>
                <ElButton class="register-btn" type="primary" native-type="submit" :loading="submitting" @click="submit">注册</ElButton>
            </ElForm>
            <div class="register-links">
                <RouterLink to="/login">已有账号，去登录</RouterLink>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useProjectStore } from "@/pinia"
import { useRouter } from 'vue-router'
import { ElNotification, ElButton, ElForm, ElFormItem, ElInput } from 'element-plus'
import type { FormInstance } from 'element-plus'
import userApi from "@/api/userApi"
import { request } from "@/api/request"

const projectStore = useProjectStore()
const router = useRouter()

interface RegisterForm {
    email: string
    nickname: string
    password: string
    passwordrepeat: string
    invitationCode: string
}

const formRegister = ref<RegisterForm>({
    email: '',
    nickname: '',
    password: '',
    passwordrepeat: '',
    invitationCode: ''
})

const registerForm = ref<FormInstance>()
const submitting = ref(false)
// 已有用户时邀请码必填，与日记注册一致
const invitationRequired = ref(true)

onMounted(() => {
    request('get', null, null, false, '/setup/status')
        .then((res: any) => {
            invitationRequired.value = !!res.data?.hasRegisteredUsers
        })
        .catch(() => {
            invitationRequired.value = true
        })
})

function validatePasswordRepeat(_rule: unknown, value: string, callback: (err?: Error) => void) {
    if (value !== formRegister.value.password) {
        callback(new Error("两次密码输入不一致"))
    } else {
        callback()
    }
}

const userRules = computed(() => ({
    email: [
        { required: true, message: '请填写邮箱', trigger: 'blur' },
        {
            validator: function(_rule: unknown, value: string, callback: (err?: Error) => void) {
                if (!/(\w|\d)+@(\w|\d)+\.\w+/i.test(value)) {
                    callback(new Error("邮箱格式不正确"))
                } else {
                    callback()
                }
            }
        }
    ],
    nickname: { required: true, message: '请填写昵称', trigger: 'blur' },
    password: { required: true, message: '请填写密码', trigger: 'blur' },
    passwordrepeat: [
        { required: true, message: '请再填写一次密码', trigger: 'blur' },
        { validator: validatePasswordRepeat, trigger: 'blur' }
    ],
    invitationCode: invitationRequired.value
        ? { required: true, message: '请输入邀请码', trigger: 'blur' }
        : { required: false },
}))

// 页面整体按 1920 缩放且 body 不能滚动，注册表单在可视高度内自行滚动
const viewportHeight = computed(() => {
    const scale = projectStore.pageScale || 1
    return projectStore.windowInsets.height / scale
})

function submit() {
    if (!registerForm.value || submitting.value) return
    registerForm.value.validate((valid) => {
        if (!valid) return
        submitting.value = true
        userApi.register({
            nickname: formRegister.value.nickname.trim(),
            email: formRegister.value.email.trim(),
            password: formRegister.value.password,
            invitationCode: formRegister.value.invitationCode.trim(),
        })
            .then((res: any) => {
                const needVerify = res.data && res.data.email_verified === false
                ElNotification({
                    title: res.message || '注册成功',
                    message: needVerify ? '请查收邮箱完成验证后再登录' : '请登录',
                    position: 'top-right',
                    type: 'success',
                })
                router.push('/login')
            })
            .catch(() => {})
            .finally(() => {
                submitting.value = false
            })
    })
}
</script>

<style lang="scss" scoped>

.register-bg {
    display: flex;
    justify-content: center;
    align-items: center;
    overflow-y: auto;
    background: #eef1f4;
}

.register-panel {
    z-index: 10;
    width: 420px;
    margin: 40px 0;
    padding: 40px 40px 28px;
    border: 1px solid #e6e8eb;
    border-radius: 16px;
    background: #fff;
    box-shadow: 0 12px 40px rgba(20, 24, 28, 0.08);
}

.register-title {
    text-align: left;
    margin-bottom: 28px;

    h2 {
        margin: 0;
        font-size: 24px;
        font-weight: 600;
        letter-spacing: 0.5px;
        color: #1a1d21;
    }

    p {
        margin: 8px 0 0;
        font-size: 13px;
        line-height: 1.6;
        color: #6b7280;
    }
}

.field-hint {
    margin: 6px 2px 0;
    font-size: 12px;
    line-height: 1.5;
    color: #6b7280;
}

.register-btn {
    display: block;
    width: 100%;
    margin-top: 6px;
}

.register-links {
    margin-top: 18px;
    text-align: center;

    a {
        color: #4b5563;
        font-size: 13px;
        text-decoration: none;
    }

    a:hover {
        color: #1a1d21;
    }
}
</style>
