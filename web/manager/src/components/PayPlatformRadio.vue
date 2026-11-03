<template>
    <ElRadioGroup :model-value="modelValue" class="pay-platform-group" @update:model-value="onChange">
        <ElRadio v-for="opt in options" :key="opt.value" :value="opt.value" border class="pay-platform-option">
            <span class="pay-platform-option__inner">
                <img class="pay-platform-option__icon" :src="opt.icon" :alt="opt.label">
                <span class="pay-platform-option__label">{{ opt.label }}</span>
            </span>
        </ElRadio>
    </ElRadioGroup>
</template>

<script setup lang="ts">
import type { PayPlatform } from '@/api/registrationApplicationApi'
import alipayIcon from '@/assets/alipay.png'
import wepayIcon from '@/assets/wepay.png'

defineProps<{
    modelValue: PayPlatform | ''
}>()

const emit = defineEmits<{
    'update:modelValue': [value: PayPlatform]
}>()

const options: { value: PayPlatform; label: string; icon: string }[] = [
    { value: 'alipay', label: '支付宝', icon: alipayIcon },
    { value: 'wechat', label: '微信支付', icon: wepayIcon },
]

function onChange(value: PayPlatform | boolean | number | string | undefined) {
    if (value === 'alipay' || value === 'wechat') {
        emit('update:modelValue', value)
    }
}
</script>

<style scoped lang="scss">
.pay-platform-group {
    display: flex;
    width: 100%;
    gap: 12px;

    :deep(.el-radio) {
        flex: 1;
        margin-right: 0;
        height: auto;
        padding: 0;
    }

    :deep(.el-radio__input) {
        display: none;
    }

    :deep(.el-radio__label) {
        padding: 0;
        width: 100%;
    }

    :deep(.el-radio.is-bordered) {
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 14px 12px;
        border-radius: 10px;
        border-color: #e5e7eb;
        background: #fafafa;
        transition: border-color 0.15s ease, background 0.15s ease, box-shadow 0.15s ease;

        &:hover {
            border-color: #c5cad3;
        }

        &.is-checked {
            border-color: var(--el-color-primary);
            background: var(--el-color-primary-light-9);
            box-shadow: 0 0 0 1px var(--el-color-primary-light-7);
        }
    }
}

.pay-platform-option__inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    width: 100%;
}

.pay-platform-option__icon {
    display: block;
    width: 36px;
    height: 36px;
    object-fit: contain;
}

.pay-platform-option__label {
    font-size: 14px;
    font-weight: 500;
    color: #1a1d21;
    line-height: 1.2;
}
</style>
