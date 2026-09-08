<template>
    <ElAside :width="`${projectStore.navWidth}px`" :style="`height:${heightAside}px`">
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
    heightAside.value =  projectStore.windowInsets.height / projectStore.pageScale - 60
    heightNavbar.value =
        projectStore.navMenuIsClosed ?
        heightAside.value :
        heightAside.value - heightCopyright.value
}

const {windowInsets}  = storeToRefs(projectStore)

watch(windowInsets, resizeComponents)
// watch(projectStore.navMenuIsClosed, resizeComponents)


</script>

<style lang="scss" scoped>
@use "../assets/scss/variables" as *;
@use "sass:color";

$border-color: #ddd;
.navbar{
    border-right: 1px solid $border-color;
    overflow: hidden;
    background-color: white;
}

.side-menu{
    overflow: hidden;
    overflow-y: auto;
    &::-webkit-scrollbar {
        z-index: 50;
        width: 3px;
        height: 8px; }

    &::-webkit-scrollbar-track {
        border-color: transparent;
        background-color: transparent; }

    &::-webkit-scrollbar-thumb {
        -webkit-border-radius: 5px;
        -moz-border-radius: 5px;
        border-radius: 5px;
        background-color: color.adjust(black, $alpha: -0.9);
        transition: all .2s;
        height: 3px; }

    &:hover::-webkit-scrollbar-thumb {
        background-color: color.adjust(black, $alpha: -0.8);
        transition: all .2s; }
}
</style>
