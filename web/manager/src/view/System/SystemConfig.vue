<template>
    <Container>
        <Toolbar :is-show-search-bar="false">
            <template #left>
                <ElButton type="primary" icon="Check" :loading="saving" @click="save">保存</ElButton>
            </template>
        </Toolbar>
        <Content padding="10px" v-loading="isLoading">
            <div class="panel">
                <div class="panel-head panel-head--row">
                    <div>
                        <div class="panel-title">注册与邀请</div>
                        <p class="panel-desc">可复用的通用邀请码；留空则注册只能使用一次性邀请码</p>
                    </div>
                    <RouterLink class="manage-link" to="/user/invitation">管理邀请码</RouterLink>
                </div>

                <div class="invite-field">
                    <label class="invite-field__label" for="invitation-code">通用邀请码</label>
                    <ElInput
                        id="invitation-code"
                        v-model="form.invitation_code"
                        size="default"
                        placeholder="例如 WELCOME2026，留空表示关闭通用邀请"
                        clearable
                    />
                    <p class="invite-field__hint">
                        该码可被多人反复使用；一次性邀请码请到「邀请码」页生成与管理。
                    </p>
                </div>
            </div>

            <div class="panel panel--pay">
                <div class="panel-head panel-head--row">
                    <div>
                        <div class="panel-title">注册申请收款</div>
                        <p class="panel-desc">选择用户注册申请页展示的应付金额档位</p>
                    </div>
                    <RouterLink class="manage-link" to="/user/registration-pay-qr">管理收款码</RouterLink>
                </div>

                <div v-if="payQrList.length" class="amount-grid">
                    <button
                        v-for="item in payQrList"
                        :key="item.id"
                        type="button"
                        class="amount-card"
                        :class="{ 'is-active': form.registration_pay_amount_yuan === item.amount_yuan }"
                        @click="form.registration_pay_amount_yuan = item.amount_yuan"
                    >
                        <div class="amount-card__top">
                            <span class="amount-card__value">{{ formatAmountDisplay(item.amount_yuan) }}</span>
                            <span v-if="form.registration_pay_amount_yuan === item.amount_yuan" class="amount-card__badge">当前</span>
                        </div>
                        <div class="amount-card__platforms">
                            <span class="platform-dot" :class="{ 'is-ready': !!item.alipay_url }">支付宝</span>
                            <span class="platform-dot" :class="{ 'is-ready': !!item.wechat_url }">微信</span>
                        </div>
                        <p v-if="item.alipay_note || item.wechat_note" class="amount-card__note">
                            {{ item.alipay_note || item.wechat_note }}
                        </p>
                    </button>
                </div>
                <div v-else class="empty-pay">
                    <p>还没有金额档位，请先添加收款码后再选择当前应付金额。</p>
                    <RouterLink to="/user/registration-pay-qr">
                        <ElButton type="primary" plain size="small">去添加收款码</ElButton>
                    </RouterLink>
                </div>
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

function formatAmountDisplay(amountYuan: number) {
    if (amountYuan === 0) return '自定义'
    return formatRegistrationPayAmount(amountYuan)
}

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
    border-radius: 10px;
    padding: 16px 18px;
    box-shadow: 0 2px 10px rgba(0, 0, 0, .05);
    max-width: 860px;
    margin-bottom: 12px;
}
.panel-head {
    font-size: 14px;
    font-weight: 600;
    color: #1f2d3d;
    margin-bottom: 12px;
}
.panel-head--row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 16px;
}
.panel-title {
    font-size: 15px;
    font-weight: 650;
    color: #1f2d3d;
    line-height: 1.3;
}
.panel-desc {
    margin: 4px 0 0;
    font-size: 12px;
    color: #909399;
    font-weight: 400;
    line-height: 1.4;
}
.manage-link {
    flex-shrink: 0;
    font-size: 12px;
    color: var(--el-color-primary);
    text-decoration: none;
    padding-top: 2px;

    &:hover {
        text-decoration: underline;
    }
}

.invite-field {
    max-width: 420px;
}

.invite-field__label {
    display: block;
    margin-bottom: 8px;
    font-size: 13px;
    font-weight: 500;
    color: #606266;
    line-height: 1.2;
}

.invite-field__hint {
    margin: 8px 0 0;
    font-size: 12px;
    color: #909399;
    line-height: 1.5;
}

.amount-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 10px;
}

.amount-card {
    appearance: none;
    border: 1px solid #e5e7eb;
    background: #fafbfc;
    border-radius: 10px;
    padding: 14px 14px 12px;
    text-align: left;
    cursor: pointer;
    transition: border-color .15s ease, background .15s ease, box-shadow .15s ease;

    &:hover {
        border-color: #c5cad3;
        background: #fff;
    }

    &.is-active {
        border-color: var(--el-color-primary);
        background: var(--el-color-primary-light-9);
        box-shadow: 0 0 0 1px var(--el-color-primary-light-7);
    }
}

.amount-card__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 10px;
}

.amount-card__value {
    font-size: 22px;
    font-weight: 700;
    color: #1f2d3d;
    letter-spacing: -0.02em;
    line-height: 1.1;
}

.amount-card__badge {
    flex-shrink: 0;
    font-size: 11px;
    font-weight: 600;
    color: var(--el-color-primary);
    background: #fff;
    border: 1px solid var(--el-color-primary-light-5);
    border-radius: 999px;
    padding: 1px 7px;
    line-height: 1.4;
}

.amount-card__platforms {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}

.platform-dot {
    font-size: 11px;
    color: #a0a6b0;
    background: #eef0f3;
    border-radius: 4px;
    padding: 2px 6px;
    line-height: 1.4;

    &.is-ready {
        color: #2f7a4c;
        background: #e8f6ee;
    }
}

.amount-card__note {
    margin: 8px 0 0;
    font-size: 11px;
    color: #909399;
    line-height: 1.4;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

.empty-pay {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 10px;
    padding: 18px 16px;
    border-radius: 10px;
    background: #f7f8fa;
    border: 1px dashed #dcdfe6;

    p {
        margin: 0;
        font-size: 13px;
        color: #606266;
        line-height: 1.5;
    }
}
</style>
