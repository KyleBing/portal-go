<template>
    <div class="register-bg" :style="`height:${viewportHeight}px`">
        <div class="register-panel">
            <div class="register-title">
                <h2>注册申请</h2>
                <p>
                    付费用户请填写下方信息提交注册申请。我们会在 1 日内完成审核，请耐心等待；
                    审核结果将以邮件形式通知您，通过后会附邀请码。
                </p>
                <p class="pay-fee-note">
                    费用用于服务器租用，希望日记可以长期维护下去，感谢您的支持。
                </p>
            </div>
            <ElForm
                v-if="!submitted"
                size="large"
                :model="form"
                :rules="rules"
                ref="formRef"
                label-width="0"
                @submit.prevent="submit"
            >
                <ElFormItem prop="email">
                    <ElInput v-model="form.email" clearable autocomplete="email" placeholder="联系邮箱" />
                </ElFormItem>
                <ElFormItem prop="paidAccount">
                    <ElInput v-model="form.paidAccount" clearable autocomplete="off" placeholder="付费昵称（支付宝/微信昵称）" />
                </ElFormItem>
                <ElFormItem prop="payPlatform">
                    <PayPlatformRadio v-model="form.payPlatform" />
                </ElFormItem>
                <div v-if="form.payPlatform" class="pay-qr-block">
                    <p v-if="registerFeeYuan > 0" class="pay-qr-amount">
                        <span class="pay-qr-amount__label">应付金额</span>
                        <span class="pay-qr-amount__value"><span class="pay-qr-amount__currency">¥</span>{{ registerFeeYuan }}</span>
                    </p>
                    <p v-else class="pay-qr-amount pay-qr-amount--custom">
                        <span class="pay-qr-amount__label">应付金额</span>
                        <span class="pay-qr-amount__value pay-qr-amount__value--custom">按约定金额支付</span>
                    </p>
                    <p class="pay-qr-hint">
                        请使用{{ payPlatformLabel }}扫码支付{{ registerFeeYuan > 0 ? '上方金额' : '约定金额' }}，并在表单中填写付款账户信息。
                    </p>
                    <img class="pay-qr-image" :src="payQrImageSrc" :alt="`${payPlatformLabel}收款码`">
                    <p class="pay-qr-payee">收款方：邴新科</p>
                </div>
                <ElFormItem prop="paidConfirmed" class="paid-confirm" :inline-message="true">
                    <ElCheckbox v-model="form.paidConfirmed">已支付</ElCheckbox>
                </ElFormItem>
                <ElButton class="register-btn" type="primary" native-type="submit" icon="Promotion" :loading="submitting" @click="submit">
                    提交申请
                </ElButton>
            </ElForm>
            <div v-else class="success-box">
                <p>申请已提交。我们会在 1 日内完成审核，请留意邮箱通知。</p>
                <RouterLink to="/login">登录</RouterLink>
            </div>
            <div class="register-links">
                <RouterLink to="/register">注册</RouterLink>
                <RouterLink to="/login">登录</RouterLink>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useProjectStore } from '@/pinia'
import type { FormInstance, FormItemRule, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import registrationApplicationApi, { type PayPlatform } from '@/api/registrationApplicationApi'
import PayPlatformRadio from '@/components/PayPlatformRadio.vue'
import systemApi from '@/api/systemApi'
import { isRegistrationPayAmount, payQrImage } from '@/utils/payQrAssets'

const projectStore = useProjectStore()
const viewportHeight = computed(() => projectStore.windowInsets.height)

const formRef = ref<FormInstance>()
const submitting = ref(false)
const submitted = ref(false)

const form = ref({
    email: '',
    paidAccount: '',
    payPlatform: 'alipay' as PayPlatform,
    paidConfirmed: false, // 用户确认已完成扫码支付，提交前必须勾选
})

const registerFeeYuan = ref(28)

const payPlatformLabel = computed(() => (form.value.payPlatform === 'wechat' ? '微信' : '支付宝'))

const payQrImageSrc = computed(() =>
    payQrImage(form.value.payPlatform as PayPlatform, registerFeeYuan.value),
)

function loadPayAmount() {
    systemApi
        .getPublic()
        .then((res: any) => {
            const n = Number(res.data?.registration_pay_amount_yuan)
            if (isRegistrationPayAmount(n)) {
                registerFeeYuan.value = n
            }
        })
        .catch(() => {})
}

onMounted(loadPayAmount)

// 未勾选「已支付」时不允许提交
const validatePaidConfirmed: FormItemRule['validator'] = (_rule, value, callback) => {
    if (value !== true) {
        callback(new Error('请勾选已支付后再提交'))
        return
    }
    callback()
}

const rules: FormRules = {
    email: [
        { required: true, message: '请填写邮箱', trigger: 'blur' },
        { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
    ],
    paidAccount: [{ required: true, message: '请填写付费昵称', trigger: 'blur' }],
    payPlatform: [{ required: true, message: '请选择付费平台', trigger: 'change' }],
    paidConfirmed: [{ validator: validatePaidConfirmed, trigger: 'change' }],
}

function submit() {
    const el = formRef.value
    if (!el) return
    el.validate((valid) => {
        if (!valid) return
        submitting.value = true
        registrationApplicationApi
            .submit({
                email: form.value.email.trim(),
                paidAccount: form.value.paidAccount.trim(),
                payPlatform: form.value.payPlatform as PayPlatform,
            })
            .then((res: any) => {
                submitted.value = true
                ElMessage.success(res.message || '提交成功')
            })
            .finally(() => {
                submitting.value = false
            })
    })
}
</script>

<style scoped lang="scss">
.register-bg {
    display: flex;
    justify-content: center;
    align-items: flex-start;
    min-height: 100%;
    background: #f4f5f7;
    overflow: auto;
}

.register-panel {
    z-index: 10;
    width: 440px;
    margin: 40px 0;
    padding: 40px 40px 28px;
    border: 1px solid #e6e8eb;
    border-radius: 16px;
    background: #fff;
    box-shadow: 0 12px 40px rgba(20, 24, 28, 0.08);

    :deep(.el-input__wrapper) {
        background-color: #fff !important;
    }
}

.register-title {
    text-align: left;
    margin-bottom: 28px;

    h2 {
        margin: 0;
        font-size: 24px;
        font-weight: 600;
        color: #1a1d21;
    }

    p {
        margin: 8px 0 0;
        font-size: 13px;
        line-height: 1.65;
        color: #6b7280;
    }

    .pay-fee-note {
        margin-top: 12px;
        padding: 10px 12px;
        border-radius: 8px;
        background: #f0f9ff;
        border: 1px solid #bae6fd;
        color: #0369a1;
    }
}

.register-btn {
    display: block;
    width: 100%;
    margin-top: 6px;
}

.register-links {
    margin-top: 18px;
    text-align: center;
    display: flex;
    justify-content: center;
    gap: 16px;

    a {
        color: #4b5563;
        font-size: 13px;
        text-decoration: none;
    }

    a:hover {
        color: #1a1d21;
    }
}

.pay-qr-block {
    margin: 4px 0 8px;
    padding: 16px;
    text-align: center;
    border: 1px solid #e5e7eb;
    border-radius: 12px;
    background: #fafafa;
}

.pay-qr-amount {
    margin: 0 0 12px;
    padding: 12px 16px;
    border-radius: 10px;
    background: #fff;
    border: 1px solid #e5e7eb;
}

.pay-qr-amount__label {
    display: block;
    margin-bottom: 4px;
    font-size: 13px;
    font-weight: 500;
    color: #6b7280;
    letter-spacing: 0.02em;
}

.pay-qr-amount__value {
    display: block;
    font-size: 36px;
    font-weight: 700;
    line-height: 1.1;
    color: #c2410c;
    letter-spacing: -0.02em;
}

.pay-qr-amount__currency {
    margin-right: 2px;
    font-size: 22px;
    font-weight: 600;
    vertical-align: 0.08em;
}

.pay-qr-amount__value--custom {
    font-size: 22px;
    color: #374151;
}

.pay-qr-hint {
    margin: 0 0 12px;
    font-size: 13px;
    line-height: 1.55;
    color: #6b7280;
}

.pay-qr-image {
    display: block;
    width: min(220px, 100%);
    margin: 0 auto;
    border-radius: 8px;
}

.pay-qr-payee {
    margin: 10px 0 0;
    font-size: 14px;
    font-weight: 500;
    color: #374151;
}

.paid-confirm {
    margin-top: 4px;
    margin-bottom: 8px;

    :deep(.el-checkbox__label) {
        font-size: 14px;
        color: #1a1d21;
    }

    // 校验提示跟在勾选框右侧，避免压住提交按钮
    :deep(.el-form-item__error) {
        padding-top: 0;
        line-height: 1.4;
    }
}

.success-box {
    font-size: 14px;
    line-height: 1.65;
    color: #374151;

    a {
        display: inline-block;
        margin-top: 12px;
        color: #2563eb;
        text-decoration: none;
    }
}
</style>
