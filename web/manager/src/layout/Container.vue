<template>
    <div class="container" :style="`height: ${projectStore.contentInsets.heightContent}px; padding: ${props.padding}px`">
        <slot></slot>
    </div>

</template>

<script setup lang="ts">
import {useProjectStore} from "../pinia";
const projectStore = useProjectStore()

const props = withDefaults(defineProps<{
    padding?: number
}>(),{
    padding: 0
})
</script>

<style lang="scss" scoped>
.container {
    background-color: white;
    padding: 0;
    position: relative;
    overflow: hidden;
    overflow-y: auto;
}

</style>
