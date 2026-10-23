<template>
    <div class="container" :style="`${styleHeight} padding: ${props.padding}px`">
        <slot></slot>
    </div>

</template>

<script setup lang="ts">
import { computed } from 'vue'
import {useProjectStore} from "../pinia";
const projectStore = useProjectStore()

const props = withDefaults(defineProps<{
    isFixHeight?: boolean
    padding?: number
}>(),{
    isFixHeight: true,
    padding: 0
})

const styleHeight = computed(() => {
    if (props.isFixHeight) {
        return `height: ${projectStore.contentInsets.heightContent}px;`
    }
    return 'height: auto;'
})
</script>

<style lang="scss" scoped>
@use '../assets/scss/plugin' as *;

.container {
    background-color: $theme-bg-main;
    padding: 0;
    position: relative;
    overflow: hidden;
    overflow-y: auto;
}

</style>
