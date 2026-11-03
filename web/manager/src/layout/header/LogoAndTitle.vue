<template>
    <div class="logo">
        <img :src="logo" alt="LOGO">
    </div>
    <h2 class="title" v-if="projectStore.projectName">{{ projectStore.projectName }}</h2>
</template>

<script setup lang="ts">
import {useProjectStore} from "@/pinia";
import logo from "@/assets/logo.png"
const projectStore = useProjectStore()

</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;

.logo {
    img {
        display: block;
        height: 32px;
    }
}
.title{
    letter-spacing: 1px;
    white-space: nowrap;
    color: $text-main;
    margin-left: 10px;
    font-size: 22px;
}

</style>
