<template>
    <div v-if="showToolbarActions" :class="['tool-bar', layout]" :style="styleHeight">
        <div class="search-container">
            <div class="left">
                <slot name="left"/>
            </div>
            <div class="center">
                <slot name="center"/>
            </div>
            <div class="right">
                <slot name="right"/>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import {computed, onBeforeUnmount, useSlots, watch} from 'vue'
import {useProjectStore} from "@/pinia";

const slots = useSlots()
const projectStore = useProjectStore()

const props = withDefaults(defineProps<{
    isShowSearchBar?: boolean
    layout?: 'horizontal' | 'vertical'
    isHeightFromContent?: boolean
}>(), {
    isShowSearchBar: true,
    layout: 'vertical',
    isHeightFromContent: false,
})

/** 关闭「搜索栏」模式时仍渲染 left/center/right 插槽（如仅操作按钮） */
const showToolbarActions = computed(() => {
    return props.isShowSearchBar || !!slots.left || !!slots.center || !!slots.right
})

const toolbarHeightPx = computed(() => {
    if (!showToolbarActions.value) return 0
    if (props.isHeightFromContent) return null
    return 50
})

const styleHeight = computed(() => {
    if (props.isHeightFromContent) return 'height: auto;'
    return `height: ${toolbarHeightPx.value}px;`
})

watch(toolbarHeightPx, (height) => {
    projectStore.contentInsets.heightToolbar = height ?? 0
}, {immediate: true})

onBeforeUnmount(() => {
    projectStore.contentInsets.heightToolbar = 0
})
</script>

<style scoped lang="scss">
@use '../assets/scss/plugin' as *;

.tool-bar {
    padding: 5px 15px;
    color: $theme-text-main;
    background-color: $theme-bg-panel;

    &.horizontal {
        display: flex;
        flex-flow: row nowrap;
        justify-content: space-between;
        align-items: center;
    }
}

.search-container {
    align-items: center;
    min-height: 40px;
    display: flex;
    flex-flow: row wrap;
    gap: 10px;
    justify-content: space-between;

    .left,
    .center,
    .right {
        display: flex;
        flex-flow: row nowrap;
        align-items: center;
    }

    .left {
        justify-content: flex-start;
        flex-shrink: 0;
    }

    .center {
        justify-content: flex-start;
        flex-grow: 1;
        flex-shrink: 0;
    }

    .right {
        justify-content: flex-end;
        flex-shrink: 0;
        flex-grow: 1;
    }
}
</style>
