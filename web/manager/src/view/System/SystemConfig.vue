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
                        <ElRadioGroup v-if="payQrList.length" v-model="form.registration_pay_amount_yuan">
                            <ElRadio
                                v-for="item in payQrList"
                                :key="item.id"
                                :value="item.amount_yuan"
                            >
                                {{ formatRegistrationPayAmount(item.amount_yuan) }}
                            </ElRadio>
                        </ElRadioGroup>
                        <p v-else class="hint block-hint">请先在「申请收款码」中添加金额档。</p>
                        <p class="hint block-hint">
                            用户打开注册申请页时看到当前档的金额、收款码和备注。收款码图片请到
                            <RouterLink to="/user/registration-pay-qr">申请收款码</RouterLink>
                            管理。
                        </p>
                    </ElFormItem>
                </ElForm>
            </div>
        </Content>
    </Container>
</template>

<script setup lang="ts">
import {ref, onMounted} from 'vue'
import {ElMessage} from 'element-plus'
import Container from '@/layout/Container.vue'
import Toolbar from '@/layout/Toolbar.vue'
import Content from '@/layout/Content.vue'
import systemApi from '@/api/systemApi'
import type { RegistrationPayQrItem } from '@/api/registrationPayQrApi'
import { formatRegistrationPayAmount, parsePayQrList } from '@/utils/payQrAssets'

const isLoading = ref(false)
const saving = ref(false)
const payQrList = ref<RegistrationPayQrItem[]>([])
const form = ref({
    invitation_code: '',
    registration_pay_amount_yuan: 28,
})

function load() {
    isLoading.value = true
    systemApi.getAdmin()
        .then((res: any) => {
            const d = res.data || {}
            payQrList.value = parsePayQrList(d.registration_pay_qr_list)
            form.value = {
                invitation_code: String(d.invitation_code || ''),
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
</style>
