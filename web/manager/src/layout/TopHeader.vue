<template>
    <div class="header">
        <div class="logo">
            <img src="../assets/logo-white.png" alt="LOGO">
        </div>
        <h2 class="header">{{ projectName }}</h2>
    </div>
</template>

<script setup lang="ts">
import packageInfo from '../../package.json'

const projectName = packageInfo.projectName
</script>

<style scoped lang="scss">
@use "../assets/scss/variables" as *;
@use "../assets/scss/utility" as *;
@use "../assets/scss/font" as *;
@use "sass:color";

.header {
    padding: 0 20px;
    background: linear-gradient(to top, $color-main, color.scale($color-main, $lightness: -30%));
    display: flex;
    font-size: 1.3rem;
    justify-content: flex-start;
    flex-flow: row nowrap;
    line-height: 60px;
    color: color.adjust(white, $alpha: -0.1);
    .logo {
        padding: 15px;
        img {
            display: block;
            height: 30px;
        }
    }
}
</style>
