<template>
    <container class="bg">
        <div class="content">
            <div class="card">
                <div class="company">
                    <div class="company-logo">
                        <img src="../assets/logo.png" alt="logo"/>
                        <h2>{{ packageInfo.projectName }}</h2>
                    </div>

                    <p class="description">
                        感谢您使用，请将疑问和建议发送邮件至
                        <a :href="`mailto:${packageInfo.author.email}`">{{ packageInfo.author.email }}</a>
                    </p>
                </div>
            </div>
        </div>
    </container>
</template>

<script setup lang="ts">
import Container from "@/layout/Container.vue"
import packageInfo from "../../package.json"
</script>

<style lang="scss" scoped>
@use "../assets/scss/variables" as *;
@use "../assets/scss/utility" as *;
@use "../assets/scss/font" as *;
@use "sass:color";

$text-about: #666;
.bg{
    background-color: $color-border;
}
.product-logo {
    font-weight: bold;
    display: flex;
    justify-content: center;
    font-size: 2rem;
    align-items: center;
    img {
        display: block;
        height: 50px;
        margin-right: 20px;
    }
}
.about-title {
    color: $text-about;
    padding-left: 15px;
    font-weight: bold;
    border-left: 3px solid $color-main;
    font-size: 1rem;
    line-height: 2;
}
.content {
    width: 500px;
    padding: 20px;
}
.card {
    color: $text-about;
    background-color: white;
    box-shadow: 1px 1px 3px color.adjust(black, $alpha: -0.82);
    padding: 20px;
    -webkit-border-radius: 6px;
    -moz-border-radius: 6px;
    border-radius: 6px;
}
.company {
    font-size: 0.8rem;
    padding: 0 10px;
}
.company-logo {
    display: flex;
    justify-content: center;
    align-items: center;
    img {
        display: block;
        height: 50px;
    }
    h2 {
        margin-left: 20px;
        font-size: 1.5rem;
        letter-spacing: 1px;
    }
}
.info {
    margin-top: 20px;
    &-item {
        padding: 3px 0;
        display: flex;
        justify-content: flex-start;
    }
}
.description {
    margin-top: 20px;
}
a {
    color: $color-main;
    &:hover {
        text-decoration: underline
    }
}
</style>
