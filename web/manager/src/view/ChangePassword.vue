<template>
    <Container class="bg">
        <div class="content">
            <div class="card">
                <h2 class="page-title">修改密码</h2>
                <p class="page-hint">修改成功后将退出登录，请使用新密码重新登录。</p>
                <ElForm
                    class="password-form"
                    :model="formPassword"
                    :rules="passwordRules"
                    ref="formChangePassword"
                    label-width="100px"
                    label-position="top"
                >
                    <ElFormItem label="新密码" prop="password">
                        <ElInput
                            v-model="formPassword.password"
                            type="password"
                            show-password
                            autocomplete="new-password"
                            placeholder="请输入新密码"
                        />
                    </ElFormItem>
                    <ElFormItem label="确认密码" prop="passwordRepeat">
                        <ElInput
                            v-model="formPassword.passwordRepeat"
                            type="password"
                            show-password
                            autocomplete="new-password"
                            placeholder="请再次输入新密码"
                        />
                    </ElFormItem>
                    <ElFormItem>
                        <ElButton type="primary" icon="Check" @click="submit">确认修改</ElButton>
                    </ElFormItem>
                </ElForm>
            </div>

            <div class="card account-data-card">
                <h2 class="page-title">账号数据</h2>
                <p class="page-hint">
                    <strong>注销账号</strong>：永久删除该账号在数据库中的全部数据，包括账号资料、日记、二维码、路书、邀请码和邮箱验证记录。注销后无法登录、无法恢复。
                </p>
                <div class="danger-actions">
                    <ElButton type="danger" plain icon="CircleClose" @click="destroyAccount">注销账号</ElButton>
                </div>
                <p class="page-hint">操作需要在确认框里完整输入当前账号邮箱后才能执行。</p>
            </div>
        </div>
    </Container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox, ElNotification } from 'element-plus'
import type { FormInstance } from 'element-plus'
import userApi from '@/api/userApi'
import { deleteAuthorization, getAuthorization } from '@/utility'
import Container from '@/layout/Container.vue'

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

function validatePasswordRepeat(_rule: unknown, value: string, callback: (err?: Error) => void) {
    if (value !== formPassword.value.password) {
        callback(new Error('两次密码输入不一致'))
    } else {
        callback()
    }
}

const passwordRules = {
    password: [{ required: true, message: '请填写新密码', trigger: 'blur' }],
    passwordRepeat: [
        { required: true, message: '请再填写一次新密码', trigger: 'blur' },
        { validator: validatePasswordRepeat, trigger: 'blur' },
    ],
}

function submit() {
    if (!formChangePassword.value) return
    formChangePassword.value.validate((valid) => {
        if (valid) {
            changePassword()
        }
    })
}

function changePassword() {
    userApi
        .changePassword({ password: formPassword.value.password })
        .then((res) => {
            ElNotification({
                title: res.message,
                message: '稍候会自动跳转至登录页面',
                position: 'top-right',
                type: 'success',
                onClose() {
                    deleteAuthorization()
                    router.push('/login')
                },
            })
        })
}

// 注销账号：输入完整邮箱确认后清除本地登录态并回到登录页
function destroyAccount() {
    confirmByEmail(
        '注销账号',
        '注销会永久删除账号资料、日记、二维码、路书、邀请码和邮箱验证记录。注销后无法登录，也无法恢复。',
        '永久注销',
    ).then(() => userApi.destroyAccount())
        .then((res) => {
            ElNotification({
                title: res.message,
                message: '账号已注销，稍后跳转至登录页',
                position: 'top-right',
                type: 'success',
                onClose() {
                    deleteAuthorization()
                    router.push('/login')
                },
            })
        })
        .catch(() => {})
}

// 破坏性操作确认：必须完整输入当前账号邮箱，一致才继续
function confirmByEmail(title: string, message: string, confirmText: string) {
    const expected = String(getAuthorization()?.email ?? '').trim()
    if (!expected) {
        return Promise.reject(new Error('没有读到账号邮箱，请重新登录后再试'))
    }
    return ElMessageBox.prompt(
        `${message}\n\n请输入账号邮箱 ${expected} 以确认。`,
        title,
        {
            type: 'warning',
            confirmButtonText: confirmText,
            cancelButtonText: '取消',
            inputPlaceholder: '输入账号邮箱',
            // 必须完全一致（忽略首尾空格与大小写）才允许提交
            inputValidator: (value: string) =>
                String(value ?? '').trim().toLowerCase() === expected.toLowerCase()
                    ? true
                    : '邮箱不一致，请完整输入账号邮箱',
        },
    )
}
</script>

<style lang="scss" scoped>
@use '../assets/scss/variables' as *;
@use 'sass:color';

$text-muted: #666;

.bg {
    background-color: $color-border;
}

.content {
    width: 500px;
    max-width: 100%;
    padding: 20px;
}

.card {
    color: $text-muted;
    background-color: white;
    box-shadow: 1px 1px 3px color.adjust(black, $alpha: -0.82);
    padding: 24px 20px 20px;
    border-radius: 6px;
}

.page-title {
    margin: 0 0 8px;
    padding-left: 12px;
    font-size: 1rem;
    font-weight: bold;
    line-height: 2;
    color: $text-muted;
    border-left: 3px solid $color-main;
}

.page-hint {
    margin: 0 0 20px;
    padding: 0 4px;
    font-size: 0.8rem;
    line-height: 1.5;
    color: color.adjust($text-muted, $lightness: 10%);
}

.password-form {
    padding: 0 4px;

    :deep(.el-input__wrapper) {
        background-color: #fff !important;
    }
}

.account-data-card {
    margin-top: 16px;
}

.danger-actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    padding: 4px;
}
</style>
