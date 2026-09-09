<template>
    <el-config-provider :locale="zhCn">
        <router-view/>
    </el-config-provider>
</template>

<script setup lang="ts">
import {useProjectStore} from "@/pinia";
import {onMounted} from "vue";
import { resizeHolePage} from "@/utility";
import {useRoute} from "vue-router";
import zhCn from "element-plus/es/locale/lang/zh-cn";

const route = useRoute()

const projectStore = useProjectStore()

onMounted(function() {
    refreshInsets()

    window.addEventListener('resize', function() {
        refreshInsets()

        let element = document.getElementById('app')!
        projectStore.pageScale = innerWidth / 1920
        resizeHolePage(route.path, element)
    })
})

function refreshInsets() {
    projectStore.pageScale = innerWidth / 1920

    projectStore.contentInsets.heightContent = window.innerHeight / projectStore.pageScale - projectStore.contentInsets.heightHeader
    projectStore.contentInsets.widthContent = window.innerWidth - projectStore.navWidth

    projectStore.windowInsets = {
        height: window.innerHeight,
        width: window.innerWidth,
    }
}

</script>

<style lang="scss">
@use "assets/scss/main";
@use "assets/scss/variables";

body{
    background-color: variables.$bg-light;
    overflow: hidden;
    background-size: cover;
    height: 100%;
    background-attachment: fixed;
    background-repeat: no-repeat;
    background-position: top;
}
</style>
