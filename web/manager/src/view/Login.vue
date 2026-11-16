<template>
    <div class="login-bg">
        <div class="login-stack">
        <div class="login-panel">
            <div class="login-title">
                <h2 class="login-title__name">{{ packageInfo.projectName }}</h2>
            </div>
            <ElForm size="large" :model="formLogin" :rules="formLoginRules" ref="formEl" label-width="0">
                <ElFormItem label="" class="mb-5" prop="account">
                    <ElInput clearable autocomplete="off" placeholder="邮箱" v-model="formLogin.email" />
                </ElFormItem>
                <ElFormItem label="" class="mb-8" prop="password">
                    <ElInput
                        clearable
                        :show-password="true"
                        type="password"
                        autocomplete="off"
                        placeholder="密码"
                        v-model="formLogin.password"
                        @keydown.enter="submit(formEl)"
                    />
                </ElFormItem>
                <ElFormItem align="center">
                    <ElButton :loading="isInLoginProcess" class="login-btn" type="primary" @click="submit(formEl)"> 登录
                    </ElButton>
                </ElFormItem>
                <ElFormItem v-if="needResendVerify" align="center">
                    <ElButton class="login-btn" @click="resendVerify">重发验证邮件</ElButton>
                </ElFormItem>
                <div class="login-links">
                    <RouterLink to="/register">注册</RouterLink>
                    <RouterLink to="/register-apply">申请注册</RouterLink>
                    <a href="javascript:;" @click="showForgot = true">找回密码</a>
                </div>
            </ElForm>
        </div>
        </div>
    </div>

    <ElDialog
        v-model="showForgot"
        title="找回密码"
        :width="forgotDialogWidth"
        :append-to-body="true"
        align-center
        destroy-on-close
        @closed="forgotEmail = ''"
    >
        <p class="forgot-tip">输入注册邮箱，将发送重置密码链接（1 小时内有效）。</p>
        <ElInput v-model="forgotEmail" placeholder="邮箱" clearable @keydown.enter="sendForgot" />
        <template #footer>
            <ElButton @click="showForgot = false">取消</ElButton>
            <ElButton type="primary" :loading="forgotSending" @click="sendForgot">发送重置邮件</ElButton>
        </template>
    </ElDialog>
</template>

<script setup lang="ts">
import {ref, computed, onMounted, onUnmounted, nextTick} from "vue";
import { FormInstance, ElNotification, ElMessage } from "element-plus";
import userApi from "../api/userApi.ts";
import { useRouter } from "vue-router";
import {  setAuthorization } from "../utility.ts";
import packageInfo from "../../package.json"
import {AnimateHeartCanvas} from "@/utils/animateHeartCanvas"
import {useMenuStore} from "@/pinia/menuStore.ts";

const router = useRouter()

const formLogin = ref<{ email: string, password: string }>({
    email: '',
    password: ''
});

const needResendVerify = ref(false)
const showForgot = ref(false)
const forgotEmail = ref('')
const forgotSending = ref(false)

let animatedBg: AnimateHeartCanvas | null = null

const forgotDialogWidth = computed(() =>
    typeof window !== 'undefined' && window.innerWidth < 480 ? 'min(92vw, 360px)' : '400px'
)

onMounted(function() {
    const narrow = typeof window !== 'undefined' && window.matchMedia('(max-width: 480px)').matches
    // 心画在 canvas 上（z-index 低于面板）；移动端减少粒子数以减轻卡顿
    animatedBg = new AnimateHeartCanvas(
        narrow ? 330 : 0,    // hmin（窄屏偏粉红，宽屏五彩）
        narrow ? 350 : 200,  // hmax
        narrow ? 90 : 140,   // count
        narrow ? 12 : 18,    // size min
        narrow ? 36 : 56,    // size max
        narrow ? 28 : 30,    // 透明度下限（%）
        narrow ? 70 : 75,    // 透明度上限（%）
        '#141416',
    )
})

onUnmounted(function() {
    animatedBg?.destroy()
    animatedBg = null
})

const formLoginRules = ref({
    email: {required: true, message: '请填写用户名', trigger: 'blur'},
    password: {required: true, message: '请填写密码', trigger: 'blur'},
})

const isInLoginProcess = ref(false)  // 登录中状态展示


const formEl = ref<FormInstance>()

function submit(formEl: FormInstance | undefined) {
    if (!formEl) return
    formEl.validate(function(valid) {
        if (valid) {
            login()
        } else {
            console.log('error submit(formEl)!!')
            return false
        }
    })
}

function login() {
    isInLoginProcess.value = true
    needResendVerify.value = false
    userApi
        .login({
            email: formLogin.value.email,
            password: formLogin.value.password
        })
        .then(function(res) {
            isInLoginProcess.value = false
            setAuthorization(
                res.data.nickname,
                res.data.uid,
                res.data.email,
                res.data.phone,
                res.data.avatar,
                res.data.token,
                res.data.group_id,
                res.data.city,
                res.data.geolocation,
            )
            ElNotification({
                title: '登录成功',
                message: '欢迎用户' + res.data.username,
                position: 'top-right',
                type: 'success',
                duration: 2000
            })

            useMenuStore().refreshRoute(router)
            useMenuStore().generateMenuArrayAndMap()

            nextTick(() => {
                router.replace('/diary/statistic')
            })
        })
        .catch(function(err) {
            console.log(err)
            isInLoginProcess.value = false
            if (err?.data?.code === 'email_not_verified') {
                needResendVerify.value = true
            }
        })
}

function resendVerify() {
    const email = formLogin.value.email.trim()
    if (!email) return
    userApi.resendVerify({ email })
        .then((res: any) => {
            ElMessage.success(res.message || '若该邮箱已注册且未验证，验证邮件已发送')
        })
}

function sendForgot() {
    const email = (forgotEmail.value || formLogin.value.email).trim()
    if (!email) {
        ElMessage.warning('请填写邮箱')
        return
    }
    forgotSending.value = true
    userApi.forgot({ email })
        .then((res: any) => {
            ElMessage.success(res.message || '若该邮箱已注册，重置邮件已发送')
            showForgot.value = false
        })
        .finally(() => {
            forgotSending.value = false
        })
}
</script>

<style scoped lang="scss">
@use "sass:color";
@use "../assets/scss/variables" as *;
@use "../assets/scss/utility" as *;
@use "../assets/scss/font" as *;

.login-stack {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 100%;
    max-width: 450px;
    margin: auto 0;
}

.login-bg {
    display: flex;
    justify-content: center;
    align-items: center;
    box-sizing: border-box;
    min-height: 100dvh;
    min-height: 100svh;
    padding:
        max(16px, env(safe-area-inset-top, 0px))
        max(16px, env(safe-area-inset-right, 0px))
        max(20px, env(safe-area-inset-bottom, 0px))
        max(16px, env(safe-area-inset-left, 0px));
    overflow-x: hidden;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    // 背景由心形 canvas 绘制（质感雅黑），此处保持透明以免挡住心
    background: transparent;

    @media (max-width: 480px) {
        align-items: flex-start;
        padding-top: max(28px, env(safe-area-inset-top, 0px));
    }
}

.login-title {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-bottom: 28px;
    text-align: center;
    color: white;
    text-shadow: 2px 2px 1px color.adjust(black, $alpha: -0.8);

    &__name {
        margin: 0;
        font-size: 1.25rem;
        font-weight: 600;
        letter-spacing: 2px;
        line-height: 1.3;
    }

    @media (min-width: 481px) {
        margin-bottom: 40px;

        &__name {
            font-size: 1.5em;
            font-weight: bold;
            letter-spacing: 3px;
        }
    }
}

.login-panel {
    /* 让 z-index 生效，保证面板在心形 canvas 之上 */
    position: relative;
    z-index: 10;
    box-sizing: border-box;
    width: 100%;
    margin: 0;
    padding: 24px 20px 22px;
    @include border-radius(16px);
    background: linear-gradient(to bottom right, color.adjust(white, $alpha: -0.5), color.adjust(white, $alpha: -0.8));
    @include backdrop-filter(saturate(150%) blur(15px));
    @include box-shadow(10px 10px 35px color.adjust(black, $alpha: -0.8));

    @media (min-width: 480px) {
        padding: 30px 50px;
        @include border-radius(20px);
    }

    // Chrome 自动填充会给 input 加黄/蓝底，用 inset 阴影盖掉
    :deep(.el-input__wrapper) {
        background-color: #fff !important;
    }

    :deep(input.el-input__inner) {
        &:-webkit-autofill,
        &:-webkit-autofill:hover,
        &:-webkit-autofill:focus,
        &:-webkit-autofill:active {
            -webkit-text-fill-color: #303133 !important;
            caret-color: #303133;
            transition: background-color 99999s ease-out 0s;
            box-shadow: 0 0 0 1000px #fff inset !important;
            -webkit-box-shadow: 0 0 0 1000px #fff inset !important;
        }
    }
}

.login-btn {
    display: block;
    width: 100%;
}

.login-links {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 8px 16px;
    margin-top: 12px;
    line-height: 1.5;

    a {
        color: rgba(255, 255, 255, 0.92);
        font-size: 13px;
        text-decoration: underline;
        padding: 4px 2px;
        -webkit-tap-highlight-color: transparent;
    }
}

.login-panel :deep(.el-form-item) {
    margin-bottom: 18px;
}

.login-panel :deep(.el-form-item.mb-8) {
    margin-bottom: 22px;
}

.forgot-tip {
    margin: 0 0 12px;
    color: #606266;
    font-size: 13px;
    line-height: 1.5;
}
</style>
