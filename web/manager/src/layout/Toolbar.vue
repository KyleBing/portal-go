<template>
    <div v-if="showToolbarActions" ref="rootEl" :class="['tool-bar', layout]">
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
import {computed, nextTick, onBeforeUnmount, onMounted, ref, useSlots, watch} from "vue";
import {useProjectStore} from "@/pinia";

const slots = useSlots()
const projectStore = useProjectStore()
const rootEl = ref<HTMLElement | null>(null)
let resizeObserver: ResizeObserver | null = null

const props = withDefaults(defineProps<{
    isShowSearchBar?: boolean
    layout?: 'horizontal' | 'vertical'
}>(), {
    isShowSearchBar: true,
    layout: 'vertical',
})

/** 关闭「搜索栏」模式时仍渲染 left/center/right 插槽（如仅操作按钮） */
const showToolbarActions = computed(() => {
    return props.isShowSearchBar || !!slots.left || !!slots.center || !!slots.right
})

// 按实际高度同步 contentInsets，避免固定 50px 把内容区算矮
function syncToolbarHeight() {
    if (!showToolbarActions.value || !rootEl.value) {
        projectStore.contentInsets.heightToolbar = 0
        return
    }
    projectStore.contentInsets.heightToolbar = rootEl.value.offsetHeight
}

watch(showToolbarActions, async (visible) => {
    if (!visible) {
        projectStore.contentInsets.heightToolbar = 0
        return
    }
    await nextTick()
    syncToolbarHeight()
}, {immediate: true})

onMounted(() => {
    resizeObserver = new ResizeObserver(() => syncToolbarHeight())
    if (rootEl.value) {
        resizeObserver.observe(rootEl.value)
    }
    syncToolbarHeight()
})

onBeforeUnmount(() => {
    resizeObserver?.disconnect()
    resizeObserver = null
    projectStore.contentInsets.heightToolbar = 0
})

</script>

<style scoped lang="scss">
@use '../assets/scss/plugin' as *;

.tool-bar {
    display: flex;
    align-items: center;
    min-height: 50px;
    padding: 5px 15px;
    color: $theme-text-main;
    background-color: $theme-bg-panel;
    box-sizing: border-box;
    border-bottom: 1px solid $theme-border-main;

    &.horizontal {
        flex-flow: row nowrap;
        justify-content: space-between;
    }
}

.search-container {
    flex: 1;
    width: 100%;

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
        gap: 8px;
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

    :deep(.el-form-item) {
        margin-bottom: 0;
    }
}
</style>
