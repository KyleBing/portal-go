<template>
    <div class="login-bg" :style="`height:${projectStore.windowInsets.height / projectStore.pageScale}px`">
        <div class="login-panel">
            <div class="login-title">
                <!--                <div class="logo">-->
                <!--                    <img src="../assets/logo/logo.png" alt="LOGO">-->
                <!--                </div>-->
                <h2>{{ packageInfo.projectName }}</h2>
            </div>
            <ElForm size="large" :model="formLogin" :rules="formLoginRules" ref="formEl" label-width="0">
                <ElFormItem label="" class="mb-5" prop="account">
                    <ElInput clearable autocomplete="off" placeholder="邮箱" v-model="formLogin.email" />
                </ElFormItem>
                <ElFormItem label="" class="mb-8" prop="password">
                    <ElInput clearable :show-password="true" type="password" autocomplete="off"
                        @keydown.enter.native="submit(formEl)" placeholder="密码" v-model="formLogin.password" />
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
                    <a href="javascript:;" @click="showForgot = true">找回密码</a>
                </div>
            </ElForm>

            <ElDialog v-model="showForgot" title="找回密码" width="400px" append-to-body>
                <p class="forgot-tip">输入注册邮箱，将发送重置密码链接（1 小时内有效）。</p>
                <ElInput v-model="forgotEmail" placeholder="邮箱" clearable />
                <template #footer>
                    <ElButton @click="showForgot = false">取消</ElButton>
                    <ElButton type="primary" :loading="forgotSending" @click="sendForgot">发送重置邮件</ElButton>
                </template>
            </ElDialog>

        </div>
    </div>
</template>

<script setup lang="ts">
import {ref, onMounted, onUnmounted, nextTick} from "vue";
import {useProjectStore} from "../pinia";
import { FormInstance, ElNotification, ElDialog, ElMessage } from "element-plus";
import userApi from "../api/userApi.ts";
const projectStore = useProjectStore()
import { useRouter } from "vue-router";
import {  setAuthorization } from "../utility.ts";
import packageInfo from "../../package.json"
import {AnimateHeartCanvas} from "animate-heart-canvas"
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

let animatedBg = null

onMounted(function() {
    animatedBg = new AnimateHeartCanvas(0,200,200,50,10, '#3d3d3d')
    // constructor(hMin, hMax, countHeart = 150, sizeMin = 50, sizeMax = 350, bgColor) {

})

onUnmounted(function() {
    animatedBg.destroy()
})

const formLoginRules = ref({
    email: {required: true, message: '请填写用户名', trigger: 'blur'},
    password: {required: true, message: '请填写密码', trigger: 'blur'},
})

const isInLoginProcess = ref(false)  // 登录中状态展示


const formEl = ref<FormInstance>()

function submit(formEl: FormInstance | undefined) {
    // ElMessage({
    //     message: '账号密码不正确',
    //     type: 'warning',
    // })
    // return
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

.login-bg {
    display: flex;
    justify-content: center;
    align-items: center;
    flex-flow: column nowrap;
    //background: url('../assets/img/bg.jpg') no-repeat center center;
    background: url('../assets/img/bg-3.jpg') no-repeat center center;
    background-size: cover;
}

.logo {
    z-index: 10;
    display: flex;
    justify-content: center;
    flex-flow: row nowrap;
    margin-right: 20px;

    img {
        display: block;
        height: 60px;
    }
}

.login-title {
    display: flex;
    justify-content: center;
    align-items: center;
    text-align: center;
    font-size: 1.2rem;
    letter-spacing: 3px;
    margin-bottom: 40px;
    color: white;
    text-shadow: 2px 2px 1px color.adjust(black, $alpha: -0.8);
}

.login-panel {
    z-index: 10;
    padding: 30px 50px;
    @include border-radius(20px);
    width: 450px;
    background: linear-gradient(to bottom right, color.adjust(white, $alpha: -0.5), color.adjust(white, $alpha: -0.8));
    //background-color: color.adjust(white, $alpha: -0.2);
    @include backdrop-filter(saturate(150%) blur(15px));
    @include box-shadow(10px 10px 35px color.adjust(black, $alpha: -0.8))
}

.login-btn {
    display: block;
    width: 100%;
}

.download-link {
    font-size: 13px;
    text-align: center;

    .download-link-a {
        text-decoration: underline;
        text-decoration-color: white;
        color: white;
    }
}

.login-links {
    text-align: center;
    margin-top: 8px;
    a {
        color: rgba(255, 255, 255, 0.9);
        font-size: 13px;
        text-decoration: underline;
        margin: 0 10px;
    }
}

.forgot-tip {
    margin: 0 0 12px;
    color: #606266;
    font-size: 13px;
    line-height: 1.5;
}
</style>
