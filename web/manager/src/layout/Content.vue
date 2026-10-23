<template>
    <div :class="['content', {'is-transparent': isTransparent}]"
         :style="`
             height: ${heightContent}px;
             padding: ${padding}
          `">
        <slot></slot>
    </div>

</template>

<script setup lang="ts">
import {useProjectStore} from "../pinia";
import { computed } from "vue";
const projectStore = useProjectStore()

const heightContent = computed(() => {
    return projectStore.contentInsets.heightContent
        - projectStore.contentInsets.heightToolbar
        + (props.heightPlus || 0)
        - (props.isShowPagination ? projectStore.contentInsets.heightPagination : 0)
})

const props = withDefaults(defineProps<{
    heightPlus?: number
    padding?: string
    isTransparent?: boolean
    isShowPagination?: boolean
}>(), {
    heightPlus: 0,
    padding: '20px 30px',
    isTransparent: false,
    isShowPagination: false
})
</script>

<style lang="scss" scoped>
@use '../assets/scss/plugin' as *;

.content {
    background-color: $theme-bg-main;
    position: relative;
    overflow: hidden;
    overflow-y: auto;

    &.is-transparent {
        background-color: transparent;
    }
}
</style>
