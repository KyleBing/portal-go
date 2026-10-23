<template>
    <el-config-provider :locale="zhCn">
        <router-view/>
    </el-config-provider>
</template>

<script setup lang="ts">
import {useProjectStore, useThemeStore} from "@/pinia";
import {onBeforeUnmount, onMounted} from "vue";
import zhCn from "element-plus/es/locale/lang/zh-cn";

const projectStore = useProjectStore()
const themeStore = useThemeStore()

themeStore.loadTheme()
projectStore.refreshInsets()

function applyLayout() {
    projectStore.refreshInsets()
}

onMounted(() => {
    applyLayout()
    window.addEventListener('resize', applyLayout)
})

onBeforeUnmount(() => {
    window.removeEventListener('resize', applyLayout)
})
</script>

<style lang="scss">
@use "assets/scss/main";

body {
    background-color: var(--theme-bg-main, #eee);
    overflow: hidden;
    background-size: cover;
    height: 100%;
    background-attachment: fixed;
    background-repeat: no-repeat;
    background-position: top;
}
</style>
