<template>
    <div class="logo">
        <img :src="logo" alt="LOGO">
    </div>
    <h2 class="title" v-if="projectStore.projectName">{{ projectStore.projectName }}</h2>
</template>

<script setup lang="ts">
import {useProjectStore} from "@/pinia";
const projectStore = useProjectStore()
import logo from "@/assets/logo-white.png"

</script>

<style scoped lang="scss">
@use "../../assets/scss/variables" as *;
@use "../../assets/scss/utility" as *;
@use "../../assets/scss/font" as *;

.logo {
    margin-left: 10px;
    padding: 10px;
    img {
        display: block;
        height: 32px;
    }
}
.title{
    letter-spacing: 1px;
    white-space: nowrap;
    color: white;
    margin-left: 10px;
    font-size: 22px;
}

</style>
