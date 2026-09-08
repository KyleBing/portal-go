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
    heightPlus?: number   // 高度变化，改变组件的高度
    padding?: string  // 20px
    isTransparent?: boolean  // 背景是否透明
    isShowPagination?: boolean  // 是否显示分页
}>(), {
    heightPlus: 0,
    padding: '20px 30px',
    isTransparent: false,
    isShowPagination: true
})
</script>

<style lang="scss" scoped>
.content {
    padding: 20px 30px;
    background-color: white;
    position: relative;
    overflow: hidden;
    overflow-y: auto;
    &.is-transparent{
        background-color: transparent;
    }
}
</style>
