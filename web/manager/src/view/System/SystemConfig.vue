<template>
    <Container>
        <Toolbar :is-show-search-bar="false">
            <template #left>
                <ElButton type="primary" icon="Check" :loading="saving" @click="save">保存</ElButton>
            </template>
        </Toolbar>
        <Content padding="10px" v-loading="isLoading">
            <div class="panel">
                <div class="panel-head">注册与邀请</div>
                <ElForm label-width="160px" size="small" :model="form">
                    <ElFormItem label="通用邀请码">
                        <ElInput v-model="form.invitation_code" placeholder="留空表示只能使用一次性邀请码" clearable/>
                    </ElFormItem>
                </ElForm>
            </div>

            <div class="panel">
                <div class="panel-head">注册申请收款</div>
                <ElForm label-width="160px" size="small" :model="form">
                    <ElFormItem label="当前应付金额">
                        <ElRadioGroup v-model="form.registration_pay_amount_yuan">
                            <ElRadio
                                v-for="opt in payAmountOptions"
                                :key="opt.value"
                                :value="opt.value"
                            >
                                {{ opt.label }}
                            </ElRadio>
                        </ElRadioGroup>
                        <p class="hint block-hint">
                            用户打开注册申请页时将看到对应金额的收款码；提交申请时会自动记录该金额到台账。默认 ¥28。
                        </p>
                    </ElFormItem>
                    <ElFormItem label="收款码预览">
                        <div class="qr-preview">
                            <div class="qr-preview__item">
                                <span class="qr-preview__label">支付宝</span>
                                <img :src="previewAlipay" alt="支付宝收款码预览">
                            </div>
                            <div class="qr-preview__item">
                                <span class="qr-preview__label">微信</span>
                                <img :src="previewWepay" alt="微信收款码预览">
                            </div>
                        </div>
                    </ElFormItem>
                </ElForm>
            </div>
        </Content>
    </Container>
</template>

<script setup lang="ts">
import {ref, computed, onMounted} from 'vue'
import {ElMessage} from 'element-plus'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'
import systemApi from '@/api/systemApi'
import { REGISTRATION_PAY_AMOUNT_OPTIONS, payQrImage } from '@/utils/payQrAssets'

const payAmountOptions = REGISTRATION_PAY_AMOUNT_OPTIONS

const isLoading = ref(false)
const saving = ref(false)
const form = ref({
    invitation_code: '',
    registration_pay_amount_yuan: 28,
})

const previewAlipay = computed(() => payQrImage('alipay', form.value.registration_pay_amount_yuan))
const previewWepay = computed(() => payQrImage('wechat', form.value.registration_pay_amount_yuan))

function load() {
    isLoading.value = true
    systemApi.getAdmin()
        .then((res: any) => {
            const d = res.data || {}
            form.value = {
                invitation_code: d.invitation_code || '',
                registration_pay_amount_yuan: Number.isFinite(Number(d.registration_pay_amount_yuan))
                    ? Number(d.registration_pay_amount_yuan)
                    : 28,
            }
        })
        .finally(() => {
            isLoading.value = false
        })
}

function save() {
    saving.value = true
    systemApi.save({
        invitation_code: form.value.invitation_code,
        registration_pay_amount_yuan: form.value.registration_pay_amount_yuan,
    })
        .then((res: any) => {
            ElMessage.success(res.message || '已保存')
            load()
        })
        .finally(() => {
            saving.value = false
        })
}

onMounted(load)
</script>

<style scoped lang="scss">
.panel {
    background: #fff;
    border-radius: 8px;
    padding: 14px 16px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, .05);
    max-width: 760px;
    margin-bottom: 12px;
}
.panel-head {
    font-size: 14px;
    font-weight: 600;
    color: #1f2d3d;
    margin-bottom: 12px;
}
.hint {
    margin-left: 10px;
    color: #909399;
    font-size: 12px;
}
.block-hint {
    display: block;
    margin: 8px 0 0;
    margin-left: 0 !important;
    line-height: 1.5;
}
.qr-preview {
    display: flex;
    flex-wrap: wrap;
    gap: 20px;
}
.qr-preview__item {
    text-align: center;
    img {
        display: block;
        width: 140px;
        border-radius: 8px;
        border: 1px solid #e4e7ed;
    }
}
.qr-preview__label {
    display: block;
    margin-bottom: 6px;
    font-size: 12px;
    color: #606266;
}
</style>
