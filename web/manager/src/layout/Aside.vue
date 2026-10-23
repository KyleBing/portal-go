<template>
    <ElAside :width="`${projectStore.navWidth}px`" :style="`height:${heightAside}px`" class="aside">
       <div class="navbar">
           <Navbar
               class="side-menu"
               :style="`height: ${heightNavbar}px`"/>
           <Copyright v-show="!projectStore.navMenuIsClosed" :height="heightCopyright"/>
       </div>
    </ElAside>
</template>

<script lang="ts" setup>
import Navbar from "./Navbar.vue"
import Copyright from "./Copyright.vue";
import {useProjectStore} from "../pinia";
import {onMounted, ref, watch} from "vue";
import {storeToRefs} from "pinia";

const projectStore = useProjectStore()

const heightAside = ref(0)
const heightNavbar = ref(0)
const heightCopyright= ref(100)


onMounted(()=> {
    resizeComponents()
})

function resizeComponents(){
    heightAside.value =  projectStore.windowInsets.height - 60
    heightNavbar.value =
        projectStore.navMenuIsClosed ?
        heightAside.value :
        heightAside.value - heightCopyright.value
}

const {windowInsets, navMenuIsClosed}  = storeToRefs(projectStore)

watch([windowInsets, navMenuIsClosed], resizeComponents)


</script>

<style lang="scss" scoped>
@use 'sass:color';
@use "../assets/scss/plugin" as *;

.aside {
    background-color: $theme-bg-panel;
    border-right: 1px solid $theme-border-card;
}

.navbar{
    overflow: hidden;
    background-color: $theme-bg-main;
}

.side-menu{
    overflow: hidden;
    overflow-y: auto;
    &::-webkit-scrollbar {
        z-index: 50;
        width: 3px;
        height: 8px; }

    &::-webkit-scrollbar-track {
        background-color: transparent; }

    &::-webkit-scrollbar-thumb {
        border-radius: 5px;
        background-color: color.adjust(black, $alpha: -0.9);
    }
}
</style>
